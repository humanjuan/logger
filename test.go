package logger

func Test() {
	_log, _ := Start("test.log", ".", Level.DEBUG)
	_log.Rotation(40, 5)
	_log.TimestampFormat(TS.Special)

	_log.Critical("This is a Critical message %d", 2023)
	_log.Info("This is an Info message")
	_log.Warn("This is a Warning message")
	_log.Error("This is an Error message")
	_log.Debug("This is a Debug message")

	data1 := "I'm a string value"
	data2 := 2022
	data3 := 15.3
	data4 := false
	data5 := []string{"dog", "cat", "fish", "bird"}
	data6 := map[string]float64{"dog": 1, "cat": 3, "fish": 0}

	_log.Info(data1)
	_log.Info(data2)
	_log.Info(data3)
	_log.Info(data4)
	_log.Info(data5)
	_log.Info(data6)

	_log.Close()
}
