////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//                                                                                                                    //
//  Author: Juan Alejandro Perez Chandia                                                                              //
//  Contact: jalejandro.ingeniero@gmail.com                                                                           //
//  Website: https://www.jpengineer.cl/                                                                               //
//                                                                                                                    //
//  This module create and write the log files                                                                        //
//                                                                                                                    //
//  Version: 1.6.1                                                                                                    //
//                                                                                                                    //
//                   Include methods that resolve the multiples instances of the logger.                              //
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

package logger

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

var __version__ = "1.6.1"

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
	status            bool
	maxSize           int
	maxRotation       int
	file              *os.File
	stats             bool
	statistic         *statistics
	message           chan string
	wg                sync.WaitGroup
	mtx               sync.Mutex
}

var _timestampFormat = TS.Special

// F o r   S t a t i s t i c s
type statistics struct {
	statsDequeue   int
	statsQueueLen  int
	statsCallWrite int
	rotationCount  int
}

///////////////////////////////////////
//       L O G   M E T H O D S       //
///////////////////////////////////////

func (_log *Log) Statistics(state bool) {
	_log.stats = state
}

func (_log *Log) Rotation(SizeMB int, Backup int) {
	_log.maxSize = SizeMB
	_log.maxRotation = Backup
}

func (_log *Log) Status() bool {
	return _log.status
}

func (_log *Log) Info(data interface{}, args ...interface{}) {
	if _log.level == "DEBUG" || _log.level == "INFO" {
		var message string
		if len(args) > 0 {
			message = fmt.Sprintf(data.(string), args...)
		} else {
			message = fmt.Sprintf("%v", data)
		}
		raw := setFormat(message, "INFO")
		_log.message <- raw
		// S t a t i s t i c s
		_log.statistic.statsCallWrite++
	}
}

func (_log *Log) Warn(data interface{}, args ...interface{}) {
	if _log.level == "DEBUG" || _log.level == "INFO" || _log.level == "WARN" {
		var message string
		if len(args) > 0 {
			message = fmt.Sprintf(data.(string), args...)
		} else {
			message = fmt.Sprintf("%v", data)
		}
		raw := setFormat(message, "WARN")
		_log.message <- raw
		// S t a t i s t i c s
		_log.statistic.statsCallWrite++
	}
}

func (_log *Log) Error(data interface{}, args ...interface{}) {
	if _log.level != "CRITICAL" {
		var message string
		if len(args) > 0 {
			message = fmt.Sprintf(data.(string), args...)
		} else {
			message = fmt.Sprintf("%v", data)
		}
		raw := setFormat(message, "ERROR")
		_log.message <- raw
		// S t a t i s t i c s
		_log.statistic.statsCallWrite++
	}
}

func (_log *Log) Critical(data interface{}, args ...interface{}) {
	var message string
	if len(args) > 0 {
		message = fmt.Sprintf(data.(string), args...)
	} else {
		message = fmt.Sprintf("%v", data)
	}

	raw := setFormat(message, "CRITICAL")
	_log.message <- raw
	// S t a t i s t i c s
	_log.statistic.statsCallWrite++
}

func (_log *Log) Debug(data interface{}, args ...interface{}) {
	if _log.level == "DEBUG" {
		var message string
		if len(args) > 0 {
			message = fmt.Sprintf(data.(string), args...)
		} else {
			message = fmt.Sprintf("%v", data)
		}
		raw := setFormat(message, "DEBUG")
		_log.message <- raw
		// S t a t i s t i c s
		_log.statistic.statsCallWrite++
	}
}

func (_log *Log) Write(p []byte) (n int, err error) {
	_log.mtx.Lock()
	defer _log.mtx.Unlock()

	err = _log.sizeCheck()
	if err != nil {
		return 0, err
	}

	n, err = _log.file.Write(p)
	if err != nil {
		fmt.Println("error: Write()\n", err)
		return n, err
	}

	// S t a t i s t i c s
	_log.statistic.statsQueueLen = len(_log.message)
	_log.statistic.statsDequeue++

	return n, nil
}

func (_log *Log) logRotate() {
	// If exist file .0 then rename it
	tmpFile := _log.file.Name() + ".0"
	_, err := os.Stat(tmpFile)
	if err == nil {

		// S t a t i s t i c s
		_log.statistic.rotationCount += 1
		fmt.Printf("-> [%d] Rotating file %s ...\n", _log.statistic.rotationCount, _log.file.Name())

		for i := _log.maxRotation; i >= 0; i = i - 1 {
			tmpFile = _log.file.Name() + "." + strconv.Itoa(i)
			_, err = os.Stat(tmpFile)

			// R o t a t i o n   l i m i t
			if err == nil && i == _log.maxRotation {
				err := os.Remove(_log.file.Name() + "." + strconv.Itoa(i))
				if err != nil {
					return
				}
				err = os.Rename(_log.file.Name()+"."+strconv.Itoa(i-1), _log.file.Name()+"."+strconv.Itoa(i))
				if err != nil {
					return
				}
			} else if err == nil {
				err = os.Rename(_log.file.Name()+"."+strconv.Itoa(i), _log.file.Name()+"."+strconv.Itoa(i+1))
				if err != nil {
					return
				}
			}
		}
	}
}

func (_log *Log) sizeCheck() error {
	currentSize, _ := _log.fileSize()
	var err error

	if currentSize >= float64(_log.maxSize) {
		err := _log.file.Close()
		if err != nil {
			return err
		}
		newName := _log.file.Name() + ".0"
		oldName := _log.file.Name()

		err = os.Rename(oldName, newName)
		if err != nil {
			return err
		}
		file, err := os.OpenFile(oldName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		*_log.file = *file
		_log.logRotate()
		return err
	}
	return err
}

func (_log *Log) fileSize() (float64, error) {
	info, err := _log.file.Stat()
	if err != nil {
		return float64(0), err
	}
	bytes := info.Size()
	kilobytes := bytes / 1024
	megabytes := (float64)(kilobytes / 1024)

	return megabytes, err
}

func (_log *Log) Close() {
	close(_log.message)
	_log.wg.Wait()

	if _log.stats {
		fmt.Println("====== S T A T I S T I C S ======")
		fmt.Println("File  Name:", _log.name)
		fmt.Println("Dequeue:", _log.statistic.statsDequeue)
		fmt.Println("Queue Length into Logger (func Close):", _log.statistic.statsQueueLen)
		fmt.Println("Total Call to Write:", _log.statistic.statsCallWrite)
	}

	time.Sleep(1 * time.Second)
	err := _log.file.Close()
	if err != nil {
		return
	}
}

func (_log *Log) TimestampFormat(format string) {
	_timestampFormat = format

}

///////////////////////////////////////
//  P U B L I C   F U N C T I O N S  //
///////////////////////////////////////

func Start(logName string, logPath string, logLevel string) (*Log, error) {
	var _log Log
	var header string

	// P a t h   V a l i d a t i o n
	if logPath[len(logPath)-1:] != "/" {
		logPath = logPath + "/"
	}
	logLevel = strings.ToUpper(logLevel)

	stat, err := os.Stat(logPath)
	if err != nil || !stat.IsDir() {
		err := fmt.Errorf("error: The path %s does not exist", logPath)
		return nil, err
	}

	// V e r i f y   l o g   l e v e l
	if !verifyLevel(logLevel) {
		logLevel = Level.INFO
		fmt.Println("warning: The log level has been configured in \"INFO\" by default.")
	}

	// O p e n / C r e a t e   L o g   f i l e
	tmpFile, err := os.OpenFile(logPath+logName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)

	if err != nil {
		return nil, err
	}

	// F i l e   h e a d e r
	_, srcFile, _, ok := runtime.Caller(1)
	_, file := filepath.Split(srcFile)

	if ok {
		header = "Logger Version: " + __version__ + " SourceFile: " + file + " Hash: " + calculateHash(srcFile)
	} else {
		header = "Logger Version: " + __version__
	}

	_, err = tmpFile.WriteString(header + "\n")
	if err != nil {
		return nil, err
	}
	// C r e a t e   b a s i c   l o g g e r
	_log = Log{
		name:        logName,
		path:        logPath,
		level:       logLevel,
		maxSize:     40,
		maxRotation: 4,
		status:      true,
		file:        tmpFile,
		stats:       false,
		statistic:   new(statistics),
		message:     make(chan string, 1),
	}

	// S t a r t
	_log.wg.Add(1)
	go _log.startWriting()

	return &_log, nil
}

///////////////////////////////////////
// P R I V A T E   F U N C T I O N S //
///////////////////////////////////////

func (_log *Log) startWriting() {
	defer _log.wg.Done()

	for msg := range _log.message {
		_, err := _log.file.Write([]byte(msg))
		if err != nil {
			fmt.Println("error: Write()\n", err)
		}
	}
}

func verifyLevel(lvl string) bool {
	var fields = reflect.TypeOf(Level)
	var values = reflect.ValueOf(Level)
	num := fields.NumField()
	result := false
	for i := 0; i < num; i++ {
		value := values.Field(i)
		if value.String() == lvl {
			result = true
			break
		}
	}
	return result
}

func calculateHash(file string) string {
	f, err := os.Open(file)
	if err != nil {
		return ""
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			fmt.Println("Error on calculateHash() when Close()")
			fmt.Println(err)
		}
	}(f)

	_hash_ := sha256.New()
	_, err = io.Copy(_hash_, f)
	if err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(_hash_.Sum(nil))
}

func setFormat(msg string, lvl string) string {
	var message = getTime() + " [" + lvl + "] " + msg + "\n"
	return message
}

func getTime() string {
	dt := time.Now()
	return dt.Format(_timestampFormat)
}
