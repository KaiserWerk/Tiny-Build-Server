package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/databaseservice"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/global"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/handler"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/helper"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/logging"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/middleware"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/shutdownManager"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/templateservice"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	Version     = "0.0.0-dev"
	VersionDate = "0000-00-00 00:00:00 +00:00"
	listenPort string
	configFile string
	err error
)

func main() {
	global.Set(Version, VersionDate)
	err = logging.Init()
	if err != nil {
		panic(err.Error())
	}
	defer shutdownManager.Initiate()

	logger := logging.GetLoggerWithContext("main")

	flag.StringVar(&listenPort, "port", "8271", "The port which the build server should listen on")
	flag.StringVar(&configFile, "config", "", "The location of the configuration file")
	flag.Parse()

	if configFile != "" {
		global.SetConfigurationFile(configFile)
	}

	logger.WithFields(logrus.Fields{
		"app": "Tiny Build Server",
		"version": global.Version,
		"versionDate": global.VersionDate,
		"port": listenPort,
		"configFile": global.GetConfigurationFile(),
	}).Info("app information")

	config := global.GetConfiguration()

	ds := databaseservice.Get()
	err := ds.AutoMigrate()
	if err != nil {
		logger.Panic("AutoMigrate panic: " + err.Error())
	}
	//ds.Quit()

	listenAddr := fmt.Sprintf(":%s", listenPort)
	logger.Trace("Server starts handling requests")
	if config.Tls.Enabled {
		logger.Debug("  TLS is enabled")
	}

	router := setupRoutes(config)

	tlsConfig := &tls.Config{
		MinVersion:               tls.VersionTLS12,
		CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
		},
	}

	server := &http.Server{
		Addr:              listenAddr,
		Handler:           router,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       10 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}

	if config.Tls.Enabled {
		server.TLSConfig = tlsConfig
		server.TLSNextProto = make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0)
	}

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, os.Kill, syscall.SIGTERM)

	//go helper.ReadConsoleInput(quit)

	go func() {
		<-quit
		logger.Debug("Server is shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		server.SetKeepAlivesEnabled(false)
		if err := server.Shutdown(ctx); err != nil {
			logger.Error("Could not gracefully shut down the server: " + err.Error())
		}
	}()

	if config.Tls.Enabled {
		if !helper.FileExists(config.Tls.CertFile) || !helper.FileExists(config.Tls.CertFile) {
			logger.Debug("TLS is enabled, but the certificate file or key file does not exist!")
			quit <- os.Interrupt
		} else {
			if err := server.ListenAndServeTLS(config.Tls.CertFile, config.Tls.KeyFile); err != nil && err != http.ErrServerClosed {
				logger.WithField("listenAddr", listenAddr).Error("Could not listen with TLS: " + err.Error())
				quit <- os.Interrupt
			}
		}
	} else {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithField("listenAddr", listenAddr).Error("Could not listen: " + err.Error())
			quit <- os.Interrupt
		}
	}

	logger.Trace("Server shutdown complete. Have a nice day!")
}

func setupRoutes(conf *entity.Configuration) *mux.Router {
	router := mux.NewRouter()
	router.Use(middleware.Limit, middleware.Headers)
	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = templateservice.ExecuteTemplate(w, "404.html", r.URL.Path)
	})

	httpHandler := handler.HttpHandler{
		Ds: databaseservice.Get(),
		SessMgr: global.GetSessionManager(),
		Logger: logging.GetCentralLogger(),
	}

	//asset file handlers
	router.HandleFunc("/assets/{file}", httpHandler.StaticAssetHandler)
	router.HandleFunc("/js/{file}", httpHandler.StaticAssetHandler)
	router.HandleFunc("/css/{file}", httpHandler.StaticAssetHandler)

	//site handlers
	router.HandleFunc("/login", httpHandler.LoginHandler).Methods(http.MethodGet, http.MethodPost)
	router.HandleFunc("/logout", httpHandler.LogoutHandler).Methods(http.MethodGet, http.MethodPost)
	router.HandleFunc("/password/request", httpHandler.RequestNewPasswordHandler).Methods(http.MethodGet, http.MethodPost)
	router.HandleFunc("/password/reset", httpHandler.ResetPasswordHandler).Methods(http.MethodGet, http.MethodPost)
	router.HandleFunc("/register", httpHandler.RegistrationHandler).Methods(http.MethodGet, http.MethodPost)
	router.HandleFunc("/register/confirm", httpHandler.RegistrationConfirmHandler).Methods(http.MethodGet, http.MethodPost)

	// Misc
	miscRouter := router.PathPrefix("").Subrouter()
	miscRouter.Use(middleware.Auth)
	miscRouter.HandleFunc("/", httpHandler.IndexHandler).Methods(http.MethodGet)

	userRouter := router.PathPrefix("/user").Subrouter()
	userRouter.Use(middleware.Auth)
	userRouter.HandleFunc("/settings", httpHandler.UserSettingsHandler).Methods(http.MethodGet, http.MethodPost)

	adminRouter := router.PathPrefix("/admin").Subrouter()
	adminRouter.Use(middleware.AuthWithAdmin)
	router.HandleFunc("/user/list", httpHandler.AdminUserListHandler).Methods(http.MethodGet)
	router.HandleFunc("/user/add", httpHandler.AdminUserAddHandler).Methods(http.MethodGet, http.MethodPost)
	router.HandleFunc("/user/{id}/edit", httpHandler.AdminUserEditHandler).Methods(http.MethodGet, http.MethodPost)
	router.HandleFunc("/user/{id}/remove", httpHandler.AdminUserRemoveHandler).Methods(http.MethodGet, http.MethodPost)
	router.HandleFunc("/settings", httpHandler.AdminSettingsHandler).Methods(http.MethodGet, http.MethodPost)

	bdRouter := router.PathPrefix("/builddefinition").Subrouter()
	bdRouter.Use(middleware.Auth)
	bdRouter.HandleFunc("/list", httpHandler.BuildDefinitionListHandler).Methods(http.MethodGet)
	bdRouter.HandleFunc("/add", httpHandler.BuildDefinitionAddHandler).Methods(http.MethodGet, http.MethodPost)
	bdRouter.HandleFunc("/{id}/show", httpHandler.BuildDefinitionShowHandler).Methods(http.MethodGet)
	bdRouter.HandleFunc("/{id}/edit", httpHandler.BuildDefinitionEditHandler).Methods(http.MethodGet, http.MethodPost)
	bdRouter.HandleFunc("/{id}/remove", httpHandler.BuildDefinitionRemoveHandler).Methods(http.MethodGet)
	bdRouter.HandleFunc("/{id}/listexecutions", httpHandler.BuildDefinitionListExecutionsHandler).Methods(http.MethodGet)
	bdRouter.HandleFunc("/{id}/restart", httpHandler.BuildDefinitionRestartHandler).Methods(http.MethodGet)
	bdRouter.HandleFunc("/{id}/artifact", httpHandler.DownloadNewestArtifactHandler).Methods(http.MethodGet)

	beRouter := router.PathPrefix("/buildexecution").Subrouter()
	beRouter.Use(middleware.Auth)
	beRouter.HandleFunc("/list", httpHandler.BuildExecutionListHandler).Methods(http.MethodGet)
	beRouter.HandleFunc("/{id}/show", httpHandler.BuildExecutionShowHandler).Methods(http.MethodGet)
	beRouter.HandleFunc("/{id}/artifact", httpHandler.DownloadSpecificArtifactHandler).Methods(http.MethodGet)

	varRouter := router.PathPrefix("/variable").Subrouter()
	varRouter.Use(middleware.Auth)
	varRouter.HandleFunc("/list", httpHandler.VariableListHandler).Methods(http.MethodGet)
	//varRouter.HandleFunc("/{id}/show", httpHandler.VariableShowHandler).Methods(http.MethodGet)
	varRouter.HandleFunc("/add", httpHandler.VariableAddHandler).Methods(http.MethodGet, http.MethodPost)
	varRouter.HandleFunc("/{id}/edit", httpHandler.VariableEditHandler).Methods(http.MethodGet, http.MethodPost)
	varRouter.HandleFunc("/{id}/remove", httpHandler.VariableRemoveHandler).Methods(http.MethodGet)

	// API handler
	router.HandleFunc("/api/v1/receive", httpHandler.PayloadReceiveHandler).Methods(http.MethodPost)

	return router
}
