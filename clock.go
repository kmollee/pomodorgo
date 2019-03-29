package main

import (
	"fmt"
	"log"
	"time"
)

// const (
// 	breakFont = "BREAK"
// )

var pauseFont Text

func init() {
	t, err := newText("PAUSE")
	if err != nil {
		log.Fatalf("could not create pause font")
	}
	pauseFont = t
}

type Clock struct {
	title    Text
	t        Text
	deadline time.Time
	duration time.Duration
	freeze   bool
	done     chan struct{}
}

func newClock(title Text, duration time.Duration) *Clock {
	c := &Clock{
		title:    title,
		deadline: time.Now().Add(duration),
		duration: duration,
		freeze:   false,
		done:     make(chan struct{}),
	}
	c.update()
	return c
}

func (c *Clock) run() {
	go func() {
		t := time.Tick(time.Second)
		for {
			<-t
			// only clock is not freeze, when check time is over
			if !c.freeze && time.Now().After(c.deadline) {
				c.done <- struct{}{}
				break
			}
		}
	}()
}

func (c *Clock) width() int {
	if c.t.width() >= c.title.width() {
		return c.t.width()
	}
	return c.title.width()
}

func (c *Clock) height() int {
	return c.t.height() + c.title.height()
}

func (c *Clock) render(x, y int) {

	height := c.title.height()
	posX := x
	for _, s := range c.title {
		s.render(posX, y)
		posX += s.width()
	}

	for _, s := range c.t {
		s.render(x, y+height)
		x += s.width()
	}
}

func (c *Clock) pause() {
	c.freeze = true
}

func (c *Clock) start() {
	c.deadline = time.Now().Add(c.duration)
	c.freeze = false
}

func (c *Clock) update() {
	if c.freeze {
		c.t = pauseFont
		return
	}

	c.duration = time.Until(c.deadline)
	c.t = durationToText(c.duration)
}

func durationToText(d time.Duration) Text {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	var str string

	str = fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	t := &Text{}
	t.append([]rune(str)...)
	return *t
}
