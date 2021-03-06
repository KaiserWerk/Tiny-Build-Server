package templateservice

import (
	"github.com/KaiserWerk/Tiny-Build-Server/internal"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/databaseservice"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/global"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/helper"
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
	ds := databaseservice.New()
	var funcMap = template.FuncMap{
		"getUsernameById":    GetUsernameById,
		"getFlashbag":        GetFlashbag(global.GetSessionManager()),
		"formatDate":         helper.FormatDate,
		"getBuildDefCaption": ds.GetBuildDefCaption,
	}
	layoutContent, err := internal.FSString(false, "/templates/_layout.html") // with leading slash?
	if err != nil {
		helper.WriteToConsole("could not get layout template: " + err.Error())
		return err
	}

	layout := template.Must(template.New("_layout.html").Parse(layoutContent)).Funcs(funcMap)

	content, err := internal.FSString(false, "/templates/content/"+file) // with leading slash?
	if err != nil {
		helper.WriteToConsole("could not find template " + file + ": " + err.Error())
		return err
	}

	tmpl := template.Must(layout.Clone())
	_, err = tmpl.Parse(content)
	if err != nil {
		helper.WriteToConsole("could not parse template into base layout: " + err.Error())
		return err
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		helper.WriteToConsole("could not execute template " + file + ": " + err.Error())
		return err
	}

	return nil
}

// ParseEmailTemplate parses and email template with the given data
func ParseEmailTemplate(messageType string, data interface{}) (string, error) {
	cont, err := internal.FSString(false, "/templates/email/"+messageType+".html")
	if err != nil {
		helper.WriteToConsole("could not get FSString email template: " + err.Error())
		return "", err
	}
	t, err := template.New(messageType).Parse(string(cont))
	if err != nil {
		helper.WriteToConsole("could not parse email template")
		return "", err
	}

	b := bytes.NewBufferString("")
	err = t.Execute(b, data)
	if err != nil {
		helper.WriteToConsole("could not execute email template")
	}

	return b.String(), nil
}

// GetFlashbag return a HTML string populated with flash messages, if available
func GetFlashbag(mgr *sessionstore.SessionManager) func() template.HTML {
	return func() template.HTML {
		if mgr == nil {
			helper.WriteToConsole("sessionManager is nil in getFlashbag")
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
	ds := databaseservice.New()
	//defer ds.Quit()

	u, err := ds.GetUserById(id)
	if err != nil {
		return "--"
	}

	return u.Displayname
}
