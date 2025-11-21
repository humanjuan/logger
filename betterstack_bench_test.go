// bench_test.go - Compatible con https://github.com/betterstack-community/go-logging-benchmarks
package acacia

import (
	"strings"
	"testing"
)

func BenchmarkHumanJuan(b *testing.B) {
	lg, _ := Start("bench.log", b.TempDir(), "INFO", WithBufferSize(5_000_000), WithBatchSize(512*1024))
	defer lg.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lg.Info("The quick brown fox jumps over the lazy dog")
	}
}

func BenchmarkHumanJuan_Parallel(b *testing.B) {
	lg, _ := Start("bench.log", b.TempDir(), "INFO", WithBufferSize(5_000_000), WithBatchSize(512*1024))
	defer lg.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			lg.Info("The quick brown fox jumps over the lazy dog")
		}
	})
}

func BenchmarkHumanJuan_1KB(b *testing.B) {
	lg, _ := Start("bench.log", b.TempDir(), "INFO", WithBufferSize(5_000_000), WithBatchSize(512*1024))
	defer lg.Close()

	msg := strings.Repeat("X", 1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lg.Info(msg)
	}
}

func BenchmarkHumanJuan_Parallel_1KB(b *testing.B) {
	lg, _ := Start("bench.log", b.TempDir(), "INFO", WithBufferSize(5_000_000), WithBatchSize(512*1024))
	defer lg.Close()

	msg := strings.Repeat("X", 1024)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			lg.Info(msg)
		}
	})
}

/*
# Benchmark básico
go test -bench=BenchmarkHumanJuan -benchmem -benchtime=5s

# Benchmark paralelo (importante para producción)
go test -bench=BenchmarkHumanJuan_Parallel -benchmem -benchtime=5s

# Con mensajes de 1KB
go test -bench=BenchmarkHumanJuan_1KB -benchmem -benchtime=5s

# Paralelo + 1KB (concurrencia real)
go test -bench=BenchmarkHumanJuan_Parallel_1KB -benchmem -benchtime=5s
*/
