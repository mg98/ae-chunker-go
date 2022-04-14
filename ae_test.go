package ae

import (
	"github.com/alecthomas/units"
	"github.com/stretchr/testify/assert"
	"math"
	"math/rand"
	"testing"
	"time"
)

// randBytes returns a random sequence of n bytes.
func randBytes(n int64) []byte {
	b := make([]byte, n)
	rnd := rand.New(rand.NewSource(time.Now().Unix()))
	if _, err := rnd.Read(b); err != nil {
		panic(err)
	}
	return b
}

// testFile is comprised of 100MiB of random bytes.
var testFile = randBytes(int64(100 * units.Mebibyte))

func TestSplit(t *testing.T) {
	t.Run("sum of chunks is equal original file", func(t *testing.T) {
		const avgSize = 361 * 1024
		var chunks [][]byte

		t.Run("run with AE_MAX", func(t *testing.T) {
			c := NewChunker(testFile, &Options{AverageSize: avgSize, Mode: MAX})
			chunks, _ = c.Split()
			var data []byte
			for _, chunk := range chunks {
				data = append(data, chunk...)
			}
			assert.Equal(t, testFile, data)
		})
		t.Run("run with AE_MN", func(t *testing.T) {
			c := NewChunker(testFile, &Options{AverageSize: avgSize, Mode: MIN})
			chunks, _ = c.Split()
			var data []byte
			for _, chunk := range chunks {
				data = append(data, chunk...)
			}
			assert.Equal(t, testFile, data)
		})

		t.Run("minimum chunk size", func(t *testing.T) {
			windowSize := int(math.Round(avgSize / (math.E - 1)))
			for _, chunk := range chunks[:len(chunks)-1] {
				assert.GreaterOrEqual(t, len(chunk), windowSize+1)
			}
		})
	})

	t.Run("zero byte input", func(t *testing.T) {
		c := NewChunker([]byte{}, &Options{AverageSize: 256*1024+123})
		chunks, _ := c.Split()
		assert.Equal(t, 0, len(chunks))
	})

	t.Run("one to four byte input", func(t *testing.T) {
		var i int64
		for i = 1; i < 5; i++ {
			c := NewChunker(randBytes(i), &Options{AverageSize: 256*1024})
			chunks, _ := c.Split()
			assert.Equal(t, 1, len(chunks))
		}
	})

	t.Run("avg size is too small", func(t *testing.T) {
		c := NewChunker(randBytes(int64(units.Megabyte)), &Options{AverageSize: 0})
		_, err := c.Split()
		assert.Error(t, err)
	})

	t.Run("max size is less than avg size", func(t *testing.T) {
		{
			c := NewChunker(randBytes(int64(units.Megabyte)), &Options{AverageSize: 512 * 1024, MaxSize: 511 * 1024})
			_, err := c.Split()
			assert.Error(t, err)
		}
		{
			c := NewChunker(randBytes(int64(units.Megabyte)), &Options{AverageSize: 512 * 1024, MaxSize: 512 * 1024})
			_, err := c.Split()
			assert.Nil(t, err)
		}
	})

	t.Run("window size << 256", func(t *testing.T) {
		avgSize := (math.E - 1) * 100 // w = 100
		c := NewChunker(randBytes(int64(units.Kilobyte)), &Options{AverageSize: int(avgSize)})
		_, err := c.Split()
		assert.Nil(t, err)
		// in error case, there will actually be an infinite loop and the test will never finish
	})

	t.Run("strictly increasing bytes", func(t *testing.T) {
		data := make([]byte, 260)
		for i := 1; i < 256; i++ {
			data[4+i] = byte(i)
		}

		{
			c := NewChunker(data, &Options{AverageSize: 10})
			chunks, _ := c.Split()
			assert.Len(t, chunks, 1)
		}

		t.Run("maximum chunk size", func(t *testing.T) {
			c := NewChunker(data, &Options{AverageSize: 10, MaxSize: 100})
			chunks, _ := c.Split()
			assert.Len(t, chunks, 3)
		})
	})
}

func BenchmarkSplit(b *testing.B) {
	b.Run("window size of 256KiB", func(b *testing.B) {
		c := NewChunker(testFile, &Options{AverageSize: 256*1024})
		_, _ = c.Split()
		b.SetBytes(int64(len(testFile)))
	})
	b.Run("window size of 512KiB", func(b *testing.B) {
		c := NewChunker(testFile, &Options{AverageSize: 512*1024})
		_, _ = c.Split()
		b.SetBytes(int64(len(testFile)))
	})
	b.Run("window size of 1MiB", func(b *testing.B) {
		c := NewChunker(testFile, &Options{AverageSize: 1024*1024})
		_, _ = c.Split()
		b.SetBytes(int64(len(testFile)))
	})
	b.Run("window size of 10MiB", func(b *testing.B) {
		c := NewChunker(testFile, &Options{AverageSize: 10*1024*1024})
		_, _ = c.Split()
		b.SetBytes(int64(len(testFile)))
	})
}
