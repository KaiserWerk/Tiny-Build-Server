package logging

import (
	"github.com/sirupsen/logrus"
	"io"
	"os"
)

type LogMode byte

const (
	ModeDiscard LogMode = 1 << iota
	ModeConsole
	ModeFile
)

func New(lvl logrus.Level, path, initialContext string, mode LogMode) (*logrus.Entry, func() error, error) {
	l := logrus.New()
	l.SetLevel(lvl)
	l.SetFormatter(&TbsFormatter{
		LevelPadding:   7,
		ContextPadding: 9,
	})
	l.SetReportCaller(false)

	var (
		cf = func() error { return nil }
		w  io.Writer
	)

	switch true {
	case mode&ModeDiscard != 0:
		w = io.Discard
	case mode&ModeConsole != 0 && mode&ModeFile != 0:
		rotator, err := newRotator(path, "tbs.log", 3<<20, 0644, 10)
		if err != nil {
			return nil, nil, err
		}
		cf = func() error {
			return rotator.Close()
		}
		w = io.MultiWriter(rotator, os.Stdout)
	case mode&ModeFile != 0 && mode&ModeConsole == 0:
		rotator, err := newRotator(path, "tbs.log", 3<<20, 0644, 10)
		if err != nil {
			return nil, nil, err
		}
		cf = func() error {
			return rotator.Close()
		}
		w = rotator
	case mode&ModeConsole != 0 && mode&ModeFile == 0:
		w = os.Stdout
	}

	l.SetOutput(w)

	return l.WithField("initialContext", initialContext), cf, nil
}
