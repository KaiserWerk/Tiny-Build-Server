package middleware

import (
	"github.com/KaiserWerk/Tiny-Build-Server/internal/configuration"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/dbservice"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/logging"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/sessionservice"
)

type MWHandler struct {
	Cfg     *configuration.AppConfig
	Ds      dbservice.IDBService
	SessSvc sessionservice.ISessionService
	Logger  logging.ILogger
}

func (h *MWHandler) ContextLogger(context string) logging.ILogger {
	return h.Logger.WithField("context", context)
}
