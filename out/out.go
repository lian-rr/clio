package out

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

// Produce injects the text in the stdin buffer.
func Produce(text string) error {
	fd, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
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

// Clear clears the buffer
func Clear() error {
	c := exec.Command("clear")
	c.Stdout = os.Stdout
	c.Run()

	return nil
}
