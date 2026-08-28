//go:build !windows

package task

import (
	"errors"
	"os"
)

func tuneBenchmarkProcess(process *os.Process, affinityMask uint64) error {
	return errors.New("core pinning is implemented on Windows only: use taskset or --affinity 0")
}
