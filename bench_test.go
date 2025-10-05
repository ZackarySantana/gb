package main

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
)

const cacheLine = 64

// Compact: A and B will sit in the same 64-byte cache line on most x86_64 CPUs.
type pair struct {
	A uint64
	B uint64
}

// Padded: put ~56 bytes between A and B so B starts on the next cache line.
type padded struct {
	A uint64
	_ [cacheLine - 8]byte
	B uint64
}

func benchTwoCounters(b *testing.B, p1, p2 *uint64) {
	// Make sure we’re actually scheduling work on (at least) 2 cores.
	runtime.GOMAXPROCS(2)

	// Do enough inner work so the coherence effects dominate goroutine overhead.
	const inner = 1 << 22 // ~4M increments per goroutine per outer iter

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset counters (not strictly necessary for the effect).
		atomic.StoreUint64(p1, 0)
		atomic.StoreUint64(p2, 0)

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			for j := 0; j < inner; j++ {
				atomic.AddUint64(p1, 1)
			}
			wg.Done()
		}()

		go func() {
			for j := 0; j < inner; j++ {
				atomic.AddUint64(p2, 1)
			}
			wg.Done()
		}()

		wg.Wait()
	}
}

func BenchmarkFalseSharing(b *testing.B) {
	var s pair
	benchTwoCounters(b, &s.A, &s.B)
}

func BenchmarkNoFalseSharing(b *testing.B) {
	var s padded
	benchTwoCounters(b, &s.A, &s.B)
}
