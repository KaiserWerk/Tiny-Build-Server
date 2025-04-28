package logging

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

type Fields = logrus.Fields

type ILogger interface {
	SetContext(context string) ILogger
	WithField(key string, value any) ILogger
	WithFields(fields Fields) ILogger
	Tracef(format string, args ...any)
	Trace(args ...any)
	Debugf(format string, args ...any)
	Debug(args ...any)
	Infof(format string, args ...any)
	Info(args ...any)
	Warnf(format string, args ...any)
	Warn(args ...any)
	Errorf(format string, args ...any)
	Error(args ...any)
	Fatalf(format string, args ...any)
	Fatal(args ...any)
	Panicf(format string, args ...any)
	Panic(args ...any)
}

type Logger struct {
	l *logrus.Entry
}

func (lg *Logger) SetContext(context string) ILogger {
	return &Logger{
		l: lg.l.WithField("context", context),
	}
}

func (lg *Logger) WithField(key string, value any) ILogger {
	return &Logger{
		l: lg.l.WithField(key, value),
	}
}

func (lg *Logger) WithFields(fields Fields) ILogger {
	return &Logger{
		l: lg.l.WithFields(logrus.Fields(fields)),
	}
}

func (lg *Logger) Tracef(format string, args ...any) {
	lg.l.Tracef(format, args...)
}

func (lg *Logger) Trace(args ...any) {
	lg.l.Trace(args...)
}

func (lg *Logger) Debugf(format string, args ...any) {
	lg.l.Debugf(format, args...)
}

func (lg *Logger) Debug(args ...any) {
	lg.l.Debug(args...)
}

func (lg *Logger) Infof(format string, args ...any) {
	lg.l.Infof(format, args...)
}

func (lg *Logger) Info(args ...any) {
	lg.l.Info(args...)
}

func (lg *Logger) Warnf(format string, args ...any) {
	lg.l.Warnf(format, args...)
}

func (lg *Logger) Warn(args ...any) {
	lg.l.Warn(args...)
}

func (lg *Logger) Errorf(format string, args ...any) {
	lg.l.Errorf(format, args...)
}

func (lg *Logger) Error(args ...any) {
	lg.l.Error(args...)
}

func (lg *Logger) Fatalf(format string, args ...any) {
	lg.l.Fatalf(format, args...)
}

func (lg *Logger) Fatal(args ...any) {
	lg.l.Fatal(args...)
}

func (lg *Logger) Panicf(format string, args ...any) {
	lg.l.Panicf(format, args...)
}

func (lg *Logger) Panic(args ...any) {
	lg.l.Panic(args...)
}

type LogMode byte

const (
	ModeDiscard LogMode = 1 << iota
	ModeConsole
	ModeFile
)

func NewLogger(lvl logrus.Level, path, initialContext string, mode LogMode, filename string) (ILogger, func() error, error) {
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
		rotator, err := newRotator(path, filename, 3<<20, 0644, 10)
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

	return &Logger{l: l.WithField("context", initialContext)}, cf, nil
}
