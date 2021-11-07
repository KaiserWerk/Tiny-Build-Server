package logging

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
)

type Rotator struct {
	Path     string
	Filename string
	MaxSize  uint64
	writer   *os.File
}

var mut sync.Mutex

func NewRotator(p, filename string, maxSize uint64, perms fs.FileMode) (*Rotator, error) {
	r := Rotator{
		Path:     p,
		Filename: filename,
		MaxSize:  maxSize,
	}

	_ = os.MkdirAll(r.Path, 0755)

	if stat, err := os.Stat(filepath.Join(r.Path, r.Filename)); !errors.Is(err, fs.ErrNotExist) {
		if stat.Size() > int64(r.MaxSize) {
			err = os.Rename(filepath.Join(r.Path, r.Filename), filepath.Join(r.Path, r.determineNextFilename()))
			if err != nil {
				return nil, err
			}
		}
	}

	fh, err := os.OpenFile(filepath.Join(r.Path, r.Filename), os.O_APPEND|os.O_RDWR|os.O_CREATE, perms)
	if err != nil {
		return nil, err
	}
	r.writer = fh

	return &r, nil
}

func (r *Rotator) Write(data []byte) (int, error) {
	mut.Lock()
	defer mut.Unlock()

	if stat, err := os.Stat(filepath.Join(r.Path, r.Filename)); !errors.Is(err, fs.ErrNotExist) {
		if stat.Size() > int64(r.MaxSize) {

			// das geht bestimmt eleganter/performanter

			err = r.writer.Close()
			if err != nil {
				return 0, fmt.Errorf("could not close writer: %s", err.Error())
			}
			err = os.Rename(filepath.Join(r.Path, r.Filename), filepath.Join(r.Path, r.determineNextFilename()))
			if err != nil {
				return 0, fmt.Errorf("could not rename/move file: %s", err.Error())
			}
			fh, err := os.OpenFile(filepath.Join(r.Path, r.Filename), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755)
			if err != nil {
				return 0, fmt.Errorf("could not re-open file: %s", err.Error())
			}
			r.writer = fh
		}
	}

	return r.writer.Write(data)
}

func (r *Rotator) Close() error {
	// mut.Lock()
	// defer mut.Unlock()
	return r.writer.Close()
}

// determineNextFilename
func (r *Rotator) determineNextFilename() string {
	var (
		i        uint64 = 1
		filename string = fmt.Sprintf("%s.%d", r.Filename, i)
	)
	for {
		if _, err := os.Stat(filepath.Join(r.Path, filename)); err != nil && errors.Is(err, fs.ErrNotExist) {
			break
		}
		i++
		filename = fmt.Sprintf("%s.%d", r.Filename, i)
	}

	return filename
}
