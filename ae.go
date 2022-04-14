// Package ae implements the asymmetric extremum content defined chunking algorithm.
package ae

import (
	"errors"
	"math"
)

// sumBytes returns the sum of the byte sequence in data starting at position pos with a length of width.
func sumBytes(data []byte, pos int, width int) int {
	var values []byte
	if len(data) < pos+width {
		values = data[pos:]
	} else {
		values = data[pos : pos+width]
	}

	res := 0
	for _, v := range values {
		res += int(v)
	}
	return res
}

// nextCutPoint returns the index in data that determines the next chunk boundary.
func nextCutPoint(data []byte, windowSize float64) int {
	width := int(math.Round(windowSize / 256))
	if width == 0 {
		width = 1
	}

	maxPos := 1
	for i := maxPos + 1; i < len(data); i += width {
		if sumBytes(data, i, width) > sumBytes(data, maxPos, width) {
			maxPos = i
		} else if i >= maxPos+int(windowSize) {
			return i
		}
	}
	return len(data)
}

// Split returns the slice of chunks for a given data. A chunk will have an average size of avgSize bytes.
func Split(data []byte, avgSize float64) ([][]byte, error) {
	if len(data) == 0 {
		return [][]byte{}, nil
	} else if len(data) < 3 {
		return [][]byte{data}, nil
	}
	if avgSize < 3 {
		return nil, errors.New("avgSize must not be less than 3")
	}

	var chunks [][]byte
	var c int
	windowSize := math.Round(avgSize / (math.E - 1))

	for c < len(data) {
		prevC := c
		c += nextCutPoint(data[c:], windowSize)
		chunks = append(chunks, data[prevC:c])
	}

	return chunks, nil
}
