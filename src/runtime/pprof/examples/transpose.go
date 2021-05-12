
// +build ignore

package main

import (
	"log"
	"math/rand"
	"os"
	"runtime/pprof"
)

const N = 16000

var a [N][N]float32

func init() {
	return
	for i := 0; i < N; i++ {
		for j := 0; j < N; j++ {
			a[i][j] = rand.Float32()
		}
	}
}

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
	for i := 0; i < 1; i++ {
		if err := run(); err != nil {
			log.Fatal(err)
		}
	}

}
