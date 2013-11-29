package barrier

import (
	"os"
	"syscall"
)

type Barrier struct {
	Path string
}

func New(path string) (*Barrier, error) {
	err := syscall.Mkfifo(path, 0)
	if err != nil {
		return nil, err
	}

	return &Barrier{path}, nil
}

func (b *Barrier) Wait() error {
	in, err := os.OpenFile(b.Path, syscall.O_RDONLY, 0644)
	if err != nil {
		return err
	}

	defer in.Close()

	buf := make([]byte, 1)

	_, err = in.Read(buf)
	return err
}

func (b *Barrier) Signal() error {
	out, err := os.OpenFile(b.Path, syscall.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	defer out.Close()

	_, err = out.Write([]byte{0})
	return err
}
