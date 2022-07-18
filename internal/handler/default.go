package handler

import (
	"github.com/KaiserWerk/Tiny-Build-Server/internal/assets"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/templateservice"
	"github.com/gorilla/mux"
	"net/http"
	"strings"
)

// IndexHandler serves the dashboard
func (h *HttpHandler) IndexHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var (
		currentUser = r.Context().Value("user").(entity.User)
		logger      = h.ContextLogger("IndexHandler")
	)

	latestBuilds, err := h.DBService.GetNewestBuildExecutions(5, "")
	if err != nil {
		logger.WithField("error", err.Error()).Error("could not fetch latest build executions")
		http.Error(w, "could not fetch latest build definitions", http.StatusInternalServerError)
		return
	}
	latestBuildDefs, err := h.DBService.GetNewestBuildDefinitions(5)
	if err != nil {
		logger.WithField("error", err.Error()).Error("could not fetch latest build definitions")
		http.Error(w, "could not fetch latest build definitions", http.StatusInternalServerError)
		return
	}

	data := struct {
		CurrentUser     entity.User
		LatestBuilds    []entity.BuildExecution
		LatestBuildDefs []entity.BuildDefinition
	}{
		CurrentUser:     currentUser,
		LatestBuilds:    latestBuilds,
		LatestBuildDefs: latestBuildDefs,
	}

	if err := templateservice.ExecuteTemplate(h.Injector(), w, "index.html", data); err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
	}
}

// StaticAssetHandler serves static file. http.FileServer does not work as desired
func (h *HttpHandler) StaticAssetHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var (
		logger = h.ContextLogger("StaticAssetHandler")
		vars   = mux.Vars(r)
		file   = vars["file"]
		path   string
	)

	switch true {
	case strings.Contains(r.URL.Path, "assets/"):
		path = "assets"
	case strings.Contains(r.URL.Path, "js/"):
		path = "js"
	case strings.Contains(r.URL.Path, "css/"):
		path = "css"
	}

	data, err := assets.GetWebAssetFile(path + "/" + file)
	if err != nil {
		logger.WithField("file", file).Warn("could not locate asset file")
		w.WriteHeader(404)
		return
	}

	var ext string
	if strings.Contains(file, ".") {
		parts := strings.Split(file, ".")
		ext = parts[len(parts)-1]
	}

	var contentType string // = http.DetectContentType(data)
	switch ext {
	case "css":
		contentType = "text/css"
	case "js":
		contentType = "text/javascript"
	case "html":
		contentType = "text/html"
	case "jpg":
		fallthrough
	case "jpeg":
		contentType = "image/jpeg"
	case "gif":
		contentType = "image/gif"
	case "png":
		contentType = "image/png"
	case "svg":
		contentType = "image/svg+xml"
	default:
		contentType = "text/plain"
	}
	w.Header().Set("Content-Type", contentType)

	_, _ = w.Write(data)
}
