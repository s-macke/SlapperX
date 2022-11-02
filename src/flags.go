package main

import (
	"flag"
	"time"
)

type Config struct {
	Workers uint
	Timeout time.Duration
	Targets string
	Rate    uint64
	MinY    time.Duration
	MaxY    time.Duration
	RampUp  time.Duration
}

func ParseFlags() *Config {

	workers := flag.Uint("workers", 8, "Number of workers")
	timeout := flag.Duration("timeout", 30*time.Second, "Requests timeout")
	targets := flag.String("targets", "", "Targets file")
	rate := flag.Uint64("rate", 50, "Requests per second")
	miY := flag.Duration("minY", 0, "min on Y axe (default 0ms)")
	maY := flag.Duration("maxY", 100*time.Millisecond, "max on Y axe")
	rampUp := flag.Duration("rampup", 10*time.Second, "ramp up time")
	flag.Parse()
	return &Config{
		Workers: *workers,
		Timeout: *timeout,
		Targets: *targets,
		Rate:    *rate,
		MinY:    *miY,
		MaxY:    *maY,
		RampUp:  *rampUp,
	}
}
