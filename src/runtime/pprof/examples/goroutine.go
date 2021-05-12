
// +build ignore

package main

import (
	"log"
	"os"
	"runtime/pprof"
	"sync"
)

var wg sync.WaitGroup

//go:noinline
func f1() int {
	defer wg.Done()

	var sum int
	for i := 0; i < 500000000; i++ {
		sum -= i / 2
		sum *= i
		sum /= i/3 + 1
		sum -= i / 4
	}

	return sum
}

//go:noinline
func f2() int {
	defer wg.Done()

	var sum int
	for i := 0; i < 500000000; i++ {
		sum -= i / 2
		sum *= i
		sum /= i/3 + 1
		sum -= i / 4
	}

	return sum
}

//go:noinline
func f3() int {
	defer wg.Done()

	var sum int
	for i := 0; i < 500000000; i++ {
		sum -= i / 2
		sum *= i
		sum /= i/3 + 1
		sum -= i / 4
	}

	return sum
}

//go:noinline
func f4() int {
	defer wg.Done()

	var sum int
	for i := 0; i < 500000000; i++ {
		sum -= i / 2
		sum *= i
		sum /= i/3 + 1
		sum -= i / 4
	}

	return sum
}

//go:noinline
func f5() int {
	defer wg.Done()

	var sum int
	for i := 0; i < 500000000; i++ {
		sum -= i / 2
		sum *= i
		sum /= i/3 + 1
		sum -= i / 4
	}

	return sum
}

//go:noinline
func f6() int {
	defer wg.Done()

	var sum int
	for i := 0; i < 500000000; i++ {
		sum -= i / 2
		sum *= i
		sum /= i/3 + 1
		sum -= i / 4
	}

	return sum
}

//go:noinline
func f7() int {
	defer wg.Done()

	var sum int
	for i := 0; i < 500000000; i++ {
		sum -= i / 2
		sum *= i
		sum /= i/3 + 1
		sum -= i / 4
	}

	return sum
}

//go:noinline
func f8() int {
	defer wg.Done()

	var sum int
	for i := 0; i < 500000000; i++ {
		sum -= i / 2
		sum *= i
		sum /= i/3 + 1
		sum -= i / 4
	}

	return sum
}

//go:noinline
func f9() int {
	defer wg.Done()

	var sum int
	for i := 0; i < 500000000; i++ {
		sum -= i / 2
		sum *= i
		sum /= i/3 + 1
		sum -= i / 4
	}

	return sum
}

//go:noinline
func f10() int {
	defer wg.Done()

	var sum int
	for i := 0; i < 500000000; i++ {
		sum -= i / 2
		sum *= i
		sum /= i/3 + 1
		sum -= i / 4
	}

	return sum
}

func run() error {
	wg.Add(10)
	defer wg.Wait()

	go f1()
	go f2()
	go f3()
	go f4()
	go f5()
	go f6()
	go f7()
	go f8()
	go f9()
	go f10()

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
