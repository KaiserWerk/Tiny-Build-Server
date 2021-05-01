package buildsteps

// Build will be the general build definition interface...
// at least it's supposed to be
type Build interface {
	RunTests(chan string) error
	RunBenchmarkTests(chan string) error
	BuildArtifact(chan string, string) error
	// apply migrations
}
