package cwlog

import (
	"Facility38/world"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

const MaxBufferLines = 100000

var (
	LogDesc  *os.File
	LogName  string
	LogReady bool

	LogBuf      []string
	LogBufLines int
	LogBufLock  sync.Mutex
)

/*
 * Log this, can use printf arguments
 * Write to buffer, async write
 */
func DoLog(withTrace bool, format string, args ...interface{}) {
	var buf string

	if withTrace {
		/* Get current time */
		ctime := time.Now()
		/* Get calling function and line */
		_, filename, line, _ := runtime.Caller(1)
		/* printf conversion */
		text := fmt.Sprintf(format, args...)
		/* Add current date */
		date := fmt.Sprintf("%2v:%2v.%2v", ctime.Hour(), ctime.Minute(), ctime.Second())
		/* Date, go file, go file line, text */
		buf = fmt.Sprintf("%v: %15v:%5v: %v\n", date, filepath.Base(filename), line, text)
	} else {
		/* Get current time */
		ctime := time.Now()
		/* printf conversion */
		text := fmt.Sprintf(format, args...)
		/* Add current date */
		date := fmt.Sprintf("%2v:%2v.%2v", ctime.Hour(), ctime.Minute(), ctime.Second())
		/* Date, go file, go file line, text */
		buf = fmt.Sprintf("%v: %v\n", date, text)
	}

	if !LogReady || LogDesc == nil {
		fmt.Print(buf)
		return
	}

	/* Add to buffer */
	LogBufLock.Lock()
	LogBuf = append(LogBuf, buf)
	LogBufLines++
	LogBufLock.Unlock()
}

func LogDaemon() {

	go func() {
		for {
			LogBufLock.Lock()

			/* Are there lines to write? */
			if LogBufLines == 0 {
				LogBufLock.Unlock()
				time.Sleep(time.Millisecond * 100)
				continue
			}

			/* Write line */
			_, err := LogDesc.WriteString(LogBuf[0])
			if err != nil {
				fmt.Println("DoLog: WriteString failure")
				LogDesc.Close()
				LogDesc = nil
			}

			if world.LogStdOut {
				fmt.Print(LogBuf[0])
			}

			/* Remove line from buffer */
			LogBuf = LogBuf[1:]
			LogBufLines--

			LogBufLock.Unlock()
		}
	}()
}

/* Prep logger */
func StartLog() {
	t := time.Now()

	/* Create our log file names */
	LogName = fmt.Sprintf("log/cw-%v-%v-%v.log", t.Day(), t.Month(), t.Year())

	/* Make log directory */
	errr := os.MkdirAll("log", os.ModePerm)
	if errr != nil {
		fmt.Print(errr.Error())
		return
	}

	/* Open log files */
	bdesc, errb := os.OpenFile(LogName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	/* Handle file errors */
	if errb != nil {
		fmt.Printf("An error occurred when attempting to create the log. Details: %s", errb)
		return
	}

	/* Save descriptors, open/closed elsewhere */
	LogDesc = bdesc
	LogReady = true
}
