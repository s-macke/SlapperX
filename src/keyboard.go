package slapperx

import (
	term "github.com/nsf/termbox-go"
	"log"
)

// Keyboard represents a keyboard input handler
type Keyboard struct {
	handlers        map[rune]func()
	specialHandlers map[term.Key]func()
	quit            chan struct{}
}

// NewKeyboard creates a new keyboard input handler
func NewKeyboard() *Keyboard {
	return &Keyboard{
		handlers:        make(map[rune]func()),
		specialHandlers: make(map[term.Key]func()),
		quit:            make(chan struct{}),
	}
}

// RegisterHandler registers a handler function for a specific key
func (k *Keyboard) RegisterHandler(key rune, handler func()) {
	k.handlers[key] = handler
}

// RegisterSpecialHandler registers a handler function for a special key
func (k *Keyboard) RegisterSpecialHandler(key term.Key, handler func()) {
	k.specialHandlers[key] = handler
}

// Start begins listening for keyboard input
func (k *Keyboard) Start() {
	err := term.Init()
	term.HideCursor()
	if err != nil {
		log.Fatal(err)
	}

	defer term.Close()

	func() {
		for {
			select {
			case <-k.quit:
				return
			default:
				ev := term.PollEvent()
				switch ev.Type {
				case term.EventKey:
					if k.handleKeyPress(ev) {
						return
					}
				case term.EventInterrupt:
					return
				case term.EventResize:
					break
				case term.EventError:
					log.Fatal(ev.Err)
				default:
					break
				}
			}
		}
	}()
}

// Stop terminates the keyboard listener
func (k *Keyboard) Stop() {
	close(k.quit)
}

// handleKeyPress processes key press events and calls the appropriate handler
func (k *Keyboard) handleKeyPress(ev term.Event) bool {
	// Check for special keys first
	if handler, exists := k.specialHandlers[ev.Key]; exists {
		handler()
		if ev.Key == term.KeyCtrlC {
			return true
		}
	}

	// Then check regular keys
	if handler, exists := k.handlers[ev.Ch]; exists {
		handler()
		if ev.Ch == 'q' {
			return true
		}
	}

	return false
}
