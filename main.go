package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

var (
	templates map[string]*template.Template
	publicDir = "/public/"
)

func main() {
	templates = populateTemplates()

	listenAddrPtr := flag.String("port", "5000", "The port which the build server should listen on")
	flag.Parse() // <.<
	listenAddr := fmt.Sprintf(":%s", *listenAddrPtr)
	writeToConsole("Server is ready to handle requests at port " + *listenAddrPtr, )

	if _, err := loadSysConfig(); err != nil {
		log.Fatal("could not handle config/app.yaml file; something went wrong")
	}

	router := mux.NewRouter()
	router.Use(limit)

	// asset file handlers
	router.PathPrefix("/css/").Handler(http.FileServer(http.Dir("public"))).Methods("GET")
	router.PathPrefix("/js/").Handler(http.FileServer(http.Dir("public"))).Methods("GET")
	router.PathPrefix("/assets/").Handler(http.FileServer(http.Dir("public"))).Methods("GET")

	// site handlers
	router.HandleFunc("/", indexHandler).Methods("GET")
	router.HandleFunc("/login", loginHandler).Methods("GET", "POST")

	// API handlers
	router.HandleFunc("/bitbucket-receive", bitBucketReceiveHandler).Methods("POST")
	router.HandleFunc("/github-receive", gitHubReceiveHandler).Methods("POST")
	router.HandleFunc("/gitlab-receive", gitLabReceiveHandler).Methods("POST")
	router.HandleFunc("/gitea-receive", giteaReceiveHandler).Methods("POST")
	router.HandleFunc("/ping", pingHandler)


	server := &http.Server{
		Addr:         listenAddr,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	externalShutdownCh := make(chan bool)
	done := make(chan bool)
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)

	go readConsoleInput(externalShutdownCh)

	go func() {
		<-quit
		writeToConsole("Server is shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		server.SetKeepAlivesEnabled(false)
		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("Could not gracefully shutdown the server: %v\n", err)
		}
		close(done)
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not listen on %s: %v\n", listenAddr, err)
	}
	<-done
	writeToConsole("Server shutdown complete. Have a nice day!")
}

func populateTemplates() map[string]*template.Template {
	result := make(map[string]*template.Template)
	const basePath = "templates"
	layout := template.Must(template.ParseFiles(basePath + "/_layout.html"))
	//template.Must(layout.ParseFiles(...
	dir, err := os.Open(basePath + "/content")
	if err != nil {
		panic("Failed to open template block directory: " + err.Error())
	}
	fis, err := dir.Readdir(-1)
	if err != nil {
		panic("Failed to contents of content directory: " + err.Error())
	}
	for _, fi := range fis {
		f, err := os.Open(basePath + "/content/" + fi.Name())
		if err != nil {
			panic("Failed to open template '"+fi.Name()+"': " + err.Error())
		}
		content, err := ioutil.ReadAll(f)
		if err != nil {
			panic("Failed to read content from file '"+fi.Name()+"': " + err.Error())
		}
		f.Close()
		tmpl := template.Must(layout.Clone())
		_, err = tmpl.Parse(string(content))
		if err != nil {
			panic("Failed to parse contents of file '"+fi.Name()+"': " + err.Error())
		}
		result[fi.Name()] = tmpl
	}
	return result
}