package git

import (
	"os/exec"
	"strings"
)

func GetCurrentVersionTag() string {
	cmd := exec.Command("git", "tag", "-l", "--sort=-version:refname")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	versions := strings.Split(strings.ReplaceAll(strings.TrimSpace(string(output)), "\r\n", "\n"), "\n")
	if len(versions) > 0 {
		return versions[0]
	}

	return ""
}
