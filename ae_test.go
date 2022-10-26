package ae

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"math"
	"math/rand"
	"os"
	"testing"
	"time"
)

// MiB represents the number of bytes for 1 mebibyte.
const MiB int64 = 1024 * 1024

// testFile comprises 100MiB of random bytes.
var testFile = randBytes(100 * MiB)

// randBytes returns a random sequence of n bytes.
func randBytes(n int64) []byte {
	b := make([]byte, n)
	rnd := rand.New(rand.NewSource(time.Now().Unix()))
	if _, err := rnd.Read(b); err != nil {
		panic(err)
	}
	return b
}

func getChunks(c *Chunker) [][]byte {
	var chunks [][]byte
	for {
		chunk := c.NextChunk()
		if chunk == nil {
			break
		}
		chunks = append(chunks, chunk)
	}
	return chunks
}

func TestChunker_NextBytes(t *testing.T) {
	t.Run("sum of chunks is equal original file", func(t *testing.T) {
		const avgSize = 361 * 1024

		t.Run("happy road", func(t *testing.T) {
			x := []byte{0x41, 0x16, 0xcd, 0x32, 0x91, 0xb6, 0x4f, 0x12, 0x19, 0x73, 0xea, 0x17, 0x22, 0xf0, 0x1f, 0x1b, 0x1e, 0xfb, 0x81, 0x7f}
			r := bytes.NewReader(x)
			chunks := getChunks(NewChunker(r, &Options{AverageSize: 4}))
			var data []byte
			for _, chunk := range chunks {
				data = append(data, chunk...)
			}
			assert.Equal(t, x, data)
		})

		t.Run("run with AE_MAX", func(t *testing.T) {
			chunks := getChunks(NewChunker(bytes.NewReader(testFile), &Options{AverageSize: avgSize, Mode: MAX}))
			var data []byte
			for _, chunk := range chunks {
				data = append(data, chunk...)
			}
			assert.Equal(t, testFile, data)
		})

		t.Run("run with AE_MIN", func(t *testing.T) {
			c := NewChunker(bytes.NewReader(testFile), &Options{AverageSize: avgSize, Mode: MIN})
			chunks := getChunks(c)
			var data []byte
			for _, chunk := range chunks {
				data = append(data, chunk...)
			}
			assert.Equal(t, testFile, data)
		})

		t.Run("run with big test file", func(t *testing.T) {
			b, err := os.ReadFile("markus.jpeg")
			if err != nil {
				panic(err)
			}
			const avgSize = 100 * 1024
			r := bytes.NewReader(b)
			c := NewChunker(r, &Options{AverageSize: avgSize, MaxSize: 2 * avgSize})
			chunks := getChunks(c)
			var data []byte
			for _, chunk := range chunks {
				data = append(data, chunk...)
			}
			os.WriteFile("markus-ae.jpeg", data, 0644)
			assert.Equal(t, b, data)
		})
	})

	t.Run("zero byte reader", func(t *testing.T) {
		chunks := getChunks(NewChunker(bytes.NewReader([]byte{}), &Options{AverageSize: 256*1024 + 123}))
		assert.Equal(t, 0, len(chunks))
	})

	t.Run("one to four byte reader", func(t *testing.T) {
		var i int64
		for i = 1; i < 5; i++ {
			chunks := getChunks(NewChunker(bytes.NewReader(randBytes(i)), &Options{AverageSize: 256 * 1024}))
			assert.Equal(t, 1, len(chunks))
		}
	})

	t.Run("avg size is zero", func(t *testing.T) {
		_ = getChunks(NewChunker(
			bytes.NewReader(randBytes(MiB)),
			&Options{AverageSize: 0},
		))
	})

	t.Run("max size is less than avg size", func(t *testing.T) {
		{
			_ = getChunks(NewChunker(
				bytes.NewReader(randBytes(MiB)),
				&Options{AverageSize: 512 * 1024, MaxSize: 511 * 1024},
			))
		}
		{
			_ = getChunks(NewChunker(
				bytes.NewReader(randBytes(MiB)),
				&Options{AverageSize: 512 * 1024, MaxSize: 512 * 1024},
			))
		}
	})

	t.Run("window size << 256", func(t *testing.T) {
		avgSize := (math.E - 1) * 100 // w = 100
		_ = getChunks(NewChunker(bytes.NewReader(randBytes(1024)), &Options{AverageSize: int(avgSize)}))
		// in error case, there will actually be an infinite loop and the test will never finish
	})

	// Legacy test for when max size was an optional setting
	/*t.Run("strictly increasing bytes", func(t *testing.T) {
		data := make([]byte, 260)
		for i := 1; i < 256; i++ {
			data[4+i] = byte(i)
		}

		chunks := getChunks(NewChunker(bytes.NewReader(data), &Options{AverageSize: 10}))
		assert.Len(t, chunks, 1)

		t.Run("maximum chunk size", func(t *testing.T) {
			chunks := getChunks(NewChunker(bytes.NewReader(data), &Options{AverageSize: 10, MaxSize: 100}))
			assert.Len(t, chunks, 3)
		})
	})*/
}
