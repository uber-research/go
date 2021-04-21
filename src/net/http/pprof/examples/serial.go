// +build ignore

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
)

//go:noinline
func J_expect_18_18(v uint64, trip uint64) uint64 {
	ret := v
	for i := trip; i > 0; i-- {
		ret += i
		ret = ret ^ (i + 0xcafebabe)
	}
	return ret
}

//go:noinline
func I_expect_16_36(v uint64, trip uint64) uint64 {
	ret := v
	for i := trip; i > 0; i-- {
		ret += i
		ret = ret ^ (i + 0xcafebabe)
	}
	return ret
}

//go:noinline
func H_expect_14_546(v uint64, trip uint64) uint64 {
	ret := v
	for i := trip; i > 0; i-- {
		ret += i
		ret = ret ^ (i + 0xcafebabe)
	}
	return ret
}

//go:noinline
func G_expect_12_73(v uint64, trip uint64) uint64 {
	ret := v
	for i := trip; i > 0; i-- {
		ret += i
		ret = ret ^ (i + 0xcafebabe)
	}
	return ret
}

//go:noinline
func F_expect_10_91(v uint64, trip uint64) uint64 {
	ret := v
	for i := trip; i > 0; i-- {
		ret += i
		ret = ret ^ (i + 0xcafebabe)
	}
	return ret
}

//go:noinline
func E_expect_9_09(v uint64, trip uint64) uint64 {
	ret := v
	for i := trip; i > 0; i-- {
		ret += i
		ret = ret ^ (i + 0xcafebabe)
	}
	return ret
}

//go:noinline
func D_expect_7_27(v uint64, trip uint64) uint64 {
	ret := v
	for i := trip; i > 0; i-- {
		ret += i
		ret = ret ^ (i + 0xcafebabe)
	}
	return ret
}

//go:noinline
func C_expect_5_46(v uint64, trip uint64) uint64 {
	ret := v
	for i := trip; i > 0; i-- {
		ret += i
		ret = ret ^ (i + 0xcafebabe)
	}
	return ret
}

//go:noinline
func B_expect_3_64(v uint64, trip uint64) uint64 {
	ret := v
	for i := trip; i > 0; i-- {
		ret += i
		ret = ret ^ (i + 0xcafebabe)
	}
	return ret
}

//go:noinline
func A_expect_1_82(v uint64, trip uint64) uint64 {
	ret := v
	for i := trip; i > 0; i-- {
		ret += i
		ret = ret ^ (i + 0xcafebabe)
	}
	return ret
}

func main() {
	go func() { log.Println(http.ListenAndServe("localhost:6060", nil)) }()

	var q uint64
	multiplier := flag.Uint64("m", 100000, "multiplier")
	flag.Parse()
	fmt.Println("multiplier=", *multiplier)
	mult := *multiplier

	for i := uint64(0); ; i++ {
		f := i + A_expect_1_82(0xebabefac23, 1*mult)
		g := i + B_expect_3_64(f, 2*mult)
		h := i + C_expect_5_46(g, 3*mult)
		k := i + D_expect_7_27(h, 4*mult)
		l := i + E_expect_9_09(k, 5*mult)
		m := i + F_expect_10_91(l, 6*mult)
		n := i + G_expect_12_73(m, 7*mult)
		o := i + H_expect_14_546(n, 8*mult)
		p := i + I_expect_16_36(o, 9*mult)
		q = i + J_expect_18_18(p, 10*mult)
	}
	fmt.Println(q)
}
