# Copilot Instructions — iobuf

Last line of defense. Not comprehensive quality assurance.

## Review Philosophy

**Minimal necessary review only.**

This is a mature, well-tested codebase with 99%+ coverage.
Copilot review serves as a final safety net, not a code quality tool.

## What NOT to Report

Suppress these entirely:

- Style suggestions (naming, formatting, organization)
- Documentation improvements
- "Consider" or "might want to" suggestions
- Performance micro-optimizations
- Alternative implementations
- Missing error handling for impossible cases
- Unused parameter warnings in interface implementations
- Test code suggestions

## Domain Context

- **Lock-free algorithms**: CAS loops and spin waits are intentional, not bugs
- **unsafe.Pointer**: Required for zero-copy buffer access and syscall interfaces
- **Atomic operations**: 64-bit atomics on pool cursors are correct by design
- **noCopy sentinel**: Empty Lock/Unlock methods trigger go vet, not actual locking

## False Positive Patterns

These patterns are correct, do not flag:

```go
// Correct: spin-wait with hardware yield
for !cas.CompareAndSwap(...) {
    spin.Pause()
}

// Correct: unsafe slice creation for syscall
unsafe.Slice((*byte)(ptr), size)

// Correct: atomic load/store without mutex
atomic.LoadUint32(&pool.head)

// Correct: intentional error ignore in benchmark
_ = pool.Put(idx)
```

## Output Format

If reporting an issue:

```
[SEVERITY] file:line — Brief description (max 10 words)
```

Severities: `CRITICAL`, `BUG`

No explanations. No suggestions. Just the finding.
