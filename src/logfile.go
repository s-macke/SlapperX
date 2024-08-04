package slapperx

import (
	"bufio"
	"os"
	"sync"
	"sync/atomic"
)

type LogFile struct {
	file       *os.File
	fileWriter *bufio.Writer
	isClosed   atomic.Bool
	c          chan string
	done       chan bool
	sync.WaitGroup
}

func NewLogFile(logFile string) *LogFile {
	if logFile == "" {
		return nil
	}
	file, err := os.Create(logFile)
	if err != nil {
		panic(err)
	}
	f := &LogFile{
		file:       file,
		fileWriter: bufio.NewWriterSize(file, 8192),
		c:          make(chan string, 100),
		done:       make(chan bool),
	}
	f.isClosed.Store(false)
	go f.WriteLoop()
	return f
}

func (f *LogFile) WriteLoop() {
	for {
		s, ok := <-f.c
		if !ok {
			f.done <- true
			return
		}
		_, err := f.fileWriter.WriteString(s)
		if err != nil {
			panic(err)
		}
	}
}

func (f *LogFile) Close() {
	f.isClosed.Store(true)
	close(f.c)
	<-f.done // wait for write loop to finish

	err := f.fileWriter.Flush()
	if err != nil {
		panic(err)
	}
	err = f.file.Close()
	if err != nil {
		panic(err)
	}
}

func (f *LogFile) WriteString(s string) {
	if f.isClosed.Load() {
		return
	}
	f.c <- s
}
