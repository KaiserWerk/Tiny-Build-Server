package panicHandler

import "github.com/sirupsen/logrus"

func Handle(l *logrus.Entry) {
	if r := recover(); r != nil {
		l.Infof("panic: %v", r)
	}
}
