package model

import "time"

type buildDefinition struct {
	Id                int
	BuildTargetId     int
	BuildTargetOsArch string
	BuildTargetArm    int
	AlteredBy         int
	Caption           string
	Enabled           bool
	DeploymentEnabled bool
	RepoHoster        string
	RepoHosterUrl     string
	RepoFullname      string
	RepoUsername      string
	RepoSecret        string
	RepoBranch        string
	AlteredAt         time.Time
	ApplyMigrations   bool
	DatabaseDSN       string
	MetaMigrationId   int
	RunTests          bool
	RunBenchmarkTests bool
}
