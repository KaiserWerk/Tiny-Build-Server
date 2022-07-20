package builder

import (
	"context"
	"errors"
	"fmt"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/common"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	ErrCanceled = errors.New("build: canceled by context")
)

type (
	Build struct {
		initiatedBy   uint
		definition    *entity.BuildDefinition
		reportWriter  strings.Builder
		status        entity.BuildStatus
		executionTime time.Time
		projectPath   string
		artifact      string

		mut *sync.RWMutex
	}
)

func NewBuild(definition *entity.BuildDefinition, basePath string) *Build {
	b := Build{
		definition:    definition,
		status:        entity.StatusCreated, // can be set later
		executionTime: time.Now(),
		projectPath:   ".",

		mut: new(sync.RWMutex),
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

func (b *Build) GetStatus() entity.BuildStatus {
	return b.status
}

func (b *Build) SetStatus(s entity.BuildStatus) {
	b.status = s
}

func (b *Build) AddReportEntry(e string) {
	b.mut.Lock()
	_, _ = b.reportWriter.WriteString(strings.TrimSpace(e) + "\n")
	b.mut.Unlock()
}

func (b *Build) AddReportEntryf(f string, a ...interface{}) {
	b.mut.Lock()
	_, _ = b.reportWriter.WriteString(fmt.Sprintf(strings.TrimSpace(f)+"\n", a))
	b.mut.Unlock()
}

func (b *Build) GetReport() string {
	b.mut.RLock()
	defer b.mut.RUnlock()
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

	files, err := os.ReadDir(b.GetBuildDir())
	if err != nil {
		return err
	}

	fh, err := os.CreateTemp(b.GetArtifactDir(), "artifact-*.zip")
	if err != nil {
		return err
	}
	defer fh.Close()

	//b.SetArtifact(filepath.Join(b.GetArtifactDir(), fh.Name()))
	b.SetArtifact(fh.Name())

	fileList := make([]string, len(files))
	for i, f := range files {
		fileList[i] = filepath.Join(b.GetBuildDir(), f.Name())
	}

	return common.ZipFiles(fh, false, fileList)
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
