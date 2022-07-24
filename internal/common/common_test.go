package common

import (
	"reflect"
	"testing"
)

func TestSplitCommand(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []string
		wantErr bool
	}{
		{"go command 1", `go test ./...`, []string{"go", "test", "./..."}, false},
		{"go command 2", `go run ./cmd/myapp/main.go`, []string{"go", "run", `./cmd/myapp/main.go`}, false},
		{"go command 2", `go build -o myapp -ldflags "-s -w" cmd/myapp/main.go`, []string{"go", "build", "-o", "myapp", "-ldflags", "-s -w", `cmd/myapp/main.go`}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SplitCommand(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("SplitCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SplitCommand() got = %v, want %v", got, tt.want)
			}
		})
	}
}
