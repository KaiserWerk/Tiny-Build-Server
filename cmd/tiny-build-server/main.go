package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/handler"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/helper"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/middleware"
	"github.com/gorilla/mux"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	version = "0.0.1"
)

var (
	versionDate = "0000-00-00 00:00:00 +00:00" // inject at compile time
	listenPort  string
	configFile  string
)

func main() {
	helper.WriteToConsole("Tiny Build Server")
	helper.WriteToConsole("  Version: " + version)
	helper.WriteToConsole("  From: " + versionDate)
	flag.StringVar(&listenPort, "port", "8271", "The port which the build server should listen on")
	flag.StringVar(&configFile, "config", "", "The location of the configuration file")
	flag.Parse()

	if configFile != "" {
		helper.SetConfigurationFile(configFile)
	}

	config := helper.GetConfiguration()

	listenAddr := fmt.Sprintf(":%s", listenPort)
	helper.WriteToConsole("  Server will be handling requests at port " + listenPort)
	if config.Tls.Enabled {
		helper.WriteToConsole("  TLS is enabled")
	}

	router := mux.NewRouter()
	router.Use(middleware.Limit)
	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		if err := helper.ExecuteTemplate(w, "404.html", r.URL.Path); err != nil {
			w.WriteHeader(404)
		}
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

	//done := make(chan bool)
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, os.Kill, syscall.SIGTERM)

	go helper.ReadConsoleInput(quit)

	go func() {
		<-quit
		helper.WriteToConsole("Server is shutting down...")
		helper.Cleanup()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		server.SetKeepAlivesEnabled(false)
		if err := server.Shutdown(ctx); err != nil {
			helper.WriteToConsole("Could not gracefully shut down the server: " + err.Error())
			quit <- os.Interrupt
		}
		//close(done)
	}()

	if config.Tls.Enabled {
		if !helper.FileExists(config.Tls.CertFile) || !helper.FileExists(config.Tls.CertFile) {
			helper.WriteToConsole("TLS is enabled, but the certificate file or key file does not exist!")
			quit <- os.Interrupt
		} else {
			if err := server.ListenAndServeTLS(config.Tls.CertFile, config.Tls.KeyFile); err != nil && err != http.ErrServerClosed {
				helper.WriteToConsole("Could not listen with TLS on " + listenAddr + ": " + err.Error())
				quit <- os.Interrupt
			}
		}
	} else {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			helper.WriteToConsole("Could not listen on " + listenAddr + ": " + err.Error())
			quit <- os.Interrupt
		}
	}
	//<-done
	helper.WriteToConsole("Server shutdown complete. Have a nice day!")
}

func setupRoutes(router *mux.Router) {
	// asset file handlers
	router.HandleFunc("/assets/{file}", handler.StaticAssetHandler)
	router.HandleFunc("/js/{file}", handler.StaticAssetHandler)
	router.HandleFunc("/css/{file}", handler.StaticAssetHandler)

	// site handlers
	router.HandleFunc("/", handler.IndexHandler).Methods("GET")
	router.HandleFunc("/login", handler.LoginHandler).Methods("GET", "POST")
	router.HandleFunc("/logout", handler.LogoutHandler).Methods("GET", "POST")
	router.HandleFunc("/password/request", handler.RequestNewPasswordHandler).Methods("GET", "POST")
	router.HandleFunc("/password/reset", handler.ResetPasswordHandler).Methods("GET", "POST")
	router.HandleFunc("/register", handler.RegistrationHandler).Methods("GET", "POST")

	router.HandleFunc("/admin/user/list", handler.AdminUserListHandler).Methods("GET")
	router.HandleFunc("/admin/user/add", handler.AdminUserAddHandler).Methods("GET", "POST")
	router.HandleFunc("/admin/user/{id}/edit", handler.AdminUserEditHandler).Methods("GET", "POST")
	router.HandleFunc("/admin/settings", handler.AdminSettingsHandler).Methods("GET", "POST")

	router.HandleFunc("/builddefinition/list", handler.BuildDefinitionListHandler).Methods("GET")
	router.HandleFunc("/builddefinition/add", handler.BuildDefinitionAddHandler).Methods("GET", "POST")
	router.HandleFunc("/builddefinition/{id}/show", handler.BuildDefinitionShowHandler).Methods("GET")
	router.HandleFunc("/builddefinition/{id}/edit", handler.BuildDefinitionEditHandler).Methods("GET", "POST")
	router.HandleFunc("/builddefinition/{id}/remove", handler.BuildDefinitionRemoveHandler).Methods("GET")
	router.HandleFunc("/builddefinition/{id}/listexecutions", handler.BuildDefinitionListExecutionsHandler).Methods("GET")
	router.HandleFunc("/builddefinition/{id}/restart", handler.BuildDefinitionRestartHandler).Methods("GET")                                 // TODO: implement handler
	router.HandleFunc("/builddefinition/{id}/deploymentdefinitions/list", handler.DeploymentDefinitionListHandler).Methods("GET")            // TODO: implement handler
	router.HandleFunc("/builddefinition/{id}/deploymentdefinitions/add", handler.DeploymentDefinitionAddHandler).Methods("GET")              // TODO: implement handler
	router.HandleFunc("/builddefinition/{id}/deploymentdefinitions/{ddid}/edit", handler.DeploymentDefinitionEditHandler).Methods("GET")     // TODO: implement handler
	router.HandleFunc("/builddefinition/{id}/deploymentdefinitions/{ddid}/remove", handler.DeploymentDefinitionRemoveHandler).Methods("GET") // TODO: implement handler

	router.HandleFunc("/buildexecution/list", handler.BuildExecutionListHandler).Methods("GET")
	router.HandleFunc("/buildexecution/{id}/show", handler.BuildExecutionShowHandler).Methods("GET")

	// API handlers
	router.HandleFunc("/api/v1/receive", handler.PayloadReceiveHandler).Methods(http.MethodPost)
	// TODO: JSON -> datasweet/jsonmap?
}
