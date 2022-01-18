package main

import (
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/buildservice"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KaiserWerk/Tiny-Build-Server/internal/configuration"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/databaseservice"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/handler"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/logging"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/middleware"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/panicHandler"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/sessionservice"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/templateservice"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/stvp/slug"
)

var (
	Version    string = "DEV"
	listenPort string
	configFile string
	logPath    string
)

func main() {
	slug.Replacement = '-'

	flag.StringVar(&listenPort, "port", "8271", "The port which the build server should listen on")
	flag.StringVar(&configFile, "config", "app.yaml", "The location of the configuration file")
	flag.StringVar(&logPath, "logpath", ".", "The path to place log files in")
	flag.Parse()

	logger, cleanup, err := logging.New(logrus.DebugLevel, ".", "main", logging.ModeConsole|logging.ModeFile)
	if err != nil {
		panic("could not create new logger: " + err.Error())
	}
	defer func() {
		if err := cleanup(); err != nil {
			panic("could not execute logger cleanup func: " + err.Error())
		}
	}()

	defer panicHandler.Handle(logger)

	config, created, err := configuration.Setup(configFile)
	if err != nil {
		logger.WithField("error", err.Error()).Error("An error occurred while setting up configuration")
		return
	}
	if created {
		logger.Info("Configuration file didn't exist so it was created")
	}

	logger.WithFields(logrus.Fields{
		"app":        "Tiny Build Server",
		"version":    Version,
		"port":       listenPort,
		"configFile": configFile,
	}).Info("app information")

	ds := databaseservice.New(config)
	if err := ds.AutoMigrate(); err != nil {
		logger.WithField("error", err.Error()).Error("AutoMigrate panic")
		return
	}

	listenAddr := fmt.Sprintf(":%s", listenPort)
	logger.Trace("Server starts handling requests")

	var tlsEnabled bool
	if config.Tls.CertFile != "" && config.Tls.KeyFile != "" {
		logger.Debug("  TLS is enabled")
		tlsEnabled = true
	}

	router, err := setupRoutes(config, ds, logger)
	if err != nil {
		logger.WithField("error", err.Error()).Error("could not set up routes")
		return
	}

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
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      20 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
	}

	if tlsEnabled {
		server.TLSConfig = tlsConfig
		server.TLSNextProto = make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Kill, syscall.SIGTERM)

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

	if tlsEnabled {
		if err := server.ListenAndServeTLS(config.Tls.CertFile, config.Tls.KeyFile); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.WithField("listenAddr", listenAddr).Error("Could not run HTTPS server: " + err.Error())
			quit <- os.Interrupt
		}
	} else {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.WithField("listenAddr", listenAddr).Error("Could not run HTTP server: " + err.Error())
			quit <- os.Interrupt
		}
	}

	logger.Trace("Server shutdown complete. Have a nice day!")
}

func setupRoutes(cfg *entity.Configuration, ds *databaseservice.DatabaseService, l *logrus.Entry) (*mux.Router, error) {
	sessMgr := sessionservice.NewSessionManager("tbs_sessid")
	bs, err := buildservice.New("data", sessMgr, l, ds)
	if err != nil {
		return nil, err
	}

	mwHandler := middleware.MWHandler{
		Cfg:     cfg,
		Ds:      ds,
		SessMgr: sessMgr,
		Logger:  l,
	}

	inj := templateservice.Injector{
		Logger:  l,
		SessMgr: sessMgr,
		Ds:      ds,
	}

	router := mux.NewRouter()
	router.Use(mwHandler.Recover, mwHandler.Limit, mwHandler.Headers)
	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = templateservice.ExecuteTemplate(&inj, w, "404.html", r.URL.Path)
	})

	httpHandler := handler.HttpHandler{
		Cfg:     cfg,
		Ds:      ds,
		Bs:      bs,
		SessMgr: sessMgr,
		Logger:  l,
	}

	//asset file handlers
	router.HandleFunc("/assets/{file}", httpHandler.StaticAssetHandler)
	router.HandleFunc("/js/{file}", httpHandler.StaticAssetHandler)
	router.HandleFunc("/css/{file}", httpHandler.StaticAssetHandler)

	//fs := http.FileServer(http.FS(assets.GetWebAssetFS()))
	//router.Handle("/assets", fs)

	//site handlers
	router.HandleFunc("/login", httpHandler.LoginHandler).Methods(http.MethodGet, http.MethodPost)
	router.HandleFunc("/logout", httpHandler.LogoutHandler).Methods(http.MethodGet, http.MethodPost)
	router.HandleFunc("/password/request", httpHandler.RequestNewPasswordHandler).Methods(http.MethodGet, http.MethodPost)
	router.HandleFunc("/password/reset", httpHandler.ResetPasswordHandler).Methods(http.MethodGet, http.MethodPost)
	router.HandleFunc("/register", httpHandler.RegistrationHandler).Methods(http.MethodGet, http.MethodPost)
	router.HandleFunc("/register/confirm", httpHandler.RegistrationConfirmHandler).Methods(http.MethodGet, http.MethodPost)

	// Misc
	miscRouter := router.PathPrefix("").Subrouter()
	miscRouter.Use(mwHandler.Auth)
	miscRouter.HandleFunc("/", httpHandler.IndexHandler).Methods(http.MethodGet)

	// user routes
	userRouter := router.PathPrefix("/user").Subrouter()
	userRouter.Use(mwHandler.Auth)
	userRouter.HandleFunc("/settings", httpHandler.UserSettingsHandler).Methods(http.MethodGet, http.MethodPost)

	// admin routes
	adminRouter := router.PathPrefix("/admin").Subrouter()
	adminRouter.Use(mwHandler.AuthWithAdmin)
	adminRouter.HandleFunc("/user/list", httpHandler.AdminUserListHandler).Methods(http.MethodGet)
	adminRouter.HandleFunc("/user/add", httpHandler.AdminUserAddHandler).Methods(http.MethodGet, http.MethodPost)
	adminRouter.HandleFunc("/user/{id}/edit", httpHandler.AdminUserEditHandler).Methods(http.MethodGet, http.MethodPost)
	adminRouter.HandleFunc("/user/{id}/remove", httpHandler.AdminUserRemoveHandler).Methods(http.MethodGet, http.MethodPost)
	adminRouter.HandleFunc("/settings", httpHandler.AdminSettingsHandler).Methods(http.MethodGet, http.MethodPost)

	// build definition
	bdRouter := router.PathPrefix("/builddefinition").Subrouter()
	bdRouter.Use(mwHandler.Auth)
	bdRouter.HandleFunc("/list", httpHandler.BuildDefinitionListHandler).Methods(http.MethodGet)
	bdRouter.HandleFunc("/add", httpHandler.BuildDefinitionAddHandler).Methods(http.MethodGet, http.MethodPost)
	bdRouter.HandleFunc("/{id}/show", httpHandler.BuildDefinitionShowHandler).Methods(http.MethodGet)
	bdRouter.HandleFunc("/{id}/edit", httpHandler.BuildDefinitionEditHandler).Methods(http.MethodGet, http.MethodPost)
	bdRouter.HandleFunc("/{id}/remove", httpHandler.BuildDefinitionRemoveHandler).Methods(http.MethodGet)
	//bdRouter.HandleFunc("/{id}/listexecutions", httpHandler.BuildDefinitionListExecutionsHandler).Methods(http.MethodGet)
	bdRouter.HandleFunc("/{id}/restart", httpHandler.BuildDefinitionRestartHandler).Methods(http.MethodGet)
	bdRouter.HandleFunc("/{id}/artifact", httpHandler.DownloadNewestArtifactHandler).Methods(http.MethodGet)

	// build execution
	beRouter := router.PathPrefix("/buildexecution").Subrouter()
	beRouter.Use(mwHandler.Auth)
	beRouter.HandleFunc("/list", httpHandler.BuildExecutionListHandler).Methods(http.MethodGet)
	beRouter.HandleFunc("/{id}/show", httpHandler.BuildExecutionShowHandler).Methods(http.MethodGet)
	beRouter.HandleFunc("/{id}/artifact", httpHandler.DownloadSpecificArtifactHandler).Methods(http.MethodGet)

	// variables
	varRouter := router.PathPrefix("/variable").Subrouter()
	varRouter.Use(mwHandler.Auth)
	varRouter.HandleFunc("/list", httpHandler.VariableListHandler).Methods(http.MethodGet)
	varRouter.HandleFunc("/add", httpHandler.VariableAddHandler).Methods(http.MethodGet, http.MethodPost)
	varRouter.HandleFunc("/{id}/edit", httpHandler.VariableEditHandler).Methods(http.MethodGet, http.MethodPost)
	varRouter.HandleFunc("/{id}/remove", httpHandler.VariableRemoveHandler).Methods(http.MethodGet)

	// API handler
	router.HandleFunc("/api/v1/receive", httpHandler.PayloadReceiveHandler).Methods(http.MethodPost)

	return router, nil
}
