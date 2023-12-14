package logs

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

type ServerLog struct {
	Logger *logrus.Logger
	
	logFileName string
}

func (s *ServerLog) InitLog() {
	s.Logger = logrus.New()
	s.Logger.SetLevel(logrus.DebugLevel)
	now := time.Now()
	s.logFileName = fmt.Sprintf("logs/log-%d-%d-%d.txt", now.Year(), now.Month(), now.Day())
	logFile, err := os.OpenFile(s.logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		s.Logger.Fatal(err)
	}

	s.Logger.SetOutput(io.MultiWriter(os.Stdout, logFile))
}
