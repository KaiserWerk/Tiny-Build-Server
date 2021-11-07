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
	rotator *Rotator
)

func New(lvl logrus.Level, context string, toConsole bool) *logrus.Entry {
	l := logrus.New()
	l.SetLevel(lvl)
	l.SetFormatter(&TbsFormatter{
		LevelPadding: 7,
		ContextPadding: 9,
	})
	l.SetReportCaller(false)
	if toConsole {
		l.SetOutput(io.MultiWriter(rotator, os.Stdout))
	} else {
		l.SetOutput(rotator)
	}

	return l.WithField("context", context)
}

func Init(dir string) {
	shutdownManager.Register(CloseFileHandle)
	rotator, err = NewRotator(dir, "tbs.log", 10 << 20, 0644)
	if err != nil {
		panic("cannot create rotator: " + err.Error())
	}
}

func CloseFileHandle(wg *sync.WaitGroup) {
	_ = rotator.Close()
	wg.Done()
}