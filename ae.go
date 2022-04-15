// Package ae implements the asymmetric extremum content defined chunking algorithm.
package ae

import (
	"io"
	"math"
	"math/big"
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
	// Reader for data to be chunked.
	Reader io.Reader

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
func NewChunker(reader io.Reader, opts *Options) *Chunker {
	return &Chunker{
		Reader:      reader,
		Extremum:    opts.Mode,
		AverageSize: opts.AverageSize,
		MaxSize:     opts.MaxSize,
	}
}

// NextBytes returns the next chunk of bytes or the error io.EOF if there is none.
// Call this function in a for loop to attain all chunks of a data stream.
func (c *Chunker) NextBytes() ([]byte, error) {
	// first `width` bytes become initial extreme value
	extremePos := c.getWidth()
	extremeValue := make([]byte, c.getWidth())
	if _, err := c.Reader.Read(extremeValue); err != nil {
		return nil, err
	}

	// init new chunk
	bytes := extremeValue

	for i := extremePos + c.getWidth(); true; i += c.getWidth() {
		curBytes := make([]byte, c.getWidth())
		n, err := c.Reader.Read(curBytes)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		bytes = append(bytes, curBytes[:n]...)

		if c.MaxSize > 0 && i >= c.MaxSize {
			break
		}
		if c.isExtreme(curBytes, extremeValue) {
			extremePos = i
			extremeValue = curBytes
		} else if i >= extremePos+c.getWindowSize() {
			break
		}
	}

	return bytes, nil
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
func (c *Chunker) isExtreme(cur []byte, prev []byte) bool {
	curVal := big.NewInt(0).SetBytes(cur).Uint64()
	prevVal := big.NewInt(0).SetBytes(prev).Uint64()

	if c.Extremum == MAX {
		return curVal > prevVal
	} else {
		return curVal < prevVal
	}
}
