package cwlog

import (
	"GameTest/consts"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

var LogDesc *os.File
var LogName string

/* Normal CW log */
func DoLog(format string, args ...interface{}) {
	if !consts.Debug {
		return
	}
	ctime := time.Now()
	_, filename, line, _ := runtime.Caller(1)

	text := fmt.Sprintf(format, args...)

	date := fmt.Sprintf("%2v:%2v.%2v", ctime.Hour(), ctime.Minute(), ctime.Second())
	buf := fmt.Sprintf("%v: %15v:%5v: %v\n", date, filepath.Base(filename), line, text)
	_, err := LogDesc.WriteString(buf)

	if consts.LogStdOut {
		fmt.Print(buf)
	}
	if err != nil {
		fmt.Println("DoLog: WriteString failure")
		LogDesc.Close()
		LogDesc = nil
		return
	}

}

/* Prep everything for the cw log */
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
}
