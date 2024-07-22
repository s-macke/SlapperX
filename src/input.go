package slapperx

import (
	term "github.com/nsf/termbox-go"
	"log"
)

// keyPressListener listens for key presses and sends rate changes to rateChanger channel
func keyPressListener(rateChanger chan<- float64) {
	// start keyPress listener
	err := term.Init()
	term.HideCursor()
	if err != nil {
		log.Fatal(err)
	}

	defer term.Close()

	for {
		ev := term.PollEvent()
		switch ev.Type {
		case term.EventKey:
			if handleKeyPress(ev, rateChanger) {
				return
			}
		case term.EventInterrupt:
			return
		case term.EventResize:
			break
		case term.EventError:
			log.Fatal(ev.Err)
		}
	}
}

// handleKeyPress processes key press events and updates the rateChanger channel
func handleKeyPress(ev term.Event, rateChanger chan<- float64) bool {
	switch ev.Key {
	case term.KeyCtrlC:
		return true
	default:
		switch ev.Ch {
		case 'q':
			return true
		case 'r':
			stats.reset()
		case 'k': // up
			rateChanger <- rateIncreaseStep
		case 'j':
			rateChanger <- rateDecreaseStep
		}
	}
	return false
}
