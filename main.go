package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/KaiserWerk/sessionstore"
	"github.com/gorilla/mux"
	"log"
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
	versionDate   = "0000-00-00 00:00:00 +00:00" // inject at compile time
	listenPort    string
	configFile    string
	centralConfig configuration
	sessMgr       *sessionstore.SessionManager
)

func main() {
	writeToConsole("Tiny Build Server")
	writeToConsole("  Version: " + version)
	writeToConsole("  From: " + versionDate)
	flag.StringVar(&listenPort, "port", "8271", "The port which the build server should listen on")
	flag.StringVar(&configFile, "config", "config/app.yaml", "The location of the configuration file")
	flag.Parse()

	centralConfig = getConfiguration()
	sessMgr = sessionstore.NewManager("tbs_sessid")

	listenAddr := fmt.Sprintf(":%s", listenPort)
	writeToConsole("  Server will be handling requests at port " + listenPort)
	if centralConfig.Tls.Enabled {
		writeToConsole("  TLS is enabled")
	}

	router := mux.NewRouter()
	router.Use(limit)
	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		if err := executeTemplate(w, "404.html", r.URL.Path); err != nil {
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
		Addr:         listenAddr,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	if centralConfig.Tls.Enabled {
		server.TLSConfig = tlsConfig
		server.TLSNextProto = make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0)
	}

	done := make(chan bool)
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go readConsoleInput(quit)

	go func() {
		<-quit
		writeToConsole("Server is shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		server.SetKeepAlivesEnabled(false)
		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("Could not gracefully shut down the server: %v\n", err)
		}
		close(done)
	}()

	if centralConfig.Tls.Enabled {
		if !fileExists(centralConfig.Tls.CertFile) || !fileExists(centralConfig.Tls.CertFile) {
			writeToConsole("TLS is enabled, but the certificate file or key file does not exist!")
			quit <- os.Interrupt
		} else {
			if err := server.ListenAndServeTLS(centralConfig.Tls.CertFile, centralConfig.Tls.KeyFile); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Could not listen on %s: %v\n", listenAddr, err)
			}
		}
	} else {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on %s: %v\n", listenAddr, err)
		}
	}
	<-done
	writeToConsole("Server shutdown complete. Have a nice day!")
}

func setupRoutes(router *mux.Router) {
	// asset file handlers
	//router.PathPrefix("/css/").Handler(http.FileServer(http.Dir("public"))).Methods("GET")
	//router.PathPrefix("/js/").Handler(http.FileServer(http.Dir("public"))).Methods("GET")
	//router.PathPrefix("/assets/").Handler(http.FileServer(http.Dir("public"))).Methods("GET")
	router.HandleFunc("/assets/{file}", staticAssetHandler)
	router.HandleFunc("/js/{file}", staticAssetHandler)
	router.HandleFunc("/css/{file}", staticAssetHandler)

	// site handlers
	router.HandleFunc("/", indexHandler).Methods("GET")
	router.HandleFunc("/login", loginHandler).Methods("GET", "POST")
	router.HandleFunc("/logout", logoutHandler).Methods("GET", "POST")
	router.HandleFunc("/password/request", requestNewPasswordHandler).Methods("GET", "POST")
	router.HandleFunc("/password/reset", resetPasswordHandler).Methods("GET", "POST")
	router.HandleFunc("/register", registrationHandler).Methods("GET", "POST")

	router.HandleFunc("/admin/settings", adminSettingsHandler).Methods("GET", "POST")
	//router.HandleFunc("/admin/buildtarget/list", adminBuildTargetListHandler).Methods("GET")
	//router.HandleFunc("/admin/buildtarget/add", adminBuildTargetAddHandler).Methods("GET", "POST")
	//router.HandleFunc("/admin/buildtarget/{id}/edit", adminBuildTargetEditHandler).Methods("GET", "POST")
	//router.HandleFunc("/admin/buildtarget/{id}/remove", adminBuildTargetRemoveHandler).Methods("GET")
	//router.HandleFunc("/admin/buildstep/list", adminBuildStepListHandler).Methods("GET")
	//router.HandleFunc("/admin/buildstep/add", adminBuildStepAddHandler).Methods("GET", "POST")
	//router.HandleFunc("/admin/buildstep/{id}/edit", adminBuildStepEditHandler).Methods("GET", "POST")
	//router.HandleFunc("/admin/buildstep/{id}/remove", adminBuildStepRemoveHandler).Methods("GET")

	router.HandleFunc("/builddefinition/list", buildDefinitionListHandler).Methods("GET")
	router.HandleFunc("/builddefinition/add", buildDefinitionAddHandler).Methods("GET", "POST")
	router.HandleFunc("/builddefinition/{id}/show", buildDefinitionShowHandler).Methods("GET")
	router.HandleFunc("/builddefinition/{id}/edit", buildDefinitionEditHandler).Methods("GET", "POST")
	router.HandleFunc("/builddefinition/{id}/remove", buildDefinitionRemoveHandler).Methods("GET")
	router.HandleFunc("/builddefinition/{id}/listexecutions", buildDefinitionListExecutionsHandler).Methods("GET")
	router.HandleFunc("/builddefinition/{id}/restart", nil).Methods("GET") // TODO

	router.HandleFunc("/buildexecution/list", buildExecutionListHandler).Methods("GET")
	router.HandleFunc("/buildexecution/{id}/show", buildExecutionShowHandler).Methods("GET")

	// API handlers
	router.HandleFunc("/api/v1/receive", payloadReceiveHandler).Methods(http.MethodPost)
	//apiRouter := router.PathPrefix("/api/v1").Subrouter()
	//apiRouter.HandleFunc("/bitbucket", bitBucketReceiveHandler).Methods("POST")
	//apiRouter.HandleFunc("/github", gitHubReceiveHandler).Methods("POST")
	//apiRouter.HandleFunc("/gitlab", gitLabReceiveHandler).Methods("POST")
	//apiRouter.HandleFunc("/gitea", giteaReceiveHandler).Methods("POST")
	// ummodeln auf service-agnostischen handler.
	// anhand der build definition wird festgestellt, welcher dienst genutzt wird.
	// JSON -> datasweet/jsonmap?
}

