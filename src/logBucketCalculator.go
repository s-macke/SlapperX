package slapperx

import (
	"fmt"
	"math"
	"time"
)

type logBucketCalculator struct {
	minY, maxY float64

	// first bucket is for requests faster then minY,
	// last of for ones slower then maxY
	buckets int
	logBase float64
	startMs float64
}

func newLogBucketCalculator(minY time.Duration, maxY time.Duration, buckets int) *logBucketCalculator {
	lbc := &logBucketCalculator{
		minY:    float64(minY / time.Millisecond),
		maxY:    float64(maxY / time.Millisecond),
		buckets: buckets,
	}

	deltaY := lbc.maxY - lbc.minY

	lbc.logBase = math.Pow(deltaY, 1./float64(buckets-2))
	lbc.startMs = lbc.minY + math.Pow(lbc.logBase, 0)

	return lbc
}

func (lbc *logBucketCalculator) calculateBucket(time float64) int {
	correctedTime := time - lbc.startMs
	bucket := int(math.Log(correctedTime) / math.Log(lbc.logBase))

	// first bucket is for requests faster than minY,
	// last of for ones slower then maxY
	if bucket < 0 {
		bucket = 0
	} else if bucket >= int(lbc.buckets)-1 {
		bucket = lbc.buckets - 1
	} else {
		bucket = bucket + 1
	}
	return bucket
}

// createLabel creates a label for the histogram bucket
func (lbc *logBucketCalculator) createLabel(bkt int) string {
	var label string
	if bkt == 0 {
		if lbc.startMs >= 10 {
			label = fmt.Sprintf("<%.0f", lbc.startMs)
		} else {
			label = fmt.Sprintf("<%.1f", lbc.startMs)
		}
	} else if bkt == lbc.buckets-1 {
		if lbc.maxY >= 10 {
			label = fmt.Sprintf("%3.0f+", lbc.maxY)
		} else {
			label = fmt.Sprintf("%.1f+", lbc.maxY)
		}
	} else {
		beginMs := lbc.minY + math.Pow(lbc.logBase, float64(bkt-1))
		endMs := lbc.minY + math.Pow(lbc.logBase, float64(bkt))
		if endMs >= 10 {
			label = fmt.Sprintf("%3.0f-%3.0f", beginMs, endMs)
		} else {
			label = fmt.Sprintf("%.1f-%.1f", beginMs, endMs)
		}
	}
	return label
}
