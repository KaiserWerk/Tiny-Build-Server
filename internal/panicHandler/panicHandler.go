package panicHandler

import "github.com/KaiserWerk/Tiny-Build-Server/internal/logging"

func Handle() {
	if r := recover(); r != nil {
		logging.GetLoggerWithContext("panicHandler").Errorf("%v", r)
	}
}
