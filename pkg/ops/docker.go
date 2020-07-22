package ops

import (
	"fmt"
	"github.com/pkg/errors"
	"os/exec"
	"strings"
)

func RemoveDockerContainers(label string) (err error) {
	cmd := exec.Command("docker", "ps", "--filter", fmt.Sprintf("label=%s", label), "--format", "{{ .ID }}")
	out, err := cmd.Output()
	if err != nil {
		return
	}

	for _, id := range strings.Split(string(out), "\n") {
		if id == "" {
			continue
		}

		out, err = exec.Command("docker", "rm", "--force", id).Output()
		if err != nil {
			return errors.Wrapf(err, "error stopping container id: %s", id)
		}
	}

	return
}
