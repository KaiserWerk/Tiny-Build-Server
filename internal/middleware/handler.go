package middleware

import (
	"github.com/KaiserWerk/Tiny-Build-Server/internal/configuration"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/dbservice"
	"github.com/KaiserWerk/sessionstore/v2"
	"github.com/sirupsen/logrus"
)

type MWHandler struct {
	Cfg     *configuration.AppConfig
	Ds      *dbservice.DBService
	SessMgr *sessionstore.SessionManager
	Logger  *logrus.Entry
}

func (h *MWHandler) ContextLogger(context string) *logrus.Entry {
	return h.Logger.WithField("context", context)
}
