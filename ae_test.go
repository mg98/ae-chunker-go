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

		chunks, _ := Split(testFile, avgSize)
		var data []byte
		for _, chunk := range chunks {
			data = append(data, chunk...)
		}
		assert.Equal(t, testFile, data)

		t.Run("minimum chunk size", func(t *testing.T) {
			windowSize := int(math.Round(avgSize / (math.E - 1)))
			for _, chunk := range chunks[:len(chunks)-1] {
				assert.GreaterOrEqual(t, len(chunk), windowSize+1)
			}
		})
	})

	t.Run("zero byte input", func(t *testing.T) {
		chunks, _ := Split([]byte{}, 256*1024+123)
		assert.Equal(t, 0, len(chunks))
	})

	t.Run("one to four byte input", func(t *testing.T) {
		var i int64
		for i = 1; i < 5; i++ {
			chunks, _ := Split(randBytes(i), 256*1024)
			assert.Equal(t, 1, len(chunks))
		}
	})

	t.Run("avg size is too small", func(t *testing.T) {
		_, err := Split(randBytes(int64(units.Megabyte)), 0)
		assert.Error(t, err)
	})

	t.Run("window size << 256", func(t *testing.T) {
		avgSize := (math.E - 1) * 100 // w = 100
		_, err := Split(randBytes(int64(units.Kilobyte)), avgSize)
		assert.Nil(t, err)
		// in error case, there will actually be an infinite loop and the test will never finish
	})
}

func BenchmarkSplit(b *testing.B) {
	b.Run("window size of 256KiB", func(b *testing.B) {
		Split(testFile, 256*1024)
		b.SetBytes(int64(len(testFile)))
	})
	b.Run("window size of 512KiB", func(b *testing.B) {
		Split(testFile, 512*1024)
		b.SetBytes(int64(len(testFile)))
	})
	b.Run("window size of 1MiB", func(b *testing.B) {
		Split(testFile, 1024*1024)
		b.SetBytes(int64(len(testFile)))
	})
	b.Run("window size of 10MiB", func(b *testing.B) {
		Split(testFile, 10*1024*1024)
		b.SetBytes(int64(len(testFile)))
	})
}
