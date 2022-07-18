package handler

import (
	"github.com/KaiserWerk/sessionstore/v2"
	"github.com/sirupsen/logrus"

	"github.com/KaiserWerk/Tiny-Build-Server/internal/buildservice"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/configuration"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/dbservice"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/deploymentservice"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/mailer"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/templateservice"
)

type HttpHandler struct {
	Configuration *configuration.AppConfig
	DBService     *dbservice.DBService
	BuildService  *buildservice.BuildService
	DeployService *deploymentservice.DeploymentService
	SessMgr       *sessionstore.SessionManager
	Logger        *logrus.Entry
	Mailer        *mailer.Mailer
}

func (h *HttpHandler) ContextLogger(context string) *logrus.Entry {
	return h.Logger.WithField("context", context)
}

func (h *HttpHandler) Injector() *templateservice.Injector {
	return &templateservice.Injector{
		Logger:  h.Logger,
		SessMgr: h.SessMgr,
		Ds:      h.DBService,
	}
}
