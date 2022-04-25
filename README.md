# AE Chunker (GO)

[![GoDoc](http://img.shields.io/badge/godoc-reference-blue.svg)](https://pkg.go.dev/github.com/mg98/ae-chunker-go)
[![Test](https://github.com/mg98/ae-chunker-go/actions/workflows/test.yml/badge.svg)](https://github.com/mg98/ae-chunker-go/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/mg98/ae-chunker-go/branch/main/graph/badge.svg?token=R3OYXX1HC7)](https://codecov.io/gh/mg98/ae-chunker-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/mg98/ae-chunker-go?)](https://goreportcard.com/report/github.com/mg98/ae-chunker-go)
![License](https://img.shields.io/github/license/mg98/ae-chunker-go)

**ae-chunker-go** is a best-effort Go implementation of the chunking algorithm presented in
_AE: An Asymmetric Extremum Content Defined
Chunking Algorithm for Fast and
Bandwidth-Efficient Data Deduplication_
by Yucheng Zhang et al. ([PDF](https://ranger.uta.edu/~jiang/publication/Conferences/2015/2015-INFOCOM-AE-%20An%20Asymmetric%20Extremum%20Content%20Defined%20Chunking%20Algorithm%20for%20Fast%20and%20Bandwidth-Efficient%20Data%20Deduplication.pdf)).

## Install

```
go get -u github.com/mg98/ae-chunker-go
```

## Example

```go
import (
    "bytes"
    "fmt"
    "io"
    "log"
    "math/rand"
    "time"
    "github.com/ae-chunker-go"
)

func main() {
    data := make([]byte, 1024*1024)  // 1 MiB
    rnd := rand.New(rand.NewSource(time.Now().Unix()))
    if _, err := rnd.Read(data); err != nil {
        log.Fatal(err)
    }

    chunker := ae.NewChunker(bytes.NewReader(data), &ae.Options{
    	AverageSize: 256*1024,  // 256 KiB
    	MaxSize: 512*1024,      // 512 KiB
    })
    var chunks [][]byte
    for {
    	chunk, err := chunker.NextBytes()
    	if err == io.EOF {
    		break
        } else if err != nil {
        	log.Fatal(err)
        }
        chunks = append(chunks, chunk)
    }
    
    fmt.Printf(
        "Data divided into %d chunks. First chunk is %d bytes.\n",
        len(chunks),
        len(chunks[0]),
    )
    // Example output: Data divided into 5 chunks. First chunk is 224098 bytes.
}
```

## Benchmarks

### Performance

The task was to divide 100 MiB of random bytes into chunks with an average size of 256 KiB
(CPU: _Intel(R) Core(TM) i7-8750H CPU @ 2.20GHz_).

| Chunking Algorithm                                                      |      Speed | Processed Bytes | Allocated Bytes | Distinct Mem. Alloc. |
|-------------------------------------------------------------------------|-----------:|----------------:|----------------:|---------------------:|
| ae-chunker-go                                                           | 168 sec/op |     622.94 MB/s |    507.27 MB/op |       7769 allocs/op |
| [fastcdc-go](https://github.com/jotfs/fastcdc-go)                       |  88 sec/op |    1194.74 MB/s |      2.10 MB/op |          3 allocs/op |
| [go-ipfs-chunker](https://github.com/ipfs/go-ipfs-chunker) (Rabin)      | 414 sec/op |     253.54 MB/s |    108.83 MB/op |       1192 allocs/op |
| [go-ipfs-chunker](https://github.com/ipfs/go-ipfs-chunker) (Buzhash)    |  81 sec/op |    1288.27 MB/s |    106.48 MB/op |        406 allocs/op |
| [go-ipfs-chunker](https://github.com/ipfs/go-ipfs-chunker) (Fixed Size) |  22 sec/op |    4773.13 MB/s |    104.87 MB/op |        405 allocs/op |

### Deduplication Efficiency

This metric measures how well deduplication performs with multiple versions of a file. 
More precisely, we define the _Deduplication Elimination Ratio (DER)_ as the ratio of the size of the input data
to the size of the altered data (the higher the better).

For the evaluation, the uncompressed TAR archives of 20 consecutive versions of the GCC source code were used
(a total of 12 GB). The algorithms were run configured for an average chunk size 
of 8 KB and (where applicable) a maximum chunk size of 16 KB. 
Because the Buzhash library does not support flexible chunk sizes 
the tests were repeated with 256 KB average and 512 KB max size for better comparison.
Generally, smaller chunk sizes make better deduplication.

| Chunking Algorithm                                                      | DER (8K/16K) | DER (256K/512K) |
|-------------------------------------------------------------------------|-------------:|----------------:|
| ae-chunker-go                                                           |     1.056510 |        1.002392 |
| [fastcdc-go](https://github.com/jotfs/fastcdc-go)                       |     1.000643 |        1.000000 |
| [go-ipfs-chunker](https://github.com/ipfs/go-ipfs-chunker) (Rabin)      |     1.354034 |        1.058422 |
| [go-ipfs-chunker](https://github.com/ipfs/go-ipfs-chunker) (Buzhash)    |          n/a |        1.083399 |
| [go-ipfs-chunker](https://github.com/ipfs/go-ipfs-chunker) (Fixed Size) |     1.032097 |        1.000579 |


### Chunk Size Variance

The following plots show the chunk size distribution on a set of random bytes of 1 GiB.
The algorithm was run with the options
`&ae.Options{AverageSize: 256*1024}` and `&ae.Options{AverageSize: 256*1024, MaxSize: 512*1024}`,
respectively.

<img src="./img/csd256kib.png" width="49%"> <img src="./img/csd256kib512kib.png" width="49%">
