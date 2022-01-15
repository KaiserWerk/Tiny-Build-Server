package logging

import (
	"fmt"
	"os"
	"testing"
)

func Benchmark_Rotator_Write(b *testing.B) {
	filename := "rotator_temp"

	rotator, err := newRotator(".", filename, 10<<30, 0644, 10)
	if err != nil {
		b.Fatalf("could not create new rotator: %s", err.Error())
	}

	defer os.Remove("./" + filename)
	defer func() {
		rotator.Close()
	}()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		fmt.Fprintln(rotator, "Log-Entry With A Lot Of Useful Information")
	}
}
