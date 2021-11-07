package templateservice

import (
	"github.com/KaiserWerk/Tiny-Build-Server/internal/assets"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/databaseservice"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/global"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/helper"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/logging"
	"github.com/sirupsen/logrus"
	"io"
)

import (
	"bytes"
	"html/template"
	"strings"

	"github.com/KaiserWerk/sessionstore"
)

// ExecuteTemplate executed a template with the supplied data into the io.Writer w
func ExecuteTemplate(w io.Writer, file string, data interface{}) error {
	var (
		ds = databaseservice.Get()
		funcMap = template.FuncMap{
			"getUsernameById":    GetUsernameById,
			"getFlashbag":        GetFlashbag(global.GetSessionManager()),
			"formatDate":         helper.FormatDate,
			"getBuildDefCaption": ds.GetBuildDefCaption,
		}
		logger = logging.New(logrus.InfoLevel, "ExecuteTemplate", true)
	)
	layoutContent, err := assets.GetTemplate("_layout.html")
	if err != nil {
		logger.WithField("error", err.Error()).Error("could not get layout template")
		return err
	}

	layout := template.Must(template.New("_layout.html").Parse(string(layoutContent))).Funcs(funcMap)

	content, err := assets.GetTemplate("content/"+file)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err.Error(),
			"file": file,
		}).Error("could not find template")
		return err
	}

	tmpl := template.Must(layout.Clone())
	_, err = tmpl.Parse(string(content))
	if err != nil {
		logger.WithField("error", err.Error()).Error("could not parse template into base layout")
		return err
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err.Error(),
			"file": file,
		}).Error("could not execute template")
		return err
	}

	return nil
}

// ParseEmailTemplate parses and email template with the given data
func ParseEmailTemplate(messageType string, data interface{}) (string, error) {
	logger := logging.New(logrus.InfoLevel, "ParseEmailTemplate", true)
	cont, err := assets.GetTemplate("email/"+messageType+".html")
	if err != nil {
		logger.WithField("error", err.Error()).Error("could not get email template")
		return "", err
	}
	t, err := template.New(messageType).Parse(string(cont))
	if err != nil {
		logger.WithField("error", err.Error()).Error("could not parse email template")
		return "", err
	}

	b := new(bytes.Buffer)
	err = t.Execute(b, data)
	if err != nil {
		logger.WithField("error", err.Error()).Error("could not execute email template")
		return "", err
	}

	return b.String(), nil
}

// GetFlashbag return a HTML string populated with flash messages, if available
func GetFlashbag(mgr *sessionstore.SessionManager) func() template.HTML {
	return func() template.HTML {
		logger := logging.New(logrus.InfoLevel, "GetFlashbag", true)
		if mgr == nil {
			logger.Error("sessionManager is nil")
			return template.HTML("")
		}
		var sb strings.Builder
		var source string
		const msgSuccess = `<div class="alert alert-success alert-dismissable"><a href="#" class="close" data-dismiss="alert" aria-label="close">&times;</a><strong>Success!</strong> %%message%%</div>`
		const msgError = `<div class="alert alert-danger alert-dismissable"><a href="#" class="close" data-dismiss="alert" aria-label="close">&times;</a><strong>Error!</strong> %%message%%</div>`
		const msgWarning = `<div class="alert alert-warning alert-dismissable"><a href="#" class="close" data-dismiss="alert" aria-label="close">&times;</a><strong>Warning!</strong> %%message%%</div>`
		const msgInfo = `<div class="alert alert-info alert-dismissable"><a href="#" class="close" data-dismiss="alert" aria-label="close">&times;</a><strong>Info!</strong> %%message%%</div>`

		for _, v := range mgr.GetMessages() {
			if v.MessageType == "success" {
				source = msgSuccess
			} else if v.MessageType == "error" {
				source = msgError
			} else if v.MessageType == "warning" {
				source = msgWarning
			} else if v.MessageType == "info" {
				source = msgInfo
			}

			sb.WriteString(strings.Replace(source, "%%message%%", v.Content, 1))
		}

		return template.HTML(sb.String())
	}
}

// GetUsernameById returns a username by id
func GetUsernameById(id int) string {
	ds := databaseservice.Get()
	//defer ds.Quit()

	u, err := ds.GetUserById(id)
	if err != nil {
		return "--"
	}

	return u.Displayname
}
