package logger

func Test() {
	var _log, _ = Start("test.log", "./", Logger.Level.DEBUG)
	_log.Rotation(40, 5)
	_log.Statistics(true)

	_log.Critical("github.com/jpengineer/logger --> This is a Critical log test")
	_log.Info("github.com/jpengineer/logger --> This is a INFO log test")
	_log.Warn("github.com/jpengineer/logger --> This is a WARNING log test")
	_log.Error("github.com/jpengineer/logger --> This is a ERROR log test")
	_log.Debug("github.com/jpengineer/logger --> This is a DEBUG log test")

	_log.Close()
}
