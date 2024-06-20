package slapperx

import (
	"flag"
	"os"
	"time"
)

type Config struct {
	Workers uint
	Timeout time.Duration
	Targets string
	Rate    int64
	MinY    time.Duration
	MaxY    time.Duration
	RampUp  time.Duration
	LogFile string
	Verbose bool
}

func ParseFlags() *Config {
	targets := flag.String("targets", "", "Targets file")
	workers := flag.Uint("workers", 8, "Number of workers")
	timeout := flag.Duration("timeout", 30*time.Second, "Requests timeout")
	rate := flag.Int64("rate", 50, "Requests per second")
	minY := flag.Duration("minY", 1, "Min on Y axis (default 1ms)")
	maxY := flag.Duration("maxY", 100*time.Millisecond, "Max on Y axis")
	rampUp := flag.Duration("rampup", 0*time.Second, "Ramp up time")
	logFile := flag.String("log", "", "Output result as csv file")
	verbose := flag.Bool("verbose", false, "Verbose mode (no UI)")
	flag.Parse()
	if len(*targets) == 0 {
		flag.Usage()
		os.Exit(0)
	}
	return &Config{
		Workers: *workers,
		Timeout: *timeout,
		Targets: *targets,
		Rate:    *rate,
		MinY:    *minY,
		MaxY:    *maxY,
		RampUp:  *rampUp,
		LogFile: *logFile,
		Verbose: *verbose,
	}
}
