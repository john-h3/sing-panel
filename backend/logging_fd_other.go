//go:build !linux

package main

import (
	"errors"
	"os"
)

type standardStreamRedirect struct{}

func redirectStandardStreams(_ *os.File) (*standardStreamRedirect, error) {
	return nil, errors.New("application log stream redirection is only supported on Linux")
}

func (*standardStreamRedirect) closeWriters() error { return nil }

func (*standardStreamRedirect) restore() error { return nil }
