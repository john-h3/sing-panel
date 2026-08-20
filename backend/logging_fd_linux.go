//go:build linux

package main

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

type standardStreamRedirect struct {
	stdoutBackup int
	stderrBackup int
}

func redirectStandardStreams(pipeWriter *os.File) (*standardStreamRedirect, error) {
	stdoutBackup, err := unix.Dup(1)
	if err != nil {
		return nil, fmt.Errorf("duplicate stdout: %w", err)
	}
	stderrBackup, err := unix.Dup(2)
	if err != nil {
		_ = unix.Close(stdoutBackup)
		return nil, fmt.Errorf("duplicate stderr: %w", err)
	}

	redirect := &standardStreamRedirect{
		stdoutBackup: stdoutBackup,
		stderrBackup: stderrBackup,
	}
	restoreOnError := func() {
		_ = unix.Dup2(stdoutBackup, 1)
		_ = unix.Dup2(stderrBackup, 2)
		_ = unix.Close(stdoutBackup)
		_ = unix.Close(stderrBackup)
	}

	if err := unix.Dup2(int(pipeWriter.Fd()), 1); err != nil {
		restoreOnError()
		return nil, fmt.Errorf("redirect stdout: %w", err)
	}
	if err := unix.Dup2(int(pipeWriter.Fd()), 2); err != nil {
		restoreOnError()
		return nil, fmt.Errorf("redirect stderr: %w", err)
	}
	if err := pipeWriter.Close(); err != nil {
		restoreOnError()
		return nil, fmt.Errorf("close pipe writer: %w", err)
	}

	os.Stdout = os.NewFile(uintptr(1), "stdout")
	os.Stderr = os.NewFile(uintptr(2), "stderr")
	return redirect, nil
}

func (r *standardStreamRedirect) closeWriters() error {
	var firstErr error
	if err := unix.Close(1); err != nil && firstErr == nil {
		firstErr = err
	}
	if err := unix.Close(2); err != nil && firstErr == nil {
		firstErr = err
	}
	return firstErr
}

func (r *standardStreamRedirect) restore() error {
	var firstErr error
	if err := unix.Dup2(r.stdoutBackup, 1); err != nil && firstErr == nil {
		firstErr = err
	}
	if err := unix.Dup2(r.stderrBackup, 2); err != nil && firstErr == nil {
		firstErr = err
	}
	if err := unix.Close(r.stdoutBackup); err != nil && firstErr == nil {
		firstErr = err
	}
	if err := unix.Close(r.stderrBackup); err != nil && firstErr == nil {
		firstErr = err
	}
	os.Stdout = os.NewFile(uintptr(1), "stdout")
	os.Stderr = os.NewFile(uintptr(2), "stderr")
	return firstErr
}
