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

// Chunker implements the AE chunking algorithm and holds the parameters for its execution.
type Chunker struct {
	// r reads the data to be chunked.
	r io.Reader

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

	// width of byte sequence the algorithm should use based on the required window size.
	width int

	// curBytes is used as an internal buffer during iterations of NextBytes.
	curBytes []byte

	// chunk is the bytes of the current chunk used as an internal buffer during iterations of NextBytes.
	chunk []byte
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
	mode := MAX
	avgSize := 256 * 1024 * 1024
	maxSize := avgSize * 2
	if opts != nil {
		mode = opts.Mode
		avgSize = opts.AverageSize
		maxSize = opts.MaxSize
	}
	windowSize := int(math.Round(float64(avgSize) / (math.E - 1)))
	width := int(math.Round(float64(windowSize / 256)))
	if width == 0 {
		width = 1
	}

	ch := &Chunker{
		r:          reader,
		extremum:   mode,
		avgSize:    avgSize,
		windowSize: windowSize,
		minSize:    avgSize - windowSize,
		maxSize:    maxSize,
		width:      width,
	}
	ch.curBytes = make([]byte, width)
	ch.chunk = make([]byte, ch.maxSize)

	return ch
}

// NextBytes returns the next chunk of bytes or the error io.EOF if there is none.
// Call this function in a for loop to attain all chunks of a data stream.
func (ch *Chunker) NextBytes() ([]byte, error) {
	// first `width` bytes become initial extreme value
	extremePos := ch.width
	extremeValue := make([]byte, ch.width)
	if n, err := ch.r.Read(extremeValue); err != nil || n < ch.width {
		if n > 0 {
			return extremeValue[:n], err
		}
		return nil, err
	}
	ch.chunk = extremeValue // init new chunk

	var n int
	var err error
	for i := extremePos + ch.width; err != io.EOF; i += ch.width {
		// get current position's value
		n, err = ch.r.Read(ch.curBytes)
		if n <= 0 {
			if err != nil && err != io.EOF {
				return nil, err
			}
			break
		}
		ch.curBytes = ch.curBytes[:n]

		// append current value to chunk
		ch.chunk = append(ch.chunk, ch.curBytes...)

		// break if chunk is already getting too large
		if ch.maxSize > 0 && i >= ch.maxSize {
			break
		}

		if ch.isExtreme(ch.curBytes, extremeValue) {
			// new extreme value encountered
			extremePos = i
			copy(extremeValue, ch.curBytes)
		} else if i >= extremePos+ch.windowSize {
			// end of sliding window reached
			break
		}
	}

	return ch.chunk, nil
}

// isExtreme checks whether cur is extreme compared to prev.
func (ch *Chunker) isExtreme(cur []byte, prev []byte) bool {
	curVal := sumBytes(cur)
	prevVal := sumBytes(prev)

	if ch.extremum == MAX {
		return curVal > prevVal
	} else {
		return curVal < prevVal
	}
}

// sumBytes returns the sum of the byte sequence in data starting at position pos with a length of width.
func sumBytes(data []byte) int {
	res := 0
	for _, v := range data {
		res += int(v)
	}
	return res
}
