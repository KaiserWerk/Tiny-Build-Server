package panicHandler

import (
	"github.com/KaiserWerk/Tiny-Build-Server/internal/logging"

	"github.com/sirupsen/logrus"
)

func Handle() {
	if r := recover(); r != nil {
		l := logging.New(logrus.TraceLevel, "panicHandler", false)
		l.Info("panic: %v", r)
	}
}
