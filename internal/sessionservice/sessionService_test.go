package sessionservice

import (
	"reflect"
	"testing"

	"github.com/KaiserWerk/Tiny-Build-Server/internal/configuration"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/dbservice"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"

	"github.com/KaiserWerk/sessionstore/v2"
)

func TestGetUserFromSession(t *testing.T) {
	ds := dbservice.New(&configuration.AppConfig{})

	type args struct {
		s *sessionstore.Session
	}
	tests := []struct {
		name    string
		args    args
		want    entity.User
		wantErr bool
	}{
		{name: "Test GetUserFromSession()", args: args{s: &sessionstore.Session{}}, want: entity.User{}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetUserFromSession(ds, tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUserFromSession() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetUserFromSession() got = %v, want %v", got, tt.want)
			}
		})
	}
}
