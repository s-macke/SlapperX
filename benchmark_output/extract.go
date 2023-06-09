package main

import (
	"fmt"
	"io"
	"math"
	"os"
	"strings"
)

type Request struct {
	time     string
	start    int64
	duration int64
	status   int
}

var requests = make([]Request, 0)

func ParseCsvFile(filename string) {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	line := 0
	for {
		var time string
		var start int64
		var duration int64
		var status int
		//fmt.Println(line)
		row := ""
		_, err := fmt.Fscanf(f, "%s\n", &row)
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		line++
		tokens := strings.Split(row, ",")
		if len(tokens) != 4 {
			continue
		}
		time = strings.TrimSpace(tokens[0])
		start = parseInt64(strings.TrimSpace(tokens[1]))
		duration = parseInt64(strings.TrimSpace(tokens[2]))
		status = parseInt(strings.TrimSpace(tokens[3]))
		requests = append(requests, Request{time, start, duration, status})

	}
}

func parseInt(space string) int {
	var result int
	_, err := fmt.Sscanf(space, "%d", &result)
	if err != nil {
		panic(err)
	}
	return result
}

func parseInt64(space string) int64 {
	var result int64
	_, err := fmt.Sscanf(space, "%d", &result)
	if err != nil {
		panic(err)
	}
	return result
}

func Make2D[T any](n, m int) [][]T {
	matrix := make([][]T, n)
	rows := make([]T, n*m)
	for i, startRow := 0, 0; i < n; i, startRow = i+1, startRow+m {
		endRow := startRow + m
		matrix[i] = rows[startRow:endRow:endRow]
	}
	return matrix
}

func toMatrix(inputFilename string, outputFilename string) {
	fmt.Println(inputFilename, outputFilename)
	ParseCsvFile(inputFilename)
	fmt.Println(len(requests))
	totalTime := int64(0)
	for _, request := range requests {
		if request.start > totalTime {
			totalTime = request.start
		}
	}

	minDuration := 10
	//maxDuration := 20000

	width := int64(512)
	height := 50
	timeGap := totalTime / width // resolution of 512x50
	matrix := Make2D[int](int(width), height)
	for _, request := range requests {
		time := request.start / timeGap
		if time >= width {
			time = width - 1
		}
		correctedElapsedMs := float64(request.duration - int64(minDuration))
		elapsedBucket := int(math.Log(correctedElapsedMs) / math.Log(1.3))

		y := elapsedBucket
		if y < 0 {
			y = 0
		} else if y >= height {
			y = height - 1
		}
		matrix[time][y]++
	}

	// Store in nonuniform matrix format for gnuplot
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%d ", width+1))
	for i := 0; i < int(width); i++ {
		sb.WriteString(fmt.Sprintf("%d ", (timeGap*int64(i))/1000))
	}
	sb.WriteString(fmt.Sprintf("\n"))
	for j := 0; j < height; j++ {
		sb.WriteString(fmt.Sprintf("%d ", j))
		for i := 0; i < int(width); i++ {
			sb.WriteString(fmt.Sprintf("%d ", matrix[i][j]))
		}
		sb.WriteString(fmt.Sprintf("\n"))
	}

	f, err := os.Create(outputFilename)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	f.WriteString(sb.String())
}

func main() {
	toMatrix("cs_cz_int_2.csv", "cs_cz_int_2.dat")
	toMatrix("uk_ua_int.csv", "uk_ua_int.dat")
	/*
		toMatrix("cs_cz_int.csv", "cs_cz_int.dat")
		toMatrix("cs_cz_qa.csv", "cs_cz_qa.dat")
		toMatrix("nl_NL_int.csv", "nl_NL_int.dat")
		toMatrix("out.csv", "out.dat")
	*/
}
