package sessionservice

import (
	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"github.com/KaiserWerk/sessionstore"
	"reflect"
	"testing"
)

func TestGetUserFromSession(t *testing.T) {
	type args struct {
		s sessionstore.Session
	}
	tests := []struct {
		name    string
		args    args
		want    entity.User
		wantErr bool
	}{
		{name: "Test GetUserFromSession()", args: args{s: sessionstore.Session{}}, want: entity.User{}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetUserFromSession(tt.args.s)
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
