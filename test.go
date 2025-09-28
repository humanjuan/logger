//go:build ignore
// +build ignore

// This file is an old demo kept for reference. It is excluded from builds and tests.
package logger

func Test() {
	_log, _ := Start("test.log", ".", Level.DEBUG)
	_log.Rotation(40, 5)
	_log.TimestampFormat(TS.Special)

	data1 := 2022
	data2 := []string{"dog", "cat", "fish", "bird"}

	_log.Critical("This is a Critical message %d", 2023)
	_log.Critical("Simple log message without variables")
	_log.Critical("Data: %v, Value: %v", data1, data2)
	_log.Critical(1)
	_log.Critical(1.0)
	_log.Critical(true)
	_log.Critical([]byte{1, 2, 3})
	_log.Critical(map[string]interface{}{})

	_log.Info("This is a Critical message %d", 2023)
	_log.Info("Simple log message without variables")
	_log.Info("Data: %v, Value: %v", data1, data2)
	_log.Info(1)
	_log.Info(1.0)
	_log.Info(true)
	_log.Info([]byte{1, 2, 3})
	_log.Info(map[string]interface{}{})

	_log.Warn("This is a Critical message %d", 2023)
	_log.Warn("Simple log message without variables")
	_log.Warn("Data: %v, Value: %v", data1, data2)
	_log.Warn(1)
	_log.Warn(1.0)
	_log.Warn(true)
	_log.Warn([]byte{1, 2, 3})
	_log.Warn(map[string]interface{}{})

	_log.Error("This is a Critical message %d", 2023)
	_log.Error("Simple log message without variables")
	_log.Error("Data: %v, Value: %v", data1, data2)
	_log.Error(1)
	_log.Error(1.0)
	_log.Error(true)
	_log.Error([]byte{1, 2, 3})
	_log.Error(map[string]interface{}{})

	_log.Debug("This is a Critical message %d", 2023)
	_log.Debug("Simple log message without variables")
	_log.Debug("Data: %v, Value: %v", data1, data2)
	_log.Debug(1)
	_log.Debug(1.0)
	_log.Debug(true)
	_log.Debug([]byte{1, 2, 3})
	_log.Debug(map[string]interface{}{})

	_log.Close()
}
