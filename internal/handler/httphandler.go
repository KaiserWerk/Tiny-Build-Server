package handler

import (
	"github.com/KaiserWerk/Tiny-Build-Server/internal/buildservice"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/configuration"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/dbservice"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/deploymentservice"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/logging"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/mailer"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/sessionservice"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/templateservice"
)

type HTTPHandler struct {
	Configuration  *configuration.AppConfig
	DBService      dbservice.IDBService
	BuildService   buildservice.IBuildService
	DeployService  deploymentservice.IDeploymentService
	SessionService sessionservice.ISessionService
	Logger         logging.ILogger
	Mailer         mailer.IMailer
}

func (h *HTTPHandler) ContextLogger(context string) logging.ILogger {
	return h.Logger.SetContext(context)
}

func (h *HTTPHandler) Injector() *templateservice.Injector {
	return &templateservice.Injector{
		Logger:         h.Logger,
		SessionService: h.SessionService,
		Ds:             h.DBService,
	}
}
