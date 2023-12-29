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
	Logger *logrus.Logger

	logFileName string
	once sync.Once
)

func initLog() {
	Logger = logrus.New()
	Logger.SetLevel(logrus.DebugLevel)
	now := time.Now()
	logFileName = fmt.Sprintf("logs/log-%d-%d-%d.txt", now.Year(), now.Month(), now.Day())
	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		Logger.Fatal(err)
	}

	Logger.SetOutput(io.MultiWriter(os.Stdout, logFile))
}

func GetInstance() *logrus.Logger {
	once.Do(initLog)

	return Logger
}
