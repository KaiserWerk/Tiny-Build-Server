package middleware

import (
	"github.com/KaiserWerk/Tiny-Build-Server/internal/databaseservice"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"github.com/KaiserWerk/sessionstore"
	"github.com/sirupsen/logrus"
)

type MWHandler struct {
	Cfg     *entity.Configuration
	Ds      *databaseservice.DatabaseService
	SessMgr *sessionstore.SessionManager
	Logger  *logrus.Entry
}

func (h *MWHandler) ContextLogger(context string) *logrus.Entry {
	return h.Logger.WithField("context", context)
}
