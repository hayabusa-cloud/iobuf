# Pool benchmarks

`BenchmarkPoolReadBurst` replays the lifetime of ordinary connection read buffers on one owner goroutine. Both layouts provision 64 distinct 8 KiB buffers before timing. `ExternalSlices` keeps the bytes in a separate slice indexed by `BoundedPool[int]`, as in `urex.ReadPool`. `PooledSlices` obtains the backing from `BoundedPool[[]byte].Value`.

Each burst starts with 48 occupied slots. It tries 24 new checkouts while the old reads still hold their buffers. Four established readers each receive two 128-byte records in the same buffer. The replay verifies the payloads, returns 16 old slots, and keeps the 16 new acquisitions for later bursts. Eight attempts return `iox.ErrWouldBlock`. A live unframed `urex` runtime would allocate an owned buffer for each of those readers.

| Workload | What to read |
| --- | --- |
| Sequential nonblocking Get/Put | The pool and a buffered channel each start with 1,024 indices. `ns/op` covers a successful checkout and return with one owner and no payload access. |
| Read burst | `ns/op`, `B/op` and `allocs/op` cover one complete burst. `MB/s` counts the 1 KiB of verified payload. `returns/s` and `cpu-ns/return` relate the cost of the entire burst to its 16 successful slot returns. |
| Wait for return | A separate goroutine holds the only slot for 1 ms while one caller waits in blocking `Get` using `iox.Backoff`. A buffered Go channel runs the same handoff. `cpu-ms/s` measures process CPU per wall second over the whole cycle. Handoff latency runs from the start of the holder's return to the waiter's acquisition and includes scheduler delay. |

On Unix, the benchmarks report process user plus system CPU time when the clock records a positive interval. Run one benchmark process at a time. Use the same Go version, hardware, CPU settings and `-cpu` for both revisions. Run on native amd64 and LSE2 arm64 to compare those targets.

```sh
go test -run '^$' -bench '^BenchmarkPool_Sequential_GetPut_Nonblocking' -benchmem -benchtime=200ms -count=10 -cpu=1 -timeout=120s
go test -run '^$' -bench '^BenchmarkPoolReadBurst$' -benchmem -benchtime=200ms -count=10 -cpu=1 -timeout=120s
go test -run '^$' -bench '^BenchmarkPoolWaitForReturn$' -benchmem -benchtime=1s -count=10 -cpu=1 -timeout=120s
```

The handoff benchmark warms up one transfer, then samples each return. It reports the median and, with at least 10,000 measured handoffs per case, p99.9. Use `-benchtime=10000x` for the latter. The Go iteration count is the sample count. Collect repeated runs into separate files and compare them with `benchstat before.txt after.txt`.

Run `make bench` for the full suite. Its shared-pool contention benchmarks count nonblocking checkout attempts, including misses. The read-burst benchmark reports completed returns separately.
