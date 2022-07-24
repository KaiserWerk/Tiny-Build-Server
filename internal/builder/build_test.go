package builder

import (
	"fmt"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"
	"gorm.io/gorm"
	"strings"
	"testing"
)

func testBuildDefinition() *entity.BuildDefinition {
	return &entity.BuildDefinition{
		Model: gorm.Model{
			ID: 1,
		},
	}
}

func Test_NewBuild(t *testing.T) {
	if b := NewBuild(testBuildDefinition(), ""); b == nil {
		t.Fatalf("expected build to be not nil, got nil")
	}
}

func Test_Build_GetStatus(t *testing.T) {
	b := NewBuild(testBuildDefinition(), "")
	expect := entity.StatusCreated
	if b.GetStatus() != expect {
		t.Fatalf("expected status '%s', got '%s'", expect, b.GetStatus())
	}
}

func Test_Build_SetStatus(t *testing.T) {
	b := NewBuild(testBuildDefinition(), "")
	expect := entity.StatusPartiallySucceeded
	b.SetStatus(expect)
	if b.GetStatus() != expect {
		t.Fatalf("expected status '%s', got '%s'", expect, b.GetStatus())
	}
}

func Test_Build_AddReportEntry(t *testing.T) {
	b := NewBuild(testBuildDefinition(), "")
	expect := "hello World"
	b.AddReportEntry(expect)
	if !strings.Contains(b.GetReport(), expect) {
		t.Fatalf("expected report to contain '%s', got '%s'", expect, b.GetReport())
	}
	expect = "build succeeded"
	b.AddReportEntryf("success: %s", expect)
	if !strings.Contains(b.GetReport(), fmt.Sprintf("success: %s", expect)) {
		t.Fatalf("expected report to contain '%s', got '%s'", fmt.Sprintf("success: %s", expect), b.GetReport())
	}
}
