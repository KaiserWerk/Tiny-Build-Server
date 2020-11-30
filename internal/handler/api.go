package handler

import (
	"github.com/KaiserWerk/Tiny-Build-Server/internal/helper"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/service"
	"net/http"
)

func PayloadReceiveHandler(w http.ResponseWriter, r *http.Request) {
	bd, err := helper.CheckPayloadRequest(r)
	if err != nil {
		http.Error(w, `{"status": "error", "message": "`+err.Error()+`"}`, 500)
		return
	}

	go service.StartBuildProcess(bd)
}
