package handler

import (
	"github.com/KaiserWerk/Tiny-Build-Server/internal/buildservice"
	"net/http"
)

func PayloadReceiveHandler(w http.ResponseWriter, r *http.Request) {
	bd, err := buildservice.CheckPayloadRequest(r)
	if err != nil {
		http.Error(w, ``, http.StatusBadRequest)
		return
	}
	go buildservice.StartBuildProcess(bd)
}
