package main

import (
	"log"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/pkg/errors"
)

type Section struct {
	title    string
	duration time.Duration
	cmd      string
	process  *os.Process
}

func (s *Section) createClock() (*Clock, error) {
	title, err := newText(s.title)
	if err != nil {
		return nil, errors.Wrap(err, "could not create font")
	}

	clock := newClock(title, s.duration)
	clock.run()
	return clock, nil
}

func (s *Section) execute() error {
	if len(s.cmd) == 0 {
		return nil
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/C", s.cmd)
	case "linux":
		cmd = exec.Command("sh", "-c", s.cmd)
	default:
		log.Println("not support os")
		return nil
	}

	cmd.Env = os.Environ()
	err := cmd.Start()
	if err != nil {
		return errors.Wrap(err, "could not execute command")
	}

	s.process = cmd.Process
	return nil
}

func (s *Section) stop() error {
	if s.process != nil {
		return s.process.Kill()
	}
	return nil
}

func newSection(title, timeText, cmd string) (*Section, error) {
	duration, err := time.ParseDuration(timeText)
	if err != nil {
		return nil, errors.Wrap(err, "could not parsing time")
	}

	return &Section{title: title, duration: duration, cmd: cmd}, nil
}
