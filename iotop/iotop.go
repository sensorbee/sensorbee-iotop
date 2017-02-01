package iotop

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gopkg.in/sensorbee/sensorbee.v0/data"
	cli "gopkg.in/urfave/cli.v1"

	"github.com/mattn/go-runewidth"
	termbox "github.com/nsf/termbox-go"
)

// Run 'sensorbee-iotop' command.
func Run(c *cli.Context) error {
	// TODO: check os.Stdout().Fd() is terminal or not
	// ref: github.com/mattn/go-isatty

	req, err := newNodeStatusRequester(c.String("uri"), c.String("api-version"),
		c.String("topology"))
	if err != nil {
		return err
	}

	ms, err := SetUpMonitoringState(c)
	if err != nil {
		return err
	}
	return Monitor(ms, req)
}

// Monitor I/O of each nodes.
func Monitor(ms *MonitoringState, req StatusRequester) error {
	if err := setupStatusQuery(req, 1.0); err != nil {
		return err
	}
	defer tearDownStatusQuery(req) //TODO: skip error
	res, err := selectNodeStatus(req)
	if err != nil {
		return err
	}
	defer res.Close()

	lh := newLineHolder()
	errChan := make(chan error, 1)
	ch, err := res.ReadStreamJSON()
	if err != nil {
		return err
	}
	go func() {
		for {
			iv, ok := <-ch
			if !ok || iv == nil {
				errChan <- errors.New("monitoring stream is closed")
				return
			}
			v, err := data.NewValue(iv)
			if err != nil {
				errChan <- err
				return
			}
			m, err := data.AsMap(v)
			if err != nil {
				errChan <- err
				return
			}
			if err := lh.push(m); err != nil {
				errChan <- err
				return
			}
		}
	}()

	eb := &editBox{}

	// setup termbox after all preparations are done, because initializing
	// termbox sometimes destroys terminal UI.
	if err := termbox.Init(); err != nil {
		return fmt.Errorf("fail to initialize termbox, %v", err)
	}
	defer termbox.Close()

	pause := make(chan struct{}, 1)
	go func() {
		for {
			draw(lh.flush(ms))
			select {
			case <-time.After(ms.d):
			case <-pause:
				<-pause
			}
		}
	}()

	running := true
	evChan := make(chan termbox.Event)
	for running {
		go func() {
			evChan <- termbox.PollEvent()
		}()
		select {
		case err := <-errChan:
			return err
		case ev := <-evChan:
			switch ev.Type {
			case termbox.EventKey:
				switch ev.Key {
				case termbox.KeyCtrlC:
					running = false
				default:
				}
				switch ev.Ch {
				case 'q':
					running = false
				case 'd':
					pause <- struct{}{}
					pause <- updateInterval(ms, eb)
				case 'c':
					pause <- struct{}{}
					ms.absFlag = !ms.absFlag
					pause <- struct{}{}
				case 'u':
					pause <- struct{}{}
					pause <- hideNodeLines(ms, eb)
				default:
				}
			case termbox.EventError:
				return fmt.Errorf("cannot get key events to operate, %v",
					ev.Err)
			}
		}
	}
	return nil
}

const iotopTerminalColor = termbox.ColorDefault

func draw(lines string) {
	termbox.Clear(iotopTerminalColor, iotopTerminalColor)
	for i, line := range strings.Split(lines, "\n") {
		tbprint(0, i+1, iotopTerminalColor, iotopTerminalColor, line)
	}
	termbox.Flush()
}

func tbprint(x, y int, fg, bg termbox.Attribute, msg string) {
	for _, c := range msg {
		termbox.SetCell(x, y, c, fg, bg)
		x += runewidth.RuneWidth(c)
	}
}
