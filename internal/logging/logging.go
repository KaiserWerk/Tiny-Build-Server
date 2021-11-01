package logging

import (
	"github.com/KaiserWerk/Tiny-Build-Server/internal/shutdownManager"
	"io"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
)

var (
	err error
	logFile string = "tbs.log" // TODO: log rotation?
	fh *os.File
	centralLogger *logrus.Logger
)

func Init() error {
	shutdownManager.Register(CloseFileHandle)

	centralLogger = logrus.New()
	centralLogger.SetLevel(logrus.TraceLevel)
	centralLogger.SetReportCaller(true)
	centralLogger.SetFormatter(&TbsFormatter{})
	fh, err = os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	centralLogger.SetOutput(io.MultiWriter(fh, os.Stdout))
	return nil
}

func GetLoggerWithContext(context string) *logrus.Entry {
	return centralLogger.WithField("context", context)
}

func GetCentralLogger() *logrus.Logger {
	return centralLogger
}

func CloseFileHandle(wg *sync.WaitGroup) {
	// Flush?
	_ = fh.Close()
	wg.Done()
}