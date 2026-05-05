package utils

import (
	"log"
	"os"
)

var (
	InfoLog  = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime)
	ErrorLog = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime)
)

func Info(format string, v ...interface{}) {
	InfoLog.Printf(format, v...)
}

func Error(format string, v ...interface{}) {
	ErrorLog.Printf(format, v...)
}
