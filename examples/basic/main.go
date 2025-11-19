package main

import (
	"github.com/humanjuan/acacia"
)

func main() {
	logName := "MyLogName.log"
	path := "." // write in current directory for the demo
	level := acacia.Level.DEBUG

	lg, err := acacia.Start(logName, path, level)
	if err != nil {
		panic(err)
	}
	// Optional: choose your preferred timestamp format
	lg.TimestampFormat(acacia.TS.Special)

	lg.Critical("This is a Critical message")
	lg.Info("This is an Informational message %d", 12345)
	lg.Warn("This is a Warning message")
	lg.Error("This is an Error message")
	lg.Debug("This is a Debug message")

	// Optional rotation configuration (defaults are 40MB size, 4 backups)
	// lg.Rotation(80, 5)

	lg.Close()
}
