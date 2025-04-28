package panichandler

import (
	"github.com/KaiserWerk/Tiny-Build-Server/internal/logging"
)

func Handle(l logging.ILogger) {
	if r := recover(); r != nil {
		l.Infof("panic: %v", r)
	}
}
