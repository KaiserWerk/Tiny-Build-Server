package templateservice

import (
	"bytes"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/dbservice"
	"html/template"
	"io"
	"strings"

	"github.com/KaiserWerk/Tiny-Build-Server/internal/assets"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/helper"

	"github.com/KaiserWerk/sessionstore/v2"
	"github.com/sirupsen/logrus"
)

type Injector struct {
	Logger  *logrus.Entry
	SessMgr *sessionstore.SessionManager
	Ds      *dbservice.DBService
}

// ExecuteTemplate executed a template with the supplied data into the io.Writer w
func ExecuteTemplate(inj *Injector, w io.Writer, file string, data interface{}) error {
	var (
		funcMap = template.FuncMap{
			"getFlashbag": GetFlashbag(inj.Logger, inj.SessMgr),
			"formatDate":  helper.FormatDate,
			"getUsernameById": func(id uint) string {
				return GetUsernameById(inj.Ds, id)
			},
			"getBuildDefCaption": inj.Ds.GetBuildDefCaption,
		}
	)
	layoutContent, err := assets.GetTemplate("_layout.html")
	if err != nil {
		inj.Logger.WithField("error", err.Error()).Error("could not get layout template")
		return err
	}

	layout := template.Must(template.New("_layout.html").Parse(string(layoutContent))).Funcs(funcMap)

	content, err := assets.GetTemplate("content/" + file)
	if err != nil {
		inj.Logger.WithFields(logrus.Fields{
			"error": err.Error(),
			"file":  file,
		}).Error("could not find template")
		return err
	}

	tmpl := template.Must(layout.Clone())
	_, err = tmpl.Parse(string(content))
	if err != nil {
		inj.Logger.WithField("error", err.Error()).Error("could not parse template into base layout")
		return err
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		inj.Logger.WithFields(logrus.Fields{
			"error": err.Error(),
			"file":  file,
		}).Error("could not execute template")
		return err
	}

	return nil
}

// ParseEmailTemplate parses and email template with the given data
func ParseEmailTemplate(tmpl string, data interface{}) (string, error) {
	cont, err := assets.GetTemplate("email/" + tmpl + ".html")
	if err != nil {
		return "", err
	}
	t, err := template.New(tmpl).Parse(string(cont))
	if err != nil {
		return "", err
	}

	b := new(bytes.Buffer)
	err = t.Execute(b, data)
	if err != nil {
		return "", err
	}

	return b.String(), nil
}

// GetFlashbag return a HTML string populated with flash messages, if available
func GetFlashbag(logger *logrus.Entry, mgr *sessionstore.SessionManager) func() template.HTML {
	return func() template.HTML {
		if mgr == nil {
			logger.Error("sessionManager is nil")
			return template.HTML("")
		}
		var sb strings.Builder
		//var source string
		const msgSuccess = `<div class="alert alert-success alert-dismissable"><a href="#" class="close" data-dismiss="alert" aria-label="close">&times;</a><strong>Success!</strong> %%message%%</div>`
		const msgError = `<div class="alert alert-danger alert-dismissable"><a href="#" class="close" data-dismiss="alert" aria-label="close">&times;</a><strong>Error!</strong> %%message%%</div>`
		const msgWarning = `<div class="alert alert-warning alert-dismissable"><a href="#" class="close" data-dismiss="alert" aria-label="close">&times;</a><strong>Warning!</strong> %%message%%</div>`
		const msgInfo = `<div class="alert alert-info alert-dismissable"><a href="#" class="close" data-dismiss="alert" aria-label="close">&times;</a><strong>Info!</strong> %%message%%</div>`

		//for _, v := range mgr.GetMessages() {
		//	if v.MessageType == "success" {
		//		source = msgSuccess
		//	} else if v.MessageType == "error" {
		//		source = msgError
		//	} else if v.MessageType == "warning" {
		//		source = msgWarning
		//	} else if v.MessageType == "info" {
		//		source = msgInfo
		//	}
		//
		//	sb.WriteString(strings.Replace(source, "%%message%%", v.Raw, 1))
		//}

		return template.HTML(sb.String())
	}
}

// GetUsernameById returns a username by id
func GetUsernameById(ds *dbservice.DBService, id uint) string {
	u, err := ds.GetUserById(id)
	if err != nil {
		return "--"
	}

	return u.DisplayName
}
