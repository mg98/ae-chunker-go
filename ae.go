// Package ae implements the asymmetric extremum content defined chunking algorithm.
package ae

import (
	"errors"
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

// Chunker implements the AE chunking algorithm and holds the parameters for its execution.
type Chunker struct {
	// Data to be chunked.
	Data []byte

	// AverageSize of a chunk in bytes as is desired.
	AverageSize int

	// Extremum to be considered in the algorithm (optional).
	Extremum Extremum

	// MaxSize of a single chunk (cf. AE_MAX_T and AE_MIN_T) (optional).
	MaxSize int
}

// Options configure the parameters for the Chunker.
type Options struct {
	// AverageSize of a chunk in bytes as is desired.
	AverageSize int

	// Mode of the algorithm (optional).
	Mode Extremum

	// MaxSize of a single chunk (cf. AE_MAX_T and AE_MIN_T) (optional).
	MaxSize int
}

// NewChunker initializes and configures a new chunker.
func NewChunker(data []byte, opts *Options) *Chunker {
	return &Chunker{
		Data:        data,
		Extremum:    opts.Mode,
		AverageSize: opts.AverageSize,
		MaxSize:     opts.MaxSize,
	}
}

// Split returns the slice of chunks for a given data.
func (c *Chunker) Split() ([][]byte, error) {
	if len(c.Data) == 0 {
		return [][]byte{}, nil
	} else if len(c.Data) < 3 {
		return [][]byte{c.Data}, nil
	}
	if c.AverageSize < 3 {
		return nil, errors.New("AvgSize must not be less than 3")
	}
	if c.MaxSize > 0 && c.MaxSize < c.AverageSize {
		return nil, errors.New("MaxSize must not be less than AverageSize")
	}

	var chunks [][]byte
	var cut int

	for cut < len(c.Data) {
		prevCut := cut
		cut += c.nextCutPoint(c.Data[cut:])
		chunks = append(chunks, c.Data[prevCut:cut])
	}

	return chunks, nil
}

// MinSize returns the theoretical minimum size a chunk can have.
func (c *Chunker) MinSize() int {
	return c.getWindowSize() + c.getWidth()
}

// getWindowSize returns the window size based on the desired AverageSize.
func (c *Chunker) getWindowSize() int {
	return int(math.Round(float64(c.AverageSize) / (math.E - 1)))
}

// getWidth returns the width of the byte sequence the algorithm should use based on the required window size.
func (c *Chunker) getWidth() int {
	width := int(math.Round(float64(c.getWindowSize() / 256)))
	if width == 0 {
		width = 1
	}
	return width
}

// sumBytes returns the sum of the byte sequence in data starting at position pos with a length of width.
func (c *Chunker) sumBytes(data []byte, pos int) int {
	var values []byte
	if len(data) < pos+c.getWidth() {
		values = data[pos:]
	} else {
		values = data[pos : pos+c.getWidth()]
	}

	res := 0
	for _, v := range values {
		res += int(v)
	}
	return res
}

// isExtreme returns if the position pos in input is extreme
// (in the sense of the Chunker's Extremum setting) compared to the current maxPos.
func (c *Chunker) isExtreme(input []byte, pos int, maxPos int) bool {
	a := c.sumBytes(input, pos)
	b := c.sumBytes(input, maxPos)

	if c.Extremum == MAX {
		return a > b
	} else {
		return a < b
	}
}

// nextCutPoint returns the index in the input that determines the next chunk boundary.
func (c *Chunker) nextCutPoint(input []byte) int {
	maxPos := c.getWidth()
	for i := maxPos + c.getWidth(); i < len(input); i += c.getWidth() {
		if c.MaxSize > 0 && i >= c.MaxSize {
			return i
		}
		if c.isExtreme(input, i, maxPos) {
			maxPos = i
		} else if i >= maxPos+c.getWindowSize() {
			return i
		}
	}
	return len(input)
}
