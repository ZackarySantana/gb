package main

import (
	"strings"
	"testing"
)

// Simple function to benchmark — pretend this is something heavy.
func concatJoin(n int) string {
	parts := make([]string, n)
	for i := 0; i < n; i++ {
		parts[i] = "bench"
	}
	return strings.Join(parts, "-")
}

// BenchmarkConcatJoin measures performance of concatJoin.
func BenchmarkConcatJoin(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = concatJoin(1000)
	}
}
