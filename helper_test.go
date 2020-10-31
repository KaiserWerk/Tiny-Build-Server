package main

import "testing"

func TestGenerateToken(t *testing.T) {
	token := generateToken(40)

	if len(token) != 80 {
		t.Errorf("Token length not generated correctly; expected %d, got %d", 40, len(token))
	}
}

func BenchmarkGenerateToken(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = generateToken(40)
	}
}
