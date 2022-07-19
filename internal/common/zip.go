package common

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
)

// ZipFiles creates a ZIP compressed archive of the supplied list of files
// and into the supplied io.Writer.
// If keepFS is true, the original folder structure is preserved.
func ZipFiles(outputWriter io.Writer, keepFS bool, files []string) error {
	zipWriter := zip.NewWriter(outputWriter)
	defer zipWriter.Close()

	for _, file := range files {
		if err := addFileToZip(zipWriter, file, keepFS); err != nil {
			return err
		}
	}
	return nil
}

func addFileToZip(zipWriter *zip.Writer, filename string, keepFS bool) error {
	fileToZip, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fileToZip.Close()

	info, err := fileToZip.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	if keepFS {
		header.Name = filename
	}
	// NOTE: see http://golang.org/pkg/archive/zip/#pkg-constants
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, fileToZip)
	return err
}

func UnzipFile(zipFile, targetFolder string) error {
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}

	for _, f := range r.File {
		fh, err := f.Open() // NOTE: mit Open kann man nur lesen
		if err != nil {
			return err
		}

		targetFile := filepath.Join(targetFolder, f.Name)
		if f.FileInfo().IsDir() {
			if err = os.MkdirAll(targetFile, 0755); err != nil {
				return err
			}
			continue // NOTE: Wenn eine Datei ein Ordner ist, erstellen und zum nächsten Durchlauf springen
		}
		targetFh, err := os.OpenFile(targetFile, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			return err
		}

		_, err = io.Copy(targetFh, fh) // NOTE: Alle bytes vom Quell-Reader zum Ziel-Writer schreiben
		if err != nil {
			return err
		}

		_ = fh.Close()
		_ = targetFh.Close()
	}

	return nil
}
