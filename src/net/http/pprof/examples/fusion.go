
// +build ignore

package main

import (
	"log"
	_ "math/rand"
	"net/http"
	_ "net/http/pprof"
)

const N = (1 << 20)

var a []float32
var b []float32
var c []float32
var d []float32

func init() {
	a = make([]float32, N)
	b = make([]float32, N)
	c = make([]float32, N)
	d = make([]float32, N)
	//	for i := 0; i < N; i++ {
	//		a[i] = rand.Float32()
	//		b[i] = rand.Float32()
	//	}
}

//go:noinline
func addSub() {
	for i := 0; i < N; i++ {
		c[i] = b[i] + c[i]
	}
	for i := 0; i < N; i++ {
		d[i] = b[i] - c[i]
	}
}

//go:noinline
func addSubFuse() {
	for i := 0; i < N; i++ {
		c[i] = b[i] + c[i]
		d[i] = b[i] - c[i]
	}
}

func run() error {
	addSub()
	addSubFuse()
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
