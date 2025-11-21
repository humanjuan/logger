////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//                                                                                                                    //
//  Author: Juan Alejandro Perez Chandia                                                                              //
//  Contact: juan.alejandro@humanjuan.com                                                                             //
//  Website: https://humanjuan.com/                                                                                   //
//                                                                                                                    //
//  HumanJuan Acacia - High-performance concurrent logger with real file rotation                                     //
//                                                                                                                    //
//  Version: 2.1.0                                                                                                    //
//                                                                                                                    //
//	MIT License                                                                                                       //
//	                                                                                                                  //
//	Copyright (c) 2020 Juan Alejandro                                                                                 //
//	                                                                                                                  //
//	Permission is hereby granted, free of charge, to any person obtaining a copy                                      //
//	of this software and associated documentation files (the "Software"), to deal                                     //
//	in the Software without restriction, including without limitation the rights                                      //
//	to use, copy, modify, merge, publish, distribute, sublicense, and/or sell                                         //
//	copies of the Software, and to permit persons to whom the Software is                                             //
//	furnished to do so, subject to the following conditions:                                                          //
//	                                                                                                                  //
//	The above copyright notice and this permission notice shall be included in all                                    //
//	copies or substantial portions of the Software.                                                                   //
//	                                                                                                                  //
//	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR                                        //
//	IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,                                          //
//	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE                                       //
//	AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER                                            //
//	LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,                                     //
//	OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE                                     //
//	SOFTWARE.                                                                                                         //
//                                                                                                                    //
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

package acacia

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	version           = "2.1.0"
	DefaultBufferSize = 500_000
	MinBufferSize     = 1_000
	DefaultBatchSize  = 64 * 1024 // 64 kb
	flushInterval     = 100 * time.Millisecond
	lastDayFormat     = "2006-01-02"
)

var (
	timestampFormat = TS.Special
)

type config struct {
	bufferSize int
	batchSize  int
}

type Option func(*config)

func WithBufferSize(number int) Option {
	return func(conf *config) {
		if number >= MinBufferSize {
			conf.bufferSize = number
		}
	}
}

func WithBatchSize(number int) Option {
	return func(conf *config) {
		if number > 1024 {
			conf.batchSize = number
		}
	}
}

type tsFormat struct {
	ANSIC       string // "Mon Jan _2 15:04:05 2006"
	UnixDate    string // "Mon Jan _2 15:04:05 MST 2006"
	RubyDate    string // "Mon Jan 02 15:04:05 -0700 2006"
	RFC822      string // "02 Jan 06 15:04 MST"
	RFC822Z     string // "02 Jan 06 15:04 -0700"
	RFC850      string // "Monday, 02-Jan-06 15:04:05 MST"
	RFC1123     string // "Mon, 02 Jan 2006 15:04:05 MST"
	RFC1123Z    string // "Mon, 02 Jan 2006 15:04:05 -0700"
	RFC3339     string // "2006-01-02T15:04:05Z07:00"
	RFC3339Nano string // "2006-01-02T15:04:05.999999999Z07:00"
	Kitchen     string // "3:04PM"
	Special     string // "Jan 2, 2006 15:04:05.000000 MST"
	Stamp       string // "Jan _2 15:04:05"
	StampMilli  string // "Jan _2 15:04:05.000"
	StampMicro  string // "Jan _2 15:04:05.000000"
	StampNano   string // "Jan _2 15:04:05.000000000"

}

var TS = tsFormat{
	ANSIC:       "Mon Jan _2 15:04:05 2006",
	UnixDate:    "Mon Jan _2 15:04:05 MST 2006",
	RubyDate:    "Mon Jan 02 15:04:05 -0700 2006",
	RFC822:      "02 Jan 06 15:04 MST",
	RFC822Z:     "02 Jan 06 15:04 -0700",
	RFC850:      "Monday, 02-Jan-06 15:04:05 MST",
	RFC1123:     "Mon, 02 Jan 2006 15:04:05 MST",
	RFC1123Z:    "Mon, 02 Jan 2006 15:04:05 -0700",
	RFC3339:     "2006-01-02T15:04:05Z07:00",
	RFC3339Nano: "2006-01-02T15:04:05.999999999Z07:00",
	Kitchen:     "3:04PM",
	Special:     "Jan 2, 2006 15:04:05.000000 MST",
	Stamp:       "Jan _2 15:04:05",
	StampMilli:  "Jan _2 15:04:05.000",
	StampMicro:  "Jan _2 15:04:05.000000",
	StampNano:   "Jan _2 15:04:05.000000000",
}

type getLevel struct {
	DEBUG    string
	INFO     string
	WARN     string
	ERROR    string
	CRITICAL string
}

var Level = getLevel{
	// DEBUG < INFO < WARN < ERROR < CRITICAL
	DEBUG:    "DEBUG",
	INFO:     "INFO",
	WARN:     "WARN",
	ERROR:    "ERROR",
	CRITICAL: "CRITICAL",
}

type Log struct {
	name, path, level string
	structured        bool
	status            bool
	maxSize           int64
	maxRotation       int
	daily             bool
	lastDay           string
	file              atomic.Value
	stats             bool
	statistic         *statistics
	message           chan string
	wg                sync.WaitGroup
	mtx               sync.Mutex
	buffer            []byte
}

// F o r   S t a t i s t i c s
type statistics struct {
	statsDequeue   int64
	statsQueueLen  int64
	statsCallWrite int64
	rotationCount  int64
}

///////////////////////////////////////
//       L O G   M E T H O D S       //
///////////////////////////////////////

func (_log *Log) Statistics(state bool) {
	_log.stats = state
}

func (_log *Log) StructuredJSON(state bool) {
	_log.structured = state
}

func (_log *Log) Status() bool {
	return _log.status
}

// Dropped está deprecado. Desde v2.2 el logger aplica backpressure y no
// descarta mensajes. Este método se conserva por compatibilidad y siempre
// retorna 0.
func (_log *Log) Dropped() uint64 { return 0 }

func (_log *Log) logf(level string, data interface{}, args ...interface{}) {
	if !_log.shouldLog(level) {
		return
	}

	msg := _log.formatMessage(data, args...)
	raw := _log.setFormat(msg, level)

	// Envío bloqueante: aplica backpressure y garantiza cero pérdidas.
	_log.message <- raw
	atomic.AddInt64(&_log.statistic.statsCallWrite, 1)
}

func (_log *Log) shouldLog(level string) bool {
	switch _log.level {
	case "DEBUG":
		return true
	case "INFO":
		return level == "INFO" || level == "WARN" || level == "ERROR" || level == "CRITICAL"
	case "WARN":
		return level == "WARN" || level == "ERROR" || level == "CRITICAL"
	case "ERROR":
		return level == "ERROR" || level == "CRITICAL"
	case "CRITICAL":
		return level == "CRITICAL"
	}
	return false
}

func (_log *Log) Info(data interface{}, args ...interface{}) {
	_log.logf("INFO", data, args...)
}

func (_log *Log) Warn(data interface{}, args ...interface{}) {
	_log.logf("WARN", data, args...)
}

func (_log *Log) Error(data interface{}, args ...interface{}) {
	_log.logf("ERROR", data, args...)
}

func (_log *Log) Critical(data interface{}, args ...interface{}) {
	_log.logf("CRITICAL", data, args...)
}

func (_log *Log) Debug(data interface{}, args ...interface{}) {
	_log.logf("DEBUG", data, args...)
}

func (_log *Log) formatMessage(data interface{}, args ...interface{}) string {
	if len(args) > 0 {
		return fmt.Sprintf(data.(string), args...)
	}
	return fmt.Sprintf("%v", data)
}

func (_log *Log) Write(p []byte) (int, error) {
	if !_log.shouldLog("INFO") {
		return len(p), nil
	}
	msg := string(p)
	if msg != "" && msg[len(msg)-1] != '\n' {
		msg += "\n"
	}
	_log.logf("INFO", msg)
	return len(p), nil
}

func (_log *Log) Rotation(sizeMB int, backup int) {
	if sizeMB <= 0 {
		_log.maxSize = 0
		_log.maxRotation = 0
		return
	}
	if backup < 1 {
		backup = 1
	}
	_log.maxSize = int64(sizeMB) * 1024 * 1024
	_log.maxRotation = backup
}

func (_log *Log) DailyRotation(enabled bool) {
	_log.mtx.Lock()
	_log.daily = enabled
	if enabled {
		_log.lastDay = time.Now().Format(lastDayFormat)
	}
	_log.mtx.Unlock()
	// Si se habilita la rotación diaria, rotamos inmediatamente el archivo actual
	// a un nombre con la fecha de hoy (p. ej., app.log → app-YYYY-MM-DD.log),
	// para alinear con la expectativa de los tests y simplificar el ciclo diario.
	if enabled {
		_ = _log.rotateByDate(_log.lastDay)
	}
}

// app.log → app-2025-11-18.log
// app.log.0 → app-2025-11-18.log.0
// app.log.1 → app-2025-11-18.log.1
func (_log *Log) rotateByDate(oldDay string) error {
	_log.mtx.Lock()
	base := _log.getFile().Name()
	dir, name := filepath.Dir(base), filepath.Base(base)
	oldFile := _log.getFile()
	_log.mtx.Unlock()

	oldFile.Close()

	dated := filepath.Join(dir, name+"-"+oldDay)
	os.Rename(base, dated)

	for i := 0; ; i++ {
		bak := fmt.Sprintf("%s.%d", base, i)
		if _, err := os.Stat(bak); err != nil {
			break
		}
		os.Rename(bak, dated+"."+strconv.Itoa(i))
	}

	newFile, _ := os.OpenFile(base, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	_log.setFile(newFile)
	return nil
}

func (_log *Log) logRotate() error {
	_log.mtx.Lock()
	base := _log.getFile().Name()
	oldFile := _log.getFile()
	maxRot := _log.maxRotation
	_log.mtx.Unlock()

	oldFile.Close()

	for i := maxRot - 1; i >= 0; i-- {
		src := fmt.Sprintf("%s.%d", base, i)
		dst := fmt.Sprintf("%s.%d", base, i+1)
		if _, err := os.Stat(src); err == nil {
			os.Rename(src, dst)
		}
	}
	os.Rename(base, base+".0")

	newFile, _ := os.OpenFile(base, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	_log.setFile(newFile)
	atomic.AddInt64(&_log.statistic.rotationCount, 1)
	return nil
}

func (_log *Log) sizeCheck() error {
	if _log.maxSize <= 0 {
		return nil
	}

	_log.mtx.Lock()
	f := _log.getFile()
	_log.mtx.Unlock()

	if f == nil {
		return nil
	}

	info, err := f.Stat()
	if err != nil {
		return err
	}

	if info.Size() < _log.maxSize {
		return nil
	}

	return _log.logRotate()
}

func (_log *Log) fileSize() (int64, error) {
	info, err := _log.getFile().Stat()
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

func (_log *Log) Close() {
	close(_log.message)
	_log.wg.Wait()

	if _log.stats {
		fmt.Println("====== LOGGER STATISTICS ======")
		fmt.Printf("File: %s\n", _log.name)
		fmt.Printf("Dequeue: %d\n", atomic.LoadInt64(&_log.statistic.statsDequeue))
		fmt.Printf("Queue Length (at close): %d\n", atomic.LoadInt64(&_log.statistic.statsQueueLen))
		fmt.Printf("Total Write Calls: %d\n", atomic.LoadInt64(&_log.statistic.statsCallWrite))
		fmt.Printf("Rotations: %d\n", atomic.LoadInt64(&_log.statistic.rotationCount))
	}

	if f := _log.getFile(); f != nil {
		f.Close()
	}
}

///////////////////////////////////////
//  P U B L I C   F U N C T I O N S  //
///////////////////////////////////////

func Start(logName, logPath, logLevel string, opts ...Option) (*Log, error) {
	if logPath[len(logPath)-1:] != "/" {
		logPath += "/"
	}
	logLevel = strings.ToUpper(logLevel)

	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("path %s does not exist", logPath)
	}

	if !verifyLevel(logLevel) {
		fmt.Println("warning: invalid log level, falling back to INFO")
		logLevel = "INFO"
	}

	fullPath := logPath + logName
	f, err := os.OpenFile(fullPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	cfg := &config{
		bufferSize: DefaultBufferSize,
		batchSize:  DefaultBatchSize,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	header := fmt.Sprintf("=== HumanJuan Logger v%s started at %s ===\n", version, time.Now().Format(time.RFC3339))
	_, _ = f.WriteString(header)

	log := &Log{
		name:        logName,
		path:        logPath,
		level:       logLevel,
		maxSize:     0,
		maxRotation: 0,
		daily:       false,
		lastDay:     time.Now().Format(lastDayFormat),
		status:      true,
		stats:       false,
		statistic:   &statistics{},
		message:     make(chan string, cfg.bufferSize),
		buffer:      make([]byte, 0, cfg.batchSize),
	}

	log.file.Store(f)
	log.wg.Add(1)
	go log.startWriting()

	return log, nil
}

///////////////////////////////////////
// P R I V A T E   F U N C T I O N S //
///////////////////////////////////////

func (_log *Log) startWriting() {
	defer _log.wg.Done()
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	// Buffer de lote reutilizable para reducir la contención del mutex
	batch := make([]string, 0, 1024)
	const drainLimit = 1000

	for {
		select {
		case first, ok := <-_log.message:
			if !ok {
				// Volcar cualquier resto del batch y hacer flush final
				if len(batch) > 0 {
					_log.mtx.Lock()
					for i := range batch {
						_log.buffer = append(_log.buffer, batch[i]...)
					}
					_log.mtx.Unlock()
					batch = batch[:0]
				}
				_log.flush()
				return
			}

			// Arranca el lote con el primer mensaje
			batch = append(batch, first)

			// Drenaje no bloqueante hasta agotar o alcanzar límite por lote
			for i := 1; i < drainLimit; i++ {
				select {
				case msg := <-_log.message:
					batch = append(batch, msg)
				default:
					i = drainLimit // salir del bucle
				}
			}

			// Un solo lock para volcar todo el lote al buffer principal
			_log.mtx.Lock()
			for i := range batch {
				_log.buffer = append(_log.buffer, batch[i]...)
			}
			shouldFlush := len(_log.buffer) >= cap(_log.buffer)/2
			_log.mtx.Unlock()

			// Reutilizar batch sin realloc
			batch = batch[:0]

			if shouldFlush {
				_log.flush()
			}

		case <-ticker.C:
			_log.flush()
		}
	}
}

func (_log *Log) Sync() {
	select {
	case _log.message <- "":
	default:
	}
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		_log.mtx.Lock()
		empty := len(_log.message) == 0 && len(_log.buffer) == 0
		_log.mtx.Unlock()
		if empty {
			if f := _log.getFile(); f != nil {
				_ = f.Sync()
			}
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func (_log *Log) flush() {
	_log.mtx.Lock()
	bufferCopy := append([]byte(nil), _log.buffer...)
	// Evaluar rotación diaria correctamente: solo cuando está habilitada y cambió el día
	needDaily := false
	if _log.daily {
		today := time.Now().Format(lastDayFormat)
		needDaily = (today != _log.lastDay)
	}
	// Evaluar rotación por tamaño: solo si excede el umbral
	needSize := false
	var currentFile *os.File
	if f := _log.getFile(); f != nil {
		info, _ := f.Stat()
		if info != nil && _log.maxSize > 0 {
			needSize = info.Size()+int64(len(_log.buffer)) >= _log.maxSize
		}
		currentFile = f
	}
	_log.buffer = _log.buffer[:0]
	_log.mtx.Unlock()

	if needDaily {
		oldDay := _log.lastDay
		_ = _log.rotateByDate(oldDay)
		_log.mtx.Lock()
		_log.lastDay = time.Now().Format(lastDayFormat)
		_log.mtx.Unlock()
	}
	if needSize {
		_ = _log.logRotate()
	}
	if len(bufferCopy) > 0 && currentFile != nil {
		currentFile.Write(bufferCopy)
	}
}

func (_log *Log) setFormat(msg, level string) string {
	ts := time.Now().Format(timestampFormat)
	if !_log.structured {
		return fmt.Sprintf("%s [%s] %s\n", ts, level, msg)
	}
	return fmt.Sprintf(`{"ts":"%s","level":"%s","msg":%q}`+"\n", ts, level, msg)
}

func (_log *Log) TimestampFormat(format string) {
	timestampFormat = format
}

func verifyLevel(lvl string) bool {
	v := reflect.ValueOf(Level)
	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).String() == lvl {
			return true
		}
	}
	return false
}

func (_log *Log) getFile() *os.File {
	if v := _log.file.Load(); v != nil {
		return v.(*os.File)
	}
	return nil
}

func (_log *Log) setFile(f *os.File) {
	if f != nil {
		_log.file.Store(f)
	} else {
		_log.file.Store((*os.File)(nil))
	}
}
