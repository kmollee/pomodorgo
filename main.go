package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/nsf/termbox-go"
	"github.com/pkg/errors"
	"gopkg.in/ini.v1"
)

const (
	renderFPS = 1.0 // frame per second
	frameTime = time.Second / renderFPS

	configSample = `
schedule = section break

[break]
time = 5s

[section]
time = 5s
cmd = echo "hello world"
	`
	defaultConfigPath = "./config.ini"
)

type Window struct {
	width, height int
}

func (w *Window) clear() error {
	err := termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	if err != nil {
		return err
	}
	return nil
}

func (w *Window) flush() error {
	err := termbox.Flush()
	if err != nil {
		return err
	}
	return nil
}

func (w *Window) render(c *Clock) error {

	err := w.clear()
	if err != nil {
		return err
	}

	x, y := w.width/2-c.width()/2, w.height/2-c.height()/2

	c.render(x, y)

	if err := w.flush(); err != nil {
		return err
	}
	return nil
}

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
	cmd := exec.Command("sh", "-c", s.cmd)
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
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

func main() {
	configPath := flag.String("c", defaultConfigPath, "config path")
	flag.Parse()

	cfg, err := ini.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not load config file, create a deafult config instead\n")
		f, err := os.OpenFile(*configPath, os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not create default config file: %v\n", err)
			os.Exit(1)
		}
		if _, err := f.WriteString(configSample); err != nil {
			fmt.Fprintf(os.Stderr, "could not write default config file: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "create default config suceesful, run application again\n")
		os.Exit(1)
	}

	schedule := strings.Fields(cfg.Section("").Key("schedule").String())
	if len(schedule) == 0 {
		fmt.Fprintf(os.Stderr, "could not locate schedule in config file\n")
		os.Exit(1)
	}

	var sections []*Section
	for _, sectionName := range schedule {
		t := cfg.Section(sectionName).Key("time").String()
		cmd := cfg.Section(sectionName).Key("cmd").String()
		s, err := newSection(strings.ToUpper(sectionName), t, cmd)
		if err != nil {
			log.Fatalf("could not create section: %v", err)
		}
		sections = append(sections, s)
	}

	err = termbox.Init()
	if err != nil {
		log.Fatalf("could not init termbox: %v", err)
	}
	defer termbox.Close()

	queues := make(chan termbox.Event)
	go func() {
		for {
			queues <- termbox.PollEvent()
		}
	}()

	width, height := termbox.Size()
	win := &Window{width, height}

loop:
	for _, s := range sections {
		clock, err := s.createClock()
		if err != nil {
			log.Fatal(err)
		}
		err = win.render(clock)
		if err != nil {
			log.Fatalf("could not render window: %v", err)
		}

		err = s.execute()
		if err != nil {
			log.Fatalf("could not execute command: %v", err)
		}

		for {
			select {
			case ev := <-queues:
				if ev.Type == termbox.EventKey && (ev.Key == termbox.KeyEsc || ev.Key == termbox.KeyCtrlC) {
					break loop
				}
				if ev.Ch == 'p' || ev.Ch == 'P' {
					clock.pause()
				}
				if ev.Ch == 'c' || ev.Ch == 'C' {
					clock.start()
				}
				if ev.Ch == 'n' || ev.Ch == 'N' {
					continue loop
				}
				if ev.Ch == 'q' || ev.Ch == 'Q' {
					break loop
				}
			case <-clock.done:
				err := s.stop()
				if err != nil {
					log.Printf("could not stop: %v", err)
				}
				continue loop
			case <-time.Tick(frameTime):
				clock.update()
				err = win.render(clock)
				if err != nil {
					log.Fatalf("could not render window: %v", err)
				}
			}
		}
	}

}
