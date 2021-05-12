
// +build ignore

package main

import (
	"log"
	"math/rand"
	"os"
	"runtime/pprof"
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
	for i := 0; i < N; i++ {
		a[i] = rand.Float32()
		b[i] = rand.Float32()
	}
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
	cwd, _ := os.Getwd()
	prefix := cwd + "/" + os.Args[0]
	cycleFile, err := os.Create(prefix + ".cycle.prof")
	if err != nil {
		log.Fatal(err)
	}
	defer cycleFile.Close()
	cacheRefFile, err := os.Create(prefix + ".cacheRef.prof")
	if err != nil {
		log.Fatal(err)
	}
	defer cacheRefFile.Close()
	cacheMissFile, err := os.Create(prefix + ".cacheMiss.prof")
	if err != nil {
		log.Fatal(err)
	}
	defer cacheMissFile.Close()

	if err := pprof.StartCPUProfileWithConfig(pprof.CPUCycles(cycleFile, 10000000), pprof.CPUCacheReferences(cacheRefFile, 10000), pprof.CPUCacheMisses(cacheMissFile, 10000)); err != nil {
		log.Fatal(err)
	}
	defer pprof.StopCPUProfile()
	for i := 0; i < 100; i++ {
		if err := run(); err != nil {
			log.Fatal(err)
		}
	}
}
