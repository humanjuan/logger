package acacia

import (
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

// 1. Secuencial básico
func BenchmarkSequential(b *testing.B) {
	if testing.Verbose() {
		b.Log("Iniciando benchmark secuencial...")
	}
	tmp := b.TempDir()
	lg, _ := Start("seq.log", tmp, "INFO")
	defer lg.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lg.Info("benchmark sequential %d", i)
	}
}

// 2. 100 goroutines reales
func BenchmarkConcurrent_100Goroutines(b *testing.B) {
	if testing.Verbose() {
		b.Logf("Benchmark con 100 goroutines - b.N = %d", b.N)
	}
	tmp := b.TempDir()
	lg, _ := Start("conc100.log", tmp, "INFO")
	defer lg.Close()

	b.ResetTimer()

	const concurrency = 100
	perGoroutine := b.N / concurrency

	var wg sync.WaitGroup
	for g := 0; g < concurrency; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < perGoroutine; i++ {
				lg.Info("concurrency test")
			}
		}()
	}

	for i := 0; i < b.N%concurrency; i++ {
		lg.Info("main remainder")
	}

	wg.Wait()
}

// 3. Carga extrema
func BenchmarkExtremeLoad(b *testing.B) {
	if testing.Verbose() {
		b.Log("Carga extrema: 500 goroutines + mensajes 1KB")
	}
	tmp := b.TempDir()
	lg, _ := Start("extreme.log", tmp, "INFO")
	defer lg.Close()

	longMsg := strings.Repeat("X", 1024)

	b.ResetTimer()

	const concurrency = 500
	perGoroutine := b.N / concurrency

	var wg sync.WaitGroup
	for g := 0; g < concurrency; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < perGoroutine; i++ {
				lg.Info("EXTREME %s", longMsg)
			}
		}()
	}
	wg.Wait()
}

// 4. Presión infernal
func BenchmarkDroppedUnderPressure(b *testing.B) {
	tmp := b.TempDir()
	lg, _ := Start("pressure.log", tmp, "INFO")
	defer lg.Close()

	var sent atomic.Uint64
	var dropped atomic.Uint64

	b.ResetTimer()

	const concurrency = 1000
	perGoroutine := b.N / concurrency

	var wg sync.WaitGroup
	for g := 0; g < concurrency; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < perGoroutine; i++ {
				select {
				case lg.message <- "pressure test":
					sent.Add(1)
				default:
					dropped.Add(1)
				}
			}
		}()
	}
	wg.Wait()
	b.StopTimer()

	b.Logf("══════════════════════════════════════")
	b.Logf(" BENCHMARK PRESIÓN INFERNAL - RESULTADOS")
	b.Logf("══════════════════════════════════════")
	b.Logf("Total intentados           : %d", b.N)
	b.Logf("Enviados al canal          : %d", sent.Load())
	b.Logf("Dropped (no cabían)        : %d", dropped.Load())
	b.Logf("Dropped reportados por logger: %d", lg.Dropped())
	b.Logf("Tasa de pérdida total      : %.4f%%", float64(dropped.Load()+lg.Dropped())/float64(b.N)*100)
	b.Logf("══════════════════════════════════════")
}

/*
# Básico
go test -bench=Sequential -benchmem

# Concurrencia real
go test -bench=Concurrent_100 -run=^$

# Extremo
go test -bench=Extreme -run=^$

# Presión infernal
go test -bench=Dropped -run=^$
*/
