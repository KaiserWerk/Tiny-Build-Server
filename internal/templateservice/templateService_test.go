package templateservice

import (
	"github.com/KaiserWerk/Tiny-Build-Server/internal/fixtures"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/global"
	"github.com/KaiserWerk/sessionstore"
	"html/template"
	"net/http"
	"reflect"
	"testing"
)

func TestExecuteTemplate(t *testing.T) {
	type args struct {
		w    http.ResponseWriter
		file string
		data interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test.html cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ExecuteTemplate(tt.args.w, tt.args.file, tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("ExecuteTemplate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseEmailTemplate(t *testing.T) {
	type args struct {
		messageType string
		data        interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{name: "Test ParseEmailTemplate", args: args{
			messageType: string(fixtures.Test),
			data:        "World",
		}, want: "<p>Hello World!</p>", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseEmailTemplate(tt.args.messageType, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseEmailTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseEmailTemplate() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetFlashbag(t *testing.T) {
	type args struct {
		mgr *sessionstore.SessionManager
	}
	tests := []struct {
		name string
		args args
		want func() template.HTML
	}{
		{name: "Test GetFlashbag()", args: args{mgr: global.GetSessionManager()}, want: func() template.HTML { return "" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetFlashbag(tt.args.mgr); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetFlashbag() got func, want func. hah")
			}
		})
	}
}

func TestGetUsernameById(t *testing.T) {
	type args struct {
		id int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "Test GetUsernameById()", args: args{id: 1}, want: "Admin"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetUsernameById(tt.args.id); got != tt.want {
				t.Errorf("GetUsernameById() = %v, want %v", got, tt.want)
			}
		})
	}
}
