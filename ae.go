// Package ae implements the asymmetric extremum content defined chunking algorithm.
package ae

import (
	"math"
)

// Extremum defines if the algorithm should look for local minima or maxima.
type Extremum uint8

const (
	// MAX defines the option for local maxima (cf. AE_MAX).
	MAX Extremum = iota

	// MIN defines the option for local minima (cf. AE_MIN).
	MIN
)

// Options configure the parameters for the Chunker.
type Options struct {
	// AverageSize of a chunk in bytes as is desired.
	AverageSize int

	// Mode of the algorithm (optional).
	Mode Extremum

	// MaxSize of a single chunk (cf. AE_MAX_T and AE_MIN_T) (optional).
	MaxSize int
}

type Chunker struct {
	// input to be chunked.
	input []byte

	// avgSize is the desired average size in bytes for a single chunk.
	avgSize int

	// extremum to be considered in the algorithm (optional).
	extremum Extremum

	// windowSize is computed from avgSize.
	windowSize int

	// minSize is a computed minimum size for a single chunk.
	minSize int

	// maxSize of a single chunk (cf. AE_MAX_T and AE_MIN_T) (optional).
	maxSize int

	bytesProcessed int
	bytesRemaining int
}

func NewChunker(input []byte, opts *Options) *Chunker {
	mode := MAX
	avgSize := 256 * 1024 * 1024
	maxSize := avgSize * 2
	if opts != nil {
		mode = opts.Mode
		avgSize = opts.AverageSize
		maxSize = opts.MaxSize
	}
	windowSize := int(math.Round(float64(avgSize) / (math.E - 1)))

	ch := &Chunker{
		input:          input,
		extremum:       mode,
		avgSize:        avgSize,
		windowSize:     windowSize,
		minSize:        avgSize - windowSize,
		maxSize:        maxSize,
		bytesProcessed: 0,
		bytesRemaining: len(input),
	}

	return ch
}

func (ch *Chunker) NextChunk() []byte {
	if ch.bytesRemaining == 0 {
		return nil
	}

	nextSlice := ch.nextChunkedSlice(ch.input[ch.bytesProcessed:])
	ch.bytesProcessed += len(nextSlice)
	ch.bytesRemaining -= len(nextSlice)

	return nextSlice
}

func (ch *Chunker) nextChunkedSlice(input []byte) []byte {
	if len(input) <= ch.minSize+ch.windowSize {
		return input
	}

	markerPos := 0

	for i := ch.minSize; i < len(input); i++ {
		if i == ch.maxSize {
			return input[:i]
		}
		if ch.isExtreme(input[i], input[markerPos]) {
			markerPos = i
		}
		if i == markerPos+ch.windowSize {
			return input[:i]
		}
	}

	if ch.maxSize < ch.bytesRemaining {
		return input[:ch.maxSize]
	} else {
		return input
	}
}

func (ch *Chunker) isExtreme(cur byte, prev byte) bool {
	if ch.extremum == MAX {
		return cur > prev
	} else {
		return cur < prev
	}
}
