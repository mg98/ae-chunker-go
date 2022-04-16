package ae

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"io"
	"math"
	"math/rand"
	"testing"
	"time"
)

// MiB represents the number of bytes for 1 mebibyte.
const MiB int64 = 1024 * 1024

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
		chunk, err := c.NextBytes()
		if err == io.EOF {
			break
		}
		chunks = append(chunks, chunk)
	}
	return chunks
}

// testFile is comprised of 100MiB of random bytes.
var testFile = randBytes(100 * MiB)

func TestChunker_NextBytes(t *testing.T) {
	t.Run("sum of chunks is equal original file", func(t *testing.T) {
		const avgSize = 361 * 1024

		t.Run("run with AE_MAX", func(t *testing.T) {
			chunks := getChunks(NewChunker(bytes.NewReader(testFile), &Options{AverageSize: avgSize, Mode: MAX}))
			var data []byte
			for _, chunk := range chunks {
				data = append(data, chunk...)
			}
			assert.Equal(t, testFile, data)
		})

		t.Run("run with AE_MN", func(t *testing.T) {
			c := NewChunker(bytes.NewReader(testFile), &Options{AverageSize: avgSize, Mode: MIN})
			chunks := getChunks(c)
			var data []byte
			for _, chunk := range chunks {
				data = append(data, chunk...)
			}
			assert.Equal(t, testFile, data)
		})
	})

	t.Run("zero byte input", func(t *testing.T) {
		chunks := getChunks(NewChunker(bytes.NewReader([]byte{}), &Options{AverageSize: 256*1024 + 123}))
		assert.Equal(t, 0, len(chunks))
	})

	t.Run("one to four byte input", func(t *testing.T) {
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

	t.Run("strictly increasing bytes", func(t *testing.T) {
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
	})
}

func TestChunker_MinSize(t *testing.T) {
	ch := NewChunker(bytes.NewReader(testFile), &Options{AverageSize: 264*1024+5})
	chunks := getChunks(ch)
	t.Run("minimum chunk size", func(t *testing.T) {
		for _, chunk := range chunks[:len(chunks)-1] {
			assert.GreaterOrEqual(t, len(chunk), ch.MinSize())
		}
		assert.Greater(t, ch.MinSize(), 0)
	})
}

func BenchmarkSplit(b *testing.B) {
	b.Run("window size of 256KiB", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			_ = getChunks(NewChunker(bytes.NewReader(testFile), &Options{AverageSize: 256 * 1024}))
		}
		b.SetBytes(int64(len(testFile)))
		b.ReportAllocs()
	})
	b.Run("window size of 512KiB", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			_ = getChunks(NewChunker(bytes.NewReader(testFile), &Options{AverageSize: 512 * 1024}))
		}
		b.SetBytes(int64(len(testFile)))
		b.ReportAllocs()
	})
	b.Run("window size of 1MiB", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			_ = getChunks(NewChunker(bytes.NewReader(testFile), &Options{AverageSize: 1024 * 1024}))
		}
		b.SetBytes(int64(len(testFile)))
		b.ReportAllocs()
	})
	b.Run("window size of 10MiB", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			_ = getChunks(NewChunker(bytes.NewReader(testFile), &Options{AverageSize: 10 * 1024 * 1024}))
		}
		b.SetBytes(int64(len(testFile)))
		b.ReportAllocs()
	})
}
