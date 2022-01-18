package handler

import (
	"github.com/KaiserWerk/Tiny-Build-Server/internal/buildservice"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/databaseservice"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/templateservice"
	"github.com/KaiserWerk/sessionstore"
	"github.com/sirupsen/logrus"
)

type HttpHandler struct {
	Cfg     *entity.Configuration
	Ds      *databaseservice.DatabaseService
	Bs      *buildservice.BuildService
	SessMgr *sessionstore.SessionManager
	Logger  *logrus.Entry
}

func (h *HttpHandler) ContextLogger(context string) *logrus.Entry {
	return h.Logger.WithField("context", context)
}

func (h *HttpHandler) Injector() *templateservice.Injector {
	return &templateservice.Injector{
		Logger:  h.Logger,
		SessMgr: h.SessMgr,
		Ds:      h.Ds,
	}
}
