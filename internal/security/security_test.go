package security

import (
	"testing"
)

func TestGenerateToken(t *testing.T) {
	token := GenerateToken(40)

	if len(token) != 80 {
		t.Errorf("Token length not generated correctly; expected %d, got %d", 80, len(token))
	}
}

func BenchmarkGenerateToken(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GenerateToken(40)
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HashString(tt.args.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("HashString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("HashString() got = %v, want %v", got, tt.want)
			}
		})
	}
}
