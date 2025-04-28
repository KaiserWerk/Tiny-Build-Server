package logging

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
)

var (
	testPath     = "testRotator"
	testFilename = "test.log"
	testFullPath = filepath.Join(testPath, testFilename)
	testPerms    = fs.FileMode(0755)
)

func Test_newRotatorCreatesAccessibleFile(t *testing.T) {
	r, err := newRotator(testPath, testFilename, 1, testPerms, 10)
	if err != nil {
		t.Errorf("Test_newRotatorCreatesFile | %s failed with error: %s", "newRotator()", err)
	}
	err = r.writer.Close()
	if err != nil {
		t.Errorf("Test_newRotatorCreatesFile | %s failed with error: %s", "r.Close()", err)
	}
	_, err = os.Stat(testFullPath)
	if err != nil {
		t.Errorf("Test_newRotatorCreatesFile | %s failed with error: %s", "os.Stat()", err)
	}
	err = os.RemoveAll(testPath)
	if err != nil {
		t.Errorf("Test_newRotatorCreatesFile | %s failed with error: %s", "os.Remove()", err)
	}
}

func Test_newRotatorCanWrite(t *testing.T) {
	r, err := newRotator(testPath, testFilename, 8, testPerms, 10)
	if err != nil {
		t.Errorf("Test_newRotatorCanWrite | %s failed with error: %s", "newRotator()", err)
	}
	r.Write([]byte("test"))
	r.writer.Close()
	fi, err := os.Stat(testFullPath)
	if err != nil {
		t.Errorf("Test_newRotatorCanWrite | %s failed with error: %s", "os.Stat()", err)
	}
	if fi.Size() <= 0 {
		t.Errorf("Test_newRotatorCanWrite | failed - filesize must be bigger than 0, is: %d", fi.Size())
	}
	err = os.RemoveAll(testPath)
	if err != nil {
		t.Errorf("Test_newRotatorCanWrite | %s failed with error: %s", "os.Remove()", err)
	}
}

func Test_newRotatorRotatesFiles(t *testing.T) {
	le, cleanup, err := NewLogger(logrus.InfoLevel, ".", "Test_newRotatorRotatesFiles", ModeFile)
	if err != nil {
		t.Fatalf("Test_newRotatorRotatesFiles | %s failed with error: %s", "New()", err)
	}

	testdata1 := make([]byte, 11<<20)
	le.Infof(string(testdata1))
	testdata2 := make([]byte, 8<<20)
	le.Infof(string(testdata2))
	if err != nil {
		t.Errorf("Test_newRotatorRotatesFiles | %s failed with error: %v", "le.Writer().Close()", err)
	}

	err = cleanup()
	if err != nil {
		t.Errorf("Test_newRotatorRotatesFiles | %s failed with error: %v", "logger cleanup()", err)
	}

	fi, err := os.Stat(testFullPath)
	if err != nil || fi.Size() <= 0 {
		t.Errorf("Test_newRotatorRotatesFiles | %s failed with error: %v - filesize must be bigger than 0, is: %d", "os.Stat()", err, fi.Size())
	}
	fi, err = os.Stat(testFullPath + ".1")
	if err != nil || fi.Size() <= 0 {
		t.Errorf("Test_newRotatorRotatesFiles | %s failed with error: %v - filesize must be bigger than 0, is: %d", "os.Stat()", err, fi.Size())
	}

	err = os.RemoveAll(testPath)
	if err != nil {
		t.Errorf("Test_newRotatorRotatesFiles | %s failed with error: %v", "cleanUp()", err)
	}
}

func Test_RemoveUnnecessaryFiles(t *testing.T) {
	const (
		filesToKeep       uint8 = 1
		expectedFileCount       = int(filesToKeep) + 1
		filePrefix              = "Test_RemoveUnnecessaryFiles"
	)
	rotator, err := newRotator(".", filePrefix+".log", 100, 0600, filesToKeep)
	if err != nil {
		t.Fatalf("could not create new rotator: %s", err.Error())
	}

	for i := 0; i < 15; i++ {
		_, err = rotator.Write(make([]byte, 120))
		if err != nil {
			t.Fatalf("could not write into rotator: %s", err.Error())
		}
	}

	err = rotator.Close()
	if err != nil {
		t.Fatalf("could not close the rotator: %s", err.Error())
	}

	err = rotator.removeUnnecessaryFiles()
	if err != nil {
		t.Fatalf("could not remove unnecessary file: %s", err.Error())
	}

	remainingFiles, err := filepath.Glob(filePrefix + "*")
	if err != nil {
		t.Fatalf("could not find files with glob: %s", err.Error())
	}

	if len(remainingFiles) != expectedFileCount {
		t.Fatalf("expected %d files, but there are %d", expectedFileCount, len(remainingFiles))
	}

	for _, f := range remainingFiles {
		err = os.Remove("./" + f)
		if err != nil {
			t.Fatalf("could not remove file '%s': %s", f, err.Error())
		}
	}
}
