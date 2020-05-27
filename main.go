package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	//dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	//if err != nil {
	//	log.Fatal(err)
	//}
	//fmt.Println(dir)

	listenAddrPtr := flag.String("port", "5000", "The port which the build server should listen on")
	flag.Parse() // <.<
	listenAddr := fmt.Sprintf(":%s", *listenAddrPtr)
	fmt.Printf("Server is ready to handle requests at port %s\n", *listenAddrPtr)

	if _, err := loadSysConfig(); err != nil {
		log.Fatal("could not handle config/app.yaml file; something went wrong")
	}

	router := mux.NewRouter()
	router.Use(limit)
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
		goOn := false
		for {
			select {
			case <-quit:
				goOn = true
			case <-externalShutdownCh:
				goOn = true
			}
			if goOn {
				break
			}
		}

		fmt.Println("\nServer is shutting down...")

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
	fmt.Println("Server shutdown complete. Have a nice day!")
}
