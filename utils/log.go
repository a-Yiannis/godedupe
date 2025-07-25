package utils

import (
	"fmt"
	"log"
	"os"
	"sync"
)

var (
	errorLog      *log.Logger
	logFileHandle *os.File
	logOnce       sync.Once
)

func initLog() {
	var err error
	logFileHandle, err = os.OpenFile("error.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open error log file: %v\n", err)
		return
	}
	errorLog = log.New(logFileHandle, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func CloseLog() {
	if logFileHandle != nil {
		logFileHandle.Close()
	}
}

func PrintE(err error) {
	fmt.Printf("%sError!%s %v\n", red, reset, err)
	logOnce.Do(initLog)
	if errorLog != nil {
		errorLog.Output(2, err.Error())
	}
}

func PrintEf(format string, args ...interface{}) {
	PrintE(fmt.Errorf(format, args...))
}
