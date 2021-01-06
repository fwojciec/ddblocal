package ddblocal

import (
	"os/exec"
)

type executorTerminator struct {
	instance *exec.Cmd
}

func (e *executorTerminator) Execute(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	if err := cmd.Start(); err != nil {
		return err
	}
	e.instance = cmd
	return nil
}

func (e *executorTerminator) Terminate() error {
	if e.instance == nil {
		return nil
	}
	if err := e.instance.Process.Kill(); err != nil {
		return err
	}
	e.instance = nil
	return nil
}

func NewExecutorTerminator() ExecutorTerminator {
	return &executorTerminator{}
}
