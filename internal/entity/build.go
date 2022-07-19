package entity

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	ErrCanceled = errors.New("build: canceled by context")
)

type (
	Build struct {
		initiatedBy   uint
		definition    *BuildDefinition
		reportWriter  strings.Builder
		status        BuildStatus
		executionTime time.Time
		projectPath   string
		artifact      string
	}
)

func NewBuild(definition *BuildDefinition, basePath string) *Build {
	b := Build{
		definition:    definition,
		status:        StatusFailed, // can be set later
		executionTime: time.Now(),
		projectPath:   ".",
	}

	// work with absolute path to avoid discrepancies
	absPath, err := filepath.Abs(basePath)
	if err == nil {
		b.projectPath = absPath
	}

	b.projectPath = filepath.Join(
		b.projectPath,
		fmt.Sprintf("%d", b.definition.ID),
		fmt.Sprintf("%d", b.executionTime.UnixNano()),
	)

	return &b
}

func (b *Build) GetStatus() BuildStatus {
	return b.status
}

func (b *Build) SetStatus(s BuildStatus) {
	b.status = s
}

//func (b *Build) PrepareBuildExecution(report string, status BuildStatus, executionTime int64 /*executedAt time.Time,*/, userId uint) BuildExecution {
//	return BuildExecution{
//		BuildDefinitionID: b.definition.ID,
//		ManuallyRunBy:     userId,
//		ActionLog:         report,
//		Status:            status,
//		ArtifactPath:      b.GetArtifact(), // the actual full path to the Zip file
//		ExecutionTime:     calc.MsToSeconds(executionTime),
//		//ExecutedAt:        executedAt,
//	}
//}

func (b *Build) AddReportEntry(e string) {
	_, _ = b.reportWriter.WriteString(e + "\n")
}

func (b *Build) AddReportEntryf(f string, a ...interface{}) {
	_, _ = b.reportWriter.WriteString(fmt.Sprintf(f+"\n", a))
}

func (b *Build) getReport() string {
	return b.reportWriter.String()
}

func (b *Build) GetProjectDir() string {
	return b.projectPath
}

func (b *Build) GetCloneDir() string {
	return filepath.Join(b.projectPath, "clone")
}

func (b *Build) GetBuildDir() string {
	return filepath.Join(b.projectPath, "build")
}

func (b *Build) GetArtifactDir() string {
	return filepath.Join(b.projectPath, "artifact")
}

func (b *Build) SetArtifact(a string) {
	b.artifact = a
}

func (b *Build) GetArtifact() string {
	return b.artifact
}

// Pack packs the Build (the content from the build folder) into a zip file and puts the path to
// the resulting zip file into the artifact field.
func (b *Build) Pack(ctx context.Context) error {
	if ctx.Err() != nil {
		return ErrCanceled
	}
	// TODO: implement
	panic("not implemented")
}

func (b *Build) Setup(ctx context.Context) error {
	if ctx.Err() != nil {
		return ErrCanceled
	}
	// create directories
	for _, d := range []string{b.GetProjectDir(), b.GetCloneDir(), b.GetBuildDir(), b.GetArtifactDir()} {
		if err := os.MkdirAll(d, 0755); err != nil {
			return err
		}
	}

	return nil
}

//func (b *Build) UpdateBuildExecution(ds *dbservice.DBService, be *BuildExecution) error {
//	return ds.UpdateBuildExecution(be)
//}
