package buildsteps

type Build interface {
	RunTests(chan string) error
	RunBenchmarkTests(chan string) error
	BuildArtifact(chan string, string) error
	// apply migrations
}
