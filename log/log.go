package log

import (
	"fmt"
	"io"
	"log"
	"os"
)

const (
	pfxDebug = "DEBUG\t"
	pfxInfo  = "INFO\t"
	pfxWarn  = "WARN\t"
	pfxError = "ERROR\t"
	pfxFatal = "FATAL\t"
)

var (
	logger *log.Logger
)

func Init(filePath string) {
	mv := io.Writer(os.Stdout)
	if filePath != "" {
		file, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
		if err != nil {
			fmt.Println(err)
		}
		mv = io.MultiWriter(os.Stdout, file)
	}
	logger = log.New(mv, "", log.LstdFlags|log.Lmicroseconds|log.Llongfile)
}

func write(prefix string, v ...interface{}) {
	logger.SetPrefix(prefix)
	logger.Println(v...)
}

func writeF(prefix, fmt string, v ...interface{}) {
	logger.SetPrefix(prefix)
	logger.Printf(fmt, v...)
}

func Debug(v ...interface{}) {
	write(pfxDebug, v...)
}

func Info(v ...interface{}) {
	write(pfxInfo, v...)
}

func Warn(v ...interface{}) {
	write(pfxWarn, v...)
}

func Error(v ...interface{}) {
	write(pfxError, v...)
}

func Fatal(fmt string, v ...interface{}) {
	write(pfxFatal, v...)
}

func DebugF(fmt string, v ...interface{}) {
	writeF(pfxDebug, fmt, v...)
}

func InfoF(fmt string, v ...interface{}) {
	writeF(pfxInfo, fmt, v...)
}

func WarnF(fmt string, v ...interface{}) {
	writeF(pfxWarn, fmt, v...)
}

func ErrorF(fmt string, v ...interface{}) {
	writeF(pfxError, fmt, v...)
}

func FatalF(fmt string, v ...interface{}) {
	writeF(pfxFatal, fmt, v...)
}
