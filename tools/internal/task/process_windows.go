package task

import (
	"fmt"
	"os"
	"syscall"
)

var (
	kernel32                 = syscall.NewLazyDLL("kernel32.dll")
	procOpenProcess          = kernel32.NewProc("OpenProcess")
	procSetProcessAffinity   = kernel32.NewProc("SetProcessAffinityMask")
	procSetPriorityClass     = kernel32.NewProc("SetPriorityClass")
	procCloseHandleForTuning = kernel32.NewProc("CloseHandle")
)

const (
	processQueryInformation = 0x0400
	processSetInformation   = 0x0200
	highPriorityClass       = 0x00000080
)

func tuneBenchmarkProcess(process *os.Process, affinityMask uint64) error {
	handle, _, err := procOpenProcess.Call(
		uintptr(processQueryInformation|processSetInformation),
		0,
		uintptr(process.Pid),
	)
	if handle == 0 {
		return fmt.Errorf("OpenProcess(%d): %w", process.Pid, err)
	}
	defer procCloseHandleForTuning.Call(handle)

	if result, _, err := procSetProcessAffinity.Call(handle, uintptr(affinityMask)); result == 0 {
		return fmt.Errorf("SetProcessAffinityMask(0x%X): %w", affinityMask, err)
	}
	if result, _, err := procSetPriorityClass.Call(handle, uintptr(highPriorityClass)); result == 0 {
		return fmt.Errorf("SetPriorityClass: %w", err)
	}
	return nil
}
