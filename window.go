package main

import termbox "github.com/nsf/termbox-go"

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
