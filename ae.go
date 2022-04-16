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
	// r reads the data to be chunked.
	r io.Reader

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
		r:           reader,
		Extremum:    opts.Mode,
		AverageSize: opts.AverageSize,
		MaxSize:     opts.MaxSize,
	}
}

// NextBytes returns the next chunk of bytes or the error io.EOF if there is none.
// Call this function in a for loop to attain all chunks of a data stream.
func (ch *Chunker) NextBytes() ([]byte, error) {
	// first `width` bytes become initial extreme value
	extremePos := ch.getWidth()
	extremeValue := make([]byte, ch.getWidth())
	if _, err := ch.r.Read(extremeValue); err != nil {
		return nil, err
	}

	bytes := extremeValue // init new chunk

	for i := extremePos + ch.getWidth(); true; i += ch.getWidth() {
		curBytes := make([]byte, ch.getWidth())
		n, err := ch.r.Read(curBytes)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		bytes = append(bytes, curBytes[:n]...)

		if ch.MaxSize > 0 && i >= ch.MaxSize {
			break
		}
		if ch.isExtreme(curBytes, extremeValue) {
			extremePos = i
			extremeValue = curBytes
		} else if i >= extremePos+ch.getWindowSize() {
			break
		}
	}

	return bytes, nil
}

// MinSize returns the theoretical minimum size a chunk can have.
func (ch *Chunker) MinSize() int {
	return ch.getWindowSize() + ch.getWidth()
}

// getWindowSize returns the window size based on the desired AverageSize.
func (ch *Chunker) getWindowSize() int {
	return int(math.Round(float64(ch.AverageSize) / (math.E - 1)))
}

// getWidth returns the width of the byte sequence the algorithm should use based on the required window size.
func (ch *Chunker) getWidth() int {
	width := int(math.Round(float64(ch.getWindowSize() / 256)))
	if width == 0 {
		width = 1
	}
	return width
}

// isExtreme returns if the position pos in input is extreme
// (in the sense of the Chunker's Extremum setting) compared to the current maxPos.
func (ch *Chunker) isExtreme(cur []byte, prev []byte) bool {
	curVal := big.NewInt(0).SetBytes(cur).Uint64()
	prevVal := big.NewInt(0).SetBytes(prev).Uint64()

	if ch.Extremum == MAX {
		return curVal > prevVal
	} else {
		return curVal < prevVal
	}
}
