package security

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/KaiserWerk/sessionstore/v2"
)

func TestGenerateToken(t *testing.T) {
	token := GenerateToken(40)

	if len(token) != 80 {
		t.Errorf("Token length not generated correctly; expected %d, got %d", 80, len(token))
	}
}

func TestHashString(t *testing.T) {
	type args struct {
		password string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{name: "Test HashString() + DoesHashMatch()", args: args{password: "test"}, want: "test", wantErr: false},
		{name: "Test HashString() + DoesHashMatch()", args: args{password: "r4gz1tw69s1t5g"}, want: "r4gz1tw69s1t5g", wantErr: false},
		{name: "Test HashString() + DoesHashMatch()", args: args{password: "w43ztg3et"}, want: "w43ztg3et", wantErr: false},
		{name: "Test HashString() + DoesHashMatch()", args: args{password: "MeinTollesPasswort123!"}, want: "MeinTollesPasswort123!", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HashString(tt.args.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("HashString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !DoesHashMatch(tt.want, got) {
				t.Errorf("HashString() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDoesHashMatch(t *testing.T) {
	type args struct {
		password string
		hash     string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "Test DoesHashMatch() 1", args: args{
			password: "hallo123",
			hash:     "$2a$12$3YUcQnQjsm2I.kKDlrsSkuovuQhgtSzqViOywSEwOeNjw8GwaoeQu",
		}, want: true},
		{name: "Test DoesHashMatch() 2", args: args{
			password: "r4gz1tw69s1t5g",
			hash:     "$2a$12$/6aKRIdoi6Ty5virvUmyBe0y6M2xk4n4AKOJlibCOQweXiCuJSoca",
		}, want: true},
		{name: "Test DoesHashMatch() 3", args: args{
			password: "MeinTollesPasswort123!",
			hash:     "$2a$12$8JIY1DGesX7WRpW/BNIuvedseG4lNIur1ILEJQhr4C99MAZOZTyqC",
		}, want: true},
		{name: "Test DoesHashMatch() 4", args: args{
			password: "test",
			hash:     "$2a$12$rjKereh7RdSKFNTjRjDJNedRFS/rv58L7GWT/32wk5fvEQp2WB17u",
		}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DoesHashMatch(tt.args.password, tt.args.hash); got != tt.want {
				t.Errorf("DoesHashMatch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckLogin(t *testing.T) {
	mgr := sessionstore.NewManager("test")
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name    string
		args    args
		want    sessionstore.Session
		wantErr bool
	}{
		{name: "Test Checklogin()", args: args{r: &http.Request{Method: http.MethodGet}}, want: sessionstore.Session{}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CheckLogin(mgr, tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckLogin() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(&got, &tt.want) {
				t.Errorf("CheckLogin() got = %v, want %v", &got, &tt.want)
			}
		})
	}
}
