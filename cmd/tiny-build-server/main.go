package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/databaseservice"
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

	ds := databaseservice.New()
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

	router := mux.NewRouter()
	router.Use(middleware.Limit, middleware.Headers)
	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = templateservice.ExecuteTemplate(w, "404.html", r.URL.Path)
	})

	setupRoutes(router)

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

func setupRoutes(router *mux.Router) {
	//asset file handlers
	router.HandleFunc("/assets/{file}", handler.StaticAssetHandler)
	router.HandleFunc("/js/{file}", handler.StaticAssetHandler)
	router.HandleFunc("/css/{file}", handler.StaticAssetHandler)

	//site handlers
	router.HandleFunc("/", handler.IndexHandler).Methods(http.MethodGet)
	router.HandleFunc("/login", handler.LoginHandler).Methods(http.MethodGet, http.MethodPost)
	router.HandleFunc("/logout", handler.LogoutHandler).Methods(http.MethodGet, http.MethodPost)
	router.HandleFunc("/password/request", handler.RequestNewPasswordHandler).Methods(http.MethodGet, http.MethodPost)
	router.HandleFunc("/password/reset", handler.ResetPasswordHandler).Methods(http.MethodGet, http.MethodPost)
	router.HandleFunc("/register", handler.RegistrationHandler).Methods(http.MethodGet, http.MethodPost)
	router.HandleFunc("/register/confirm", handler.RegistrationConfirmHandler).Methods(http.MethodGet, http.MethodPost)

	router.HandleFunc("/user/settings", handler.UserSettingsHandler).Methods(http.MethodGet, http.MethodPost)

	router.HandleFunc("/admin/user/list", handler.AdminUserListHandler).Methods(http.MethodGet)
	router.HandleFunc("/admin/user/add", handler.AdminUserAddHandler).Methods(http.MethodGet, http.MethodPost)
	router.HandleFunc("/admin/user/{id}/edit", handler.AdminUserEditHandler).Methods(http.MethodGet, http.MethodPost)
	router.HandleFunc("/admin/user/{id}/remove", handler.AdminUserRemoveHandler).Methods(http.MethodGet, http.MethodPost)
	router.HandleFunc("/admin/settings", handler.AdminSettingsHandler).Methods(http.MethodGet, http.MethodPost)

	router.HandleFunc("/builddefinition/list", handler.BuildDefinitionListHandler).Methods(http.MethodGet)
	router.HandleFunc("/builddefinition/add", handler.BuildDefinitionAddHandler).Methods(http.MethodGet, http.MethodPost)
	router.HandleFunc("/builddefinition/{id}/show", handler.BuildDefinitionShowHandler).Methods(http.MethodGet)
	router.HandleFunc("/builddefinition/{id}/edit", handler.BuildDefinitionEditHandler).Methods(http.MethodGet, http.MethodPost)
	router.HandleFunc("/builddefinition/{id}/remove", handler.BuildDefinitionRemoveHandler).Methods(http.MethodGet)
	router.HandleFunc("/builddefinition/{id}/listexecutions", handler.BuildDefinitionListExecutionsHandler).Methods(http.MethodGet)
	router.HandleFunc("/builddefinition/{id}/restart", handler.BuildDefinitionRestartHandler).Methods(http.MethodGet)
	router.HandleFunc("/builddefinition/{id}/artifact", handler.DownloadNewestArtifactHandler).Methods(http.MethodGet)

	router.HandleFunc("/buildexecution/list", handler.BuildExecutionListHandler).Methods(http.MethodGet)
	router.HandleFunc("/buildexecution/{id}/show", handler.BuildExecutionShowHandler).Methods(http.MethodGet)
	router.HandleFunc("/buildexecution/{id}/artifact", handler.DownloadSpecificArtifactHandler).Methods(http.MethodGet)

	router.HandleFunc("/variable/list", handler.VariableListHandler).Methods(http.MethodGet)
	router.HandleFunc("/variable/{id}/show", handler.VariableShowHandler).Methods(http.MethodGet)
	router.HandleFunc("/variable/add", handler.VariableAddHandler).Methods(http.MethodGet, http.MethodPost)
	router.HandleFunc("/variable/{id}/edit", handler.VariableEditHandler).Methods(http.MethodGet, http.MethodPost)
	router.HandleFunc("/variable/{id}/remove", handler.VariableRemoveHandler).Methods(http.MethodGet)

	// API handler
	router.HandleFunc("/api/v1/receive", handler.PayloadReceiveHandler).Methods(http.MethodPost)
}
