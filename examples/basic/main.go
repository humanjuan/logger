package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/humanjuan/acacia/v2"
)

func main() {
	// basicTest()
	stressTest()
}

func basicTest() {
	lg, err := acacia.Start("basic_test.log", "./examples", acacia.Level.DEBUG)
	if err != nil {
		panic(err)
	}
	defer lg.Close()

	// Optional: choose your preferred timestamp format
	lg.TimestampFormat(acacia.TS.Special)

	lg.Critical("This is a Critical message")
	lg.Info("This is an Informational message %d", 12345)
	lg.Warn("This is a Warning message")
	lg.Error("This is an Error message")
	lg.Debug("This is a Debug message")

	// Optional rotation configuration (defaults are 40MB size, 4 backups)
	// lg.Rotation(80, 5)
}

func stressTest() {
	lg, err := acacia.Start("stress_concurrent.log", "./examples", acacia.Level.INFO)
	if err != nil {
		panic(err)
	}

	defer lg.Close()

	lg.TimestampFormat(acacia.TS.Special)
	const workers = 500
	const messagesPerWorker = 2_000

	var wg sync.WaitGroup
	wg.Add(workers)

	start := time.Now()

	for id := 1; id <= workers; id++ {
		go func(workerID int) {
			defer wg.Done()

			for i := 1; i <= messagesPerWorker; i++ {
				lg.Info("Goroutine %02d → msg #%05d", workerID, i)
			}
		}(id)
	}

	wg.Wait()

	elapsed := time.Since(start)
	lg.Info("Test concurrente finalizado en %s", elapsed)

	fmt.Printf("DONE: %d goroutines × %d mensajes = %d líneas\n",
		workers, messagesPerWorker, workers*messagesPerWorker)
}
