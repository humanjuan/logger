package logger

func Test() {
	_log, _ := Start("test.log", ".", Level.DEBUG)
	_log.Rotation(40, 5)
	_log.TimestampFormat(TS.Special)

	_log.Critical("This is a Critical message")
	_log.Info("This is an Info message")
	_log.Warn("This is a Warning message")
	_log.Error("This is an Error message")
	_log.Debug("This is a Debug message")

	_log.Close()
}
