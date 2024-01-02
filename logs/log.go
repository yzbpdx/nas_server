package logs

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	loggerInstance *Log
	mutex sync.Mutex
)

type Log struct {
	Logger *logrus.Logger
	
	logFile *os.File
	logFileName string
}

func (l *Log) initLog() {
	l.Logger = logrus.New()
	l.Logger.SetLevel(logrus.DebugLevel)
	now := time.Now()
	l.logFileName = fmt.Sprintf("logs/log-%d-%d-%d.txt", now.Year(), now.Month(), now.Day())
	logFileTemp, err := os.OpenFile(l.logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		l.Logger.Fatal(err)
	}
	l.logFile = logFileTemp

	l.Logger.SetOutput(io.MultiWriter(os.Stdout, l.logFile))
}

func GetInstance() *Log {
	if loggerInstance == nil {
		mutex.Lock()
		defer mutex.Unlock()

		if loggerInstance == nil {
			loggerInstance = &Log{}
			loggerInstance.initLog()
		}
	}

	return loggerInstance
}

func (l *Log) CloseLogFile() {
	l.logFile.Close()
}
