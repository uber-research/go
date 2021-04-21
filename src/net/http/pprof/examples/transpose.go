// +build ignore

package main

import (
	"log"
	_ "math/rand"
	"net/http"
	_ "net/http/pprof"
)

const N = 16000

var a [N][N]float32

//go:noinline
func transposeTiled() {
	BS := 100
	for ii := 0; ii < N; ii += BS {
		for jj := 0; jj < N; jj += BS {
			for i := ii; i < ii+BS; i++ {
				for j := jj; j < jj+BS; j++ {
					v1 := a[i][j]
					v2 := a[j][i]
					a[i][j] = v2
					a[j][i] = v1
				}
			}
		}
	}
}

//go:noinline
func transpose() {
	for i := 0; i < N; i++ {
		for j := 0; j < N; j++ {
			v1 := a[i][j]
			v2 := a[j][i]
			a[i][j] = v2
			a[j][i] = v1
		}
	}
}

func run() error {
	transpose()
	transposeTiled()
	return nil
}

func main() {
	go func() { log.Println(http.ListenAndServe("localhost:6060", nil)) }()
	for {
		if err := run(); err != nil {
			log.Fatal(err)
		}
	}
}
