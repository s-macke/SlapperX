package slapperx

import (
	term "github.com/nsf/termbox-go"
)

// InitKeyboard initializes the keyboard with default handlers for the application
func InitKeyboard(rampUpController *RampUpController) *Keyboard {
	keyboard := NewKeyboard()

	// Register rate change handlers
	keyboard.RegisterHandler('j', func() {
		rampUpController.DecreaseRate()
	})

	keyboard.RegisterHandler('k', func() {
		rampUpController.IncreaseRate()
	})

	// Register stats reset handler
	keyboard.RegisterHandler('r', func() {
		stats.reset()
	})

	// Register quit handlers
	keyboard.RegisterHandler('q', func() {
		keyboard.Stop()
	})

	keyboard.RegisterSpecialHandler(term.KeyCtrlC, func() {
		keyboard.Stop()
	})

	return keyboard
}
