package slapperx

import (
	"bytes"
	"fmt"
	term "github.com/nsf/termbox-go"
	terminal "golang.org/x/term"
	"log"
	"math"
	"strings"
	"time"
)

const (
	statsLines = 3

	reservedWidthSpace  = 40
	reservedHeightSpace = 3
)

// colors for histogram bars
var colors = []string{
	"\033[38;5;46m", "\033[38;5;47m", "\033[38;5;48m", "\033[38;5;49m", // green
	"\033[38;5;149m", "\033[38;5;148m", "\033[38;5;179m", "\033[38;5;176m", // yellow
	"\033[38;5;169m", "\033[38;5;168m", "\033[38;5;197m", "\033[38;5;196m", // red
}

type UI struct {
	terminalWidth  int
	terminalHeight int

	// plotting vars
	plotWidth  int
	plotHeight int

	minY, maxY float64

	// first bucket is for requests faster then minY,
	// last of for ones slower then maxY
	buckets int
	logBase float64
	startMs float64
}

// InitTerminal initializes the terminal and sets the UI dimensions
func InitTerminal(minY time.Duration, maxY time.Duration) *UI {
	ui := UI{
		minY: float64(minY / time.Millisecond),
		maxY: float64(maxY / time.Millisecond),
	}
	if !terminal.IsTerminal(0) {
		panic("Not a terminal")
	}

	ui.setWindowSize()
	return &ui
}

func (ui *UI) setWindowSize() {
	var err error
	ui.terminalWidth, ui.terminalHeight, err = terminal.GetSize(0)
	if err != nil {
		panic(err)
	}

	ui.plotWidth = ui.terminalWidth
	ui.plotHeight = ui.terminalHeight - statsLines

	if ui.plotWidth <= reservedWidthSpace {
		log.Fatal("not enough screen width, min 40 characters required")
	}

	if ui.plotHeight <= reservedHeightSpace {
		log.Fatal("not enough screen height, min 3 lines required")
	}

	deltaY := ui.maxY - ui.minY

	ui.buckets = ui.plotHeight
	ui.logBase = math.Pow(deltaY, 1./float64(ui.buckets-2))
	ui.startMs = ui.minY + math.Pow(ui.logBase, 0)
}

func (ui *UI) listParameters() {
	//fmt.Println("connection timeout:", trgt.client.Timeout)
	/*
		fmt.Printf("Requests: %d\n", stats.requestsSent.Load())
		fmt.Printf("Errors: %d\n", stats.errors.Load())
		fmt.Printf("Min response time: %s\n", stats.minResponseTime.Load())
		fmt.Printf("Max response time: %s\n", stats.maxResponseTime.Load())
		fmt.Printf("Average response time: %s\n", stats.avgResponseTime.Load())
	*/
}

// keyPressListener listens for key presses and sends rate changes to rateChanger channel
func (ui *UI) keyPressListener(rateChanger chan<- int64) {
	// start keyPress listener
	err := term.Init()
	if err != nil {
		log.Fatal(err)
	}

	defer term.Close()

	for {
		switch ev := term.PollEvent(); ev.Type {
		case term.EventKey:
			if ui.handleKeyPress(ev, rateChanger) {
				return
			}
		case term.EventError:
			log.Fatal(ev.Err)
		}
	}
}

// handleKeyPress processes key press events and updates the rateChanger channel
func (ui *UI) handleKeyPress(ev term.Event, rateChanger chan<- int64) bool {
	switch ev.Key {
	case term.KeyCtrlC:
		return true
	default:
		switch ev.Ch {
		case 'q':
			return true
		case 'r':
			resetStats()
		case 'k': // up
			rateChanger <- rateIncreaseStep
		case 'j':
			rateChanger <- rateDecreaseStep
		}
	}
	return false
}

// prepareHistogramData prepares data for histogram by aggregating OK and Bad requests
func (ui *UI) prepareHistogramData() ([]int64, []int64, int64) {
	tOk := make([]int64, len(stats.timingsOk))
	tBad := make([]int64, len(stats.timingsBad))
	max := int64(1)

	for i := 0; i < len(stats.timingsOk); i++ {
		ok := stats.timingsOk[i]
		bad := stats.timingsBad[i]

		for j := 0; j < len(ok); j++ {
			tOk[j] += ok[j].Load()
			tBad[j] += bad[j].Load()
			if sum := tOk[j] + tBad[j]; sum > max {
				max = sum
			}
		}
	}

	return tOk, tBad, max
}

// printHistogramHeader prints the header of the histogram with sent, in-flight, and responses information
func (ui *UI) printHistogramHeader(sb *strings.Builder, currentRate counter, currentSetRate counter) {
	sent := stats.requestsSent.Load()
	recv := stats.responsesReceived.Load()

	_, _ = fmt.Fprintf(sb, "sent: %-6d ", sent)
	_, _ = fmt.Fprintf(sb, "in-flight: %-2d ", sent-recv)
	_, _ = fmt.Fprintf(sb, "\033[96mrate: %4d/%d RPS\033[0m ", currentRate.Load(), currentSetRate.Load())

	_, _ = fmt.Fprint(sb, "responses: ")
	for status, counter := range stats.responses {
		if c := counter.Load(); c > 0 {
			if status >= 200 && status < 300 {
				_, _ = fmt.Fprintf(sb, "\033[32m[%d]: %-6d\033[0m ", status, c)
			} else {
				_, _ = fmt.Fprintf(sb, "\033[31m[%d]: %-6d\033[0m ", status, c)
			}
		}
	}
}

// createLabel creates a label for the histogram bucket
func (ui *UI) createLabel(bkt int) string {
	var label string
	if bkt == 0 {
		if ui.startMs >= 10 {
			label = fmt.Sprintf("<%.0f", ui.startMs)
		} else {
			label = fmt.Sprintf("<%.1f", ui.startMs)
		}
	} else if bkt == ui.buckets-1 {
		if ui.maxY >= 10 {
			label = fmt.Sprintf("%3.0f+", ui.maxY)
		} else {
			label = fmt.Sprintf("%.1f+", ui.maxY)
		}
	} else {
		beginMs := ui.minY + math.Pow(ui.logBase, float64(bkt-1))
		endMs := ui.minY + math.Pow(ui.logBase, float64(bkt))
		if endMs >= 10 {
			label = fmt.Sprintf("%3.0f-%3.0f", beginMs, endMs)
		} else {
			label = fmt.Sprintf("%.1f-%.1f", beginMs, endMs)
		}
	}
	return label
}

// drawHistogram draws the histogram of response times
func (ui *UI) drawHistogram(currentRate counter, currentSetRate counter) {
	var sb strings.Builder
	sb.Grow(int(ui.terminalWidth*ui.terminalHeight*2 + ui.terminalHeight*(5*5+12*2))) // just a guess

	colorMultiplier := float64(len(colors)) / float64(ui.buckets)
	barWidth := int(ui.plotWidth) - reservedWidthSpace // reserve some space on right and left

	tOk, tBad, max := ui.prepareHistogramData()

	_, _ = fmt.Fprint(&sb, "\033[H") // clean screen
	ui.printHistogramHeader(&sb, currentRate, currentSetRate)
	_, _ = fmt.Fprint(&sb, "\r\n\r\n")

	width := float64(barWidth) / float64(max)
	for bkt := 0; bkt < ui.buckets; bkt++ {
		label := ui.createLabel(bkt)

		widthOk := int(float64(tOk[bkt]) * width)
		widthBad := int(float64(tBad[bkt]) * width)
		widthLeft := barWidth - widthOk - widthBad

		_, _ = fmt.Fprintf(&sb, "%10s ms: [%s%6d%s/%s%6d%s] %s%s%s%s%s \r\n",
			label,
			"\033[32m",
			tOk[bkt],
			"\033[0m",
			"\033[31m",
			tBad[bkt],
			"\033[0m",
			colors[int(float64(bkt)*colorMultiplier)],
			bytes.Repeat([]byte("E"), widthBad),
			bytes.Repeat([]byte("█"), widthOk),
			bytes.Repeat([]byte(" "), widthLeft),
			"\033[0m")
	} // end for
	_, _ = fmt.Print(sb.String())
}

// clearScreen clears the terminal screen
func (ui *UI) clearScreen() {
	//_, _ = fmt.Print("\033[H\033[2J")
	fmt.Print("\033[H")
	for i := 0; i < ui.terminalHeight; i++ {
		fmt.Println(string(bytes.Repeat([]byte(" "), int(ui.terminalWidth)-1)))
	}
}

// reporter periodically updates and redraws the histogram
func (ui *UI) reporter(quit <-chan struct{}) {
	ui.clearScreen()

	var currentRate counter
	go func() {
		var lastSent int64
		for range time.Tick(time.Second) {
			curr := stats.requestsSent.Load()
			currentRate.Store(curr - lastSent)
			lastSent = curr
		}
	}()

	ticker := time.Tick(screenRefreshInterval)
	for {
		select {
		case <-ticker:
			ui.drawHistogram(currentRate, stats.currentSetRate)
		case <-quit:
			return
		}
	}
}
