package handler

import (
	"github.com/KaiserWerk/Tiny-Build-Server/internal/databaseservice"
	"github.com/KaiserWerk/sessionstore"
	"github.com/sirupsen/logrus"
)

type HttpHandler struct {
	Ds *databaseservice.DatabaseService
	SessMgr *sessionstore.SessionManager
	Logger *logrus.Logger
}

func (h *HttpHandler) ContextLogger(context string) *logrus.Entry {
	return h.Logger.WithField("context", context)
}