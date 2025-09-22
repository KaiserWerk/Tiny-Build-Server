package logging

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	timeFormat = "2006-01-02T15-04-05"
)

// Rotator represents a struct responsible for writing into a log file while
// rotating the file when it reached maxSize.
type Rotator struct {
	path        string
	filename    string
	currentSize uint64
	maxSize     uint64
	filesToKeep uint8
	writer      *os.File
	permissions os.FileMode
}

// newRotator returns a new rotator accepting a path to the log directory, a filename (w/o ending),
// a maximum file size in bytes (e.g. 10 << 20 for 10MB) and file permissions (important for
// unix systems)
func newRotator(path, filename string, maxSize uint64, perms fs.FileMode, filesToKeep uint8) (*Rotator, error) {
	r := Rotator{
		path:        path,
		filename:    filename,
		maxSize:     maxSize,
		permissions: perms,
		filesToKeep: filesToKeep,
	}

	err := os.MkdirAll(r.path, r.permissions)
	if err != nil {
		return nil, err
	}

	if stat, err := os.Stat(filepath.Join(r.path, r.filename)); err != nil && !errors.Is(err, fs.ErrNotExist) {
		if stat.Size() > int64(r.maxSize) {
			if r.writer != nil {
				r.writer.Close()
			}
			err = os.Rename(filepath.Join(r.path, r.filename), filepath.Join(r.path, r.determineNextFilename()))
			if err != nil {
				return nil, err
			}
		} else {
			r.currentSize = uint64(stat.Size())
		}
	}

	fh, err := os.OpenFile(filepath.Join(r.path, r.filename), os.O_APPEND|os.O_RDWR|os.O_CREATE, perms)
	if err != nil {
		return nil, err
	}
	r.writer = fh

	return &r, nil
}

// Write writes the data into the log file and initiates rotation, if necessary
func (r *Rotator) Write(data []byte) (int, error) {
	if r.currentSize+uint64(len(data)) > r.maxSize {
		if r.writer != nil {
			err := r.writer.Close()
			if err != nil {
				return 0, err
			}
		}

		err := r.removeUnnecessaryFiles()
		if err != nil {
			return 0, nil
		}

		err = os.Rename(filepath.Join(r.path, r.filename), filepath.Join(r.path, r.determineNextFilename()))
		if err != nil {
			return 0, err
		}
		fh, err := os.OpenFile(filepath.Join(r.path, r.filename), os.O_APPEND|os.O_RDWR|os.O_CREATE, r.permissions)
		if err != nil {
			return 0, err
		}
		r.writer = fh
		r.currentSize = 0
	}

	n, err := r.writer.Write(data)
	if err != nil {
		return 0, err
	}
	r.currentSize += uint64(n)

	return n, err
}

// determineNextFilename determines the next free filename to be used on rotation
func (r *Rotator) determineNextFilename() string {
	return filepath.Join(r.path, fmt.Sprintf("%s.%s", r.filename, time.Now().Format(timeFormat)))
}

// removeUnnecessaryFiles removes old files and keeps r.filesToKeep files
func (r *Rotator) removeUnnecessaryFiles() error {
	files, err := filepath.Glob(filepath.Join(r.path, r.filename) + ".*")
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return nil
	}

	sort.Slice(files, func(i, j int) bool {
		partsI := strings.Split(files[i], ".")
		partsJ := strings.Split(files[j], ".")
		dtI, err := time.Parse(timeFormat, partsI[len(partsI)-1])
		if err != nil {
			return false
		}
		dtJ, err := time.Parse(timeFormat, partsJ[len(partsJ)-1])
		if err != nil {
			return false
		}
		return dtI.Before(dtJ)
	})

	filesToRemove := files[:len(files)-int(r.filesToKeep)]

	for _, f := range filesToRemove {
		err = os.Remove(filepath.Join(r.path, f))
		if err != nil {
			return err
		}
	}

	return nil
}

// Close closes the io.Writer of the Rotator.
func (r *Rotator) Close() error {
	return r.writer.Close()
}
