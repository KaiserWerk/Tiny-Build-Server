package handler

import (
	"github.com/KaiserWerk/Tiny-Build-Server/internal/buildservice"
	"net/http"
)

func PayloadReceiveHandler(w http.ResponseWriter, r *http.Request) {
	bd, err := buildservice.CheckPayloadRequest(r)
	if err != nil {
		http.Error(w, `{"status": "error", "message": "`+err.Error()+`"}`, 500)
		return
	}

	go buildservice.StartBuildProcess(bd)
}
