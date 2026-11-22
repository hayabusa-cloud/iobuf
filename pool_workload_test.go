// ©Hayabusa Cloud Co., Ltd. 2026. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package iobuf_test

import (
	"bytes"
	"slices"
	"testing"
	"time"

	"code.hybscloud.com/iobuf"
	"code.hybscloud.com/iox"
)

// BenchmarkPoolReadBurst compares two ways for a connection to borrow an 8 KiB
// read buffer. Each burst admits new readers before retiring old ones.
func BenchmarkPoolReadBurst(b *testing.B) {
	const capacity = 64
	b.Run("ExternalSlices", func(b *testing.B) {
		pool := iobuf.NewBoundedPool[int](capacity)
		pool.Fill(func() int { return 0 })
		buffers := iobuf.NewBuffers(capacity, iobuf.BufferSizeMedium)
		benchmarkReadBurst(b, pool, func(index int) []byte { return buffers[index] })
	})
	b.Run("PooledSlices", func(b *testing.B) {
		pool := iobuf.NewBoundedPool[[]byte](capacity)
		pool.Fill(func() []byte { return make([]byte, iobuf.BufferSizeMedium) })
		benchmarkReadBurst(b, pool, pool.Value)
	})
}

func benchmarkReadBurst[T any](b *testing.B, pool *iobuf.BoundedPool[T], buffer func(int) []byte) {
	const (
		established = 48
		arrivals    = 24
		retirements = 16
	)
	type reader struct {
		index int
		buf   []byte
	}
	pool.SetNonblock(true)
	var held [established]reader
	for i := range held {
		index, err := pool.Get()
		if err != nil {
			b.Fatal(err)
		}
		held[i] = reader{index, buffer(index)}
	}

	// Four established connections each receive two distinct 128-byte records
	// on the same buffer. Other connections keep their slots while idle.
	var records [2][4][128]byte
	for batch := range records {
		for peer := range records[batch] {
			for i := range records[batch][peer] {
				records[batch][peer][i] = byte(batch*4 + peer + i)
			}
		}
	}
	b.ReportAllocs()
	b.SetBytes(int64(len(records) * len(records[0]) * len(records[0][0])))
	var misses, returned int64
	oldest := 0
	b.ResetTimer()
	cpuStart := benchmarkCPUTime(b)
	for range b.N {
		var incoming [arrivals]reader
		admitted := 0
		for range arrivals {
			index, err := pool.Get()
			if err == iox.ErrWouldBlock {
				misses++
				continue
			}
			if err != nil {
				b.Fatal(err)
			}
			incoming[admitted] = reader{index, buffer(index)}
			admitted++
		}
		if admitted != retirements {
			b.Fatalf("admitted %d of %d arrivals, want %d", admitted, arrivals, retirements)
		}

		retiring := held[oldest : oldest+retirements]
		for batch := range records {
			for peer := range records[batch] {
				copy(retiring[peer].buf, records[batch][peer][:])
			}
			for peer := range records[batch] {
				want := records[batch][peer][:]
				if !bytes.Equal(retiring[peer].buf[:len(want)], want) {
					b.Fatalf("incorrect record for slot %d", retiring[peer].index)
				}
			}
		}
		for i := range retiring {
			if err := pool.Put(retiring[i].index); err != nil {
				b.Fatal(err)
			}
			retiring[i] = incoming[i]
		}
		returned += int64(len(retiring))
		oldest = (oldest + retirements) % established
	}
	cpu := benchmarkCPUTime(b) - cpuStart
	b.StopTimer()
	for _, r := range held {
		if err := pool.Put(r.index); err != nil {
			b.Fatal(err)
		}
	}
	b.ReportMetric(float64(misses)/float64(b.N), "misses/op")
	b.ReportMetric(float64(returned)/float64(b.N), "returns/op")
	b.ReportMetric(float64(returned)/b.Elapsed().Seconds(), "returns/s")
	if cpu > 0 {
		b.ReportMetric(float64(cpu)/float64(returned), "cpu-ns/return")
	}
}

// BenchmarkPoolWaitForReturn measures the CPU used while the only slot is
// held and the delay from starting its return to the waiter's successful Get.
func BenchmarkPoolWaitForReturn(b *testing.B) {
	b.Run("BoundedPool", func(b *testing.B) {
		pool := iobuf.NewBoundedPool[int](1)
		pool.Fill(func() int { return 0 })
		benchmarkWaitForReturn(b, pool.Get, pool.Put)
	})
	b.Run("Channel", func(b *testing.B) {
		slots := make(chan int, 1)
		slots <- 0
		get := func() (int, error) { return <-slots, nil }
		put := func(slot int) error {
			slots <- slot
			return nil
		}
		benchmarkWaitForReturn(b, get, put)
	})
}

func benchmarkWaitForReturn(b *testing.B, get func() (int, error), put func(int) error) {
	slot, err := get()
	if err != nil {
		b.Fatal(err)
	}
	handoff := make(chan int)
	returnedAt := make(chan time.Time, 1)
	go func() {
		defer close(returnedAt)
		for owned := range handoff {
			time.Sleep(time.Millisecond)
			start := time.Now()
			if err := put(owned); err != nil {
				panic(err) // The getter would stay blocked after a failed return.
			}
			returnedAt <- start
		}
	}()
	defer func() {
		close(handoff)
		for range returnedAt {
		}
	}()

	sample := func() time.Duration {
		handoff <- slot
		next, err := get()
		acquiredAt := time.Now()
		if err != nil {
			b.Fatal(err)
		}
		slot = next
		return acquiredAt.Sub(<-returnedAt)
	}
	sample() // Warm the goroutines and timer before collecting samples.
	delays := make([]time.Duration, b.N)
	b.ReportAllocs()
	b.ResetTimer()
	cpuStart := benchmarkCPUTime(b)
	for i := range delays {
		delays[i] = sample()
	}
	cpu := benchmarkCPUTime(b) - cpuStart
	b.StopTimer()
	if err := put(slot); err != nil {
		b.Fatal(err)
	}
	if cpu > 0 {
		b.ReportMetric(float64(cpu)/float64(b.N), "cpu-ns/return")
		b.ReportMetric(1000*float64(cpu)/float64(b.Elapsed()), "cpu-ms/s")
	}
	slices.Sort(delays)
	n := len(delays)
	b.ReportMetric(float64(delays[(n-1)/2]), "p50-handoff-ns")
	if n >= 10000 {
		b.ReportMetric(float64(delays[n-1-n/1000]), "p99.9-handoff-ns")
	}
}
