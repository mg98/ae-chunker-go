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
by Yucheng Zhang et al. ([PDF](https://ranger.uta.edu/~jiang/publication/Conferences/2015/2015-INFOCOM-AE-%20An%20Asymmetric%20Extremum%20Content%20Defined%20Chunking%20Algorithm%20for%20Fast%20and%20Bandwidth-Efficient%20Data%20Deduplication.pdfhttps://ranger.uta.edu/~jiang/publication/Conferences/2015/2015-INFOCOM-AE-%20An%20Asymmetric%20Extremum%20Content%20Defined%20Chunking%20Algorithm%20for%20Fast%20and%20Bandwidth-Efficient%20Data%20Deduplication.pdf)).

## Install

```
go get -u github.com/mg98/ae-chunker-go
```

## Example

```go
import (
    "fmt"
    "log"
    "math/rand"
    "time"
)

func main() {
    data := make([]byte, 1024*1024)  // 1 MiB
    rnd := rand.New(rand.NewSource(time.Now().Unix()))
    if _, err := rnd.Read(data); err != nil {
        log.Fatal(err)
    }

    chunks, err := Split(data, 256*1024) // chunk to have an average size of 256 KiB
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf(
        "Data divided into %d chunks. First chunk is %d bytes.\n",
        len(chunks),
        len(chunks[0]),
    )
    // Example output: Data divided into 5 chunks. First chunk is 224098 bytes.
}
```

## Performance

In the results of a benchmark test, **ae-chunker-go** has come out as the fastest content defined chunking algorithm!

The task was to divide 100 MiB of random bytes into chunks with an average size of 256 KiB
(CPU: _Intel(R) Core(TM) i7-8750H CPU @ 2.20GHz_).

| Chunking Algorithm           | Speed         |
|------------------------------|---------------|
| ae-chunker-go                | 0.08159 ns/op |
| [fastcdc-go](https://github.com/jotfs/fastcdc-go)                   | 0.09423 ns/op |
| [go-ipfs-chunker](https://github.com/ipfs/go-ipfs-chunker) (Rabin)      | 0.43320 ns/op  |
| [go-ipfs-chunker](https://github.com/ipfs/go-ipfs-chunker) (Buzhash)    | 0.08248 ns/op |
| [go-ipfs-chunker](https://github.com/ipfs/go-ipfs-chunker) (Fixed Size) | 0.01966 ns/op |

