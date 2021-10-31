package handler

import (
	"github.com/KaiserWerk/Tiny-Build-Server/internal/databaseservice"
)

type HttpHandler struct {
	Ds *databaseservice.DatabaseService
}
