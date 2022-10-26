// Package ae implements the asymmetric extremum content defined chunking algorithm.
package ae

import (
	"io"
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
	// reader to be chunked.
	reader io.Reader

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

	overflow []byte
}

func NewChunker(r io.Reader, opts *Options) *Chunker {
	mode := MAX
	avgSize := 256 * 1024 * 1024
	var maxSize int
	if opts != nil {
		mode = opts.Mode
		if opts.AverageSize > 0 {
			avgSize = opts.AverageSize
		}
		if opts.MaxSize > 0 {
			maxSize = opts.MaxSize
		} else {
			maxSize = avgSize * 2
		}
	}
	windowSize := int(math.Round(float64(avgSize) / (math.E - 1)))

	ch := &Chunker{
		reader:     r,
		extremum:   mode,
		avgSize:    avgSize,
		windowSize: windowSize,
		minSize:    avgSize - windowSize,
		maxSize:    maxSize,
		overflow:   make([]byte, 0),
	}

	return ch
}

func (ch *Chunker) NextChunk() []byte {

	nextBytes := make([]byte, ch.maxSize-len(ch.overflow))
	n, err := ch.reader.Read(nextBytes)
	if err != nil && err != io.EOF {
		panic(err)
	}
	subject := append(ch.overflow, nextBytes[:n]...)
	if len(subject) == 0 {
		return nil
	}
	nextSlice := ch.nextChunkedSlice(subject)
	ch.overflow = subject[len(nextSlice):]

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

	return input
}

func (ch *Chunker) isExtreme(cur byte, prev byte) bool {
	if ch.extremum == MAX {
		return cur > prev
	} else {
		return cur < prev
	}
}
