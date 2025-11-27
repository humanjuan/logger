package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/humanjuan/acacia/v2"
)

func main() {
	// basicTest()
	//stressTest()
	//stressByteTest()
	//stressJSONTest()
	mixContentTest()
}

func basicTest() {
	lg, err := acacia.Start("acacia.log", "./examples", acacia.Level.DEBUG)
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
	lg, err := acacia.Start("acacia.log", "./examples", acacia.Level.INFO)
	if err != nil {
		panic(err)
	}

	defer lg.Close()

	lg.TimestampFormat(acacia.TS.Special)
	const workers = 500
	const messagesPerWorker = 10_000

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

func stressByteTest() {
	lg, err := acacia.Start("acacia.log", "./examples", acacia.Level.INFO)
	if err != nil {
		panic(err)
	}

	defer lg.Close()

	lg.TimestampFormat(acacia.TS.Special)
	const workers = 500
	const messagesPerWorker = 20_000

	var wg sync.WaitGroup
	wg.Add(workers)

	start := time.Now()

	for id := 1; id <= workers; id++ {
		go func(workerID int) {
			defer wg.Done()

			for i := 1; i <= messagesPerWorker; i++ {
				lg.InfoBytes([]byte(fmt.Sprintf("Goroutine %02d → msg #%05d", workerID, i)))
			}
		}(id)
	}

	wg.Wait()

	elapsed := time.Since(start)
	lg.Info([]byte(fmt.Sprintf("Test concurrente finalizado en %s", elapsed)))

	fmt.Printf("DONE: %d goroutines × %d mensajes = %d líneas\n",
		workers, messagesPerWorker, workers*messagesPerWorker)
}

func stressJSONTest() {
	lg, err := acacia.Start("acacia.json", "./examples", acacia.Level.INFO)
	lg.StructuredJSON(true)
	lg.Rotation(5, 5)
	if err != nil {
		panic(err)
	}

	defer lg.Close()

	lg.TimestampFormat(acacia.TS.Special)
	const workers = 500
	const messagesPerWorker = 10_000

	var wg sync.WaitGroup
	wg.Add(workers)

	start := time.Now()

	for id := 1; id <= workers; id++ {
		go func(workerID int) {
			defer wg.Done()

			for i := 1; i <= messagesPerWorker; i++ {
				lg.Info(
					map[string]interface{}{
						"Goroutine": workerID,
						"msg":       i,
					})
			}
		}(id)
	}

	wg.Wait()

	elapsed := time.Since(start)
	lg.Info(
		map[string]interface{}{
			"msg": fmt.Sprintf("Test concurrente finalizado en %s", elapsed),
		})

	fmt.Printf("DONE: %d goroutines × %d mensajes = %d líneas\n",
		workers, messagesPerWorker, workers*messagesPerWorker)
}

func mixContentTest() {
	lg, err := acacia.Start("acacia.log", "./examples", acacia.Level.INFO)
	if err != nil {
		panic(err)
	}
	lg.Rotation(5, 3)
	defer lg.Close()

	data, err := os.ReadFile("./examples/basic/data.txt")
	if err != nil {
		lg.Error("Error al leer el archivo: %v", err)
	}
	msg := string(data)
	lg.TimestampFormat(acacia.TS.Special)
	const workers = 100
	const messagesPerWorker = 1000

	var wg sync.WaitGroup
	wg.Add(workers)

	start := time.Now()

	for id := 1; id <= workers; id++ {
		go func(workerID int) {
			defer wg.Done()

			for i := 1; i <= messagesPerWorker; i++ {
				lg.Info(msg)
			}
		}(id)
	}

	wg.Wait()

	elapsed := time.Since(start)
	lg.Info("Test concurrente finalizado en %s", elapsed)
	//lg.Info(msg)

}
