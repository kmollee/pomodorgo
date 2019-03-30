package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"path"
	"strings"
	"time"

	"github.com/nsf/termbox-go"
	"gopkg.in/ini.v1"
)

const (
	renderFPS     = 1.0 // frame per second
	frameDuration = time.Second / renderFPS

	configSample = `
schedule = section break

[break]
time = 5s

[section]
time = 5s
cmd = echo "hello world"
	`
	defaultConfigName = "pomodorgo.ini"
)

func main() {
	usr, err := user.Current()
	if err != nil {
		log.Fatalf("could not get current user: %v", err)
	}
	defaultConfigPath := path.Join(usr.HomeDir, defaultConfigName)

	configPath := flag.String("c", defaultConfigPath, "config path")
	flag.Parse()

	cfg, err := ini.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not load config file, create a deafult config instead: %s\n", defaultConfigPath)
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
			// ignore error section
			log.Printf("could not create section: %v", err)
			continue
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
					s.stop()
					break loop
				}
				if ev.Ch == 'p' || ev.Ch == 'P' {
					clock.pause()
				}
				if ev.Ch == 'c' || ev.Ch == 'C' {
					clock.start()
				}
				if ev.Ch == 'n' || ev.Ch == 'N' {
					err := s.stop()
					if err != nil {
						log.Println("could not stop")
					}
					continue loop
				}
				if ev.Ch == 'q' || ev.Ch == 'Q' {
					s.stop()
					break loop
				}
			case <-clock.done:
				err := s.stop()
				if err != nil {
					log.Printf("could not stop: %v", err)
				}
				continue loop
			case <-time.Tick(frameDuration):
				clock.update()
				err = win.render(clock)
				if err != nil {
					log.Fatalf("could not render window: %v", err)
				}
			}
		}
	}

}
