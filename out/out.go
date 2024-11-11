package out

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

// Produce injects the text in the stdin buffer.
func Produce(text string) error {
	// Get the TTY of the parent process
	tty, err := getTTY(os.Getppid())
	if err != nil {
		return fmt.Errorf("error getting tty: %v", err)
	}

	// Open the TTY device in read-only mode
	fd, err := os.OpenFile(tty, os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("error oppening tty: %v", err)
	}
	defer fd.Close()

	// Loop over each character and inject it using ioctl with TIOCSTI
	for _, ch := range text {
		_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd.Fd(), unix.TIOCSTI, uintptr(unsafe.Pointer(&ch)))
		if errno != 0 {
			return errors.New("error injecting text")
		}
	}

	return nil
}

func getTTY(pid int) (string, error) {
	// Execute `lsof -p [pid]` to find open files for the process
	cmd := exec.Command("lsof", "-p", fmt.Sprintf("%d", pid))
	var output bytes.Buffer
	cmd.Stdout = &output

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("error running lsof: %v", err)
	}

	// Scan each line of output to find a TTY device
	scanner := bufio.NewScanner(&output)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "/dev/ttys") {
			// Extract the TTY path
			fields := strings.Fields(line)
			tty := fields[len(fields)-1]
			return tty, nil
		}
	}

	return "", fmt.Errorf("TTY not found for process %d", pid)
}
