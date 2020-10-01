package main

import (
	"fmt"
	"net/http"
)

func buildExecutionListHandler(w http.ResponseWriter, r *http.Request) {

	t := templates["buildexecution_list.html"]
	if t != nil {
		err := t.Execute(w, nil)
		if err != nil {
			fmt.Println("error:", err.Error())
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func buildExecutionShowHandler(w http.ResponseWriter, r *http.Request) {

	t := templates["buildexecution_show.html"]
	if t != nil {
		err := t.Execute(w, nil)
		if err != nil {
			fmt.Println("error:", err.Error())
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}
