package main

import (
	"io"
	"io/ioutil"
	"math/rand"
	"testing"

	rand7i "github.com/7i/rand"
	mewrand "github.com/sanctuary/d1/rand"
	"github.com/seehuhn/mt19937"
)

const toCopy = 1024 * 1024

func BenchmarkRandbo(b *testing.B) {
	b.SetBytes(toCopy)
	r := New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		io.CopyN(ioutil.Discard, r, toCopy)
	}
}

func BenchmarkRand(b *testing.B) {
	var rprim = rand.New(rand.NewSource(0))
	b.SetBytes(toCopy)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		io.CopyN(ioutil.Discard, rprim, toCopy)
	}
}

func BenchmarkMewRand(b *testing.B) {
	b.SetBytes(toCopy)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < toCopy; j += 4 {
			mewrand.Int()
		}
	}
}

func BenchmarkMt(b *testing.B) {
	mt := mt19937.New()
	b.SetBytes(toCopy)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		io.CopyN(ioutil.Discard, mt, toCopy)
	}
}

func Benchmark7iComplex128(b *testing.B) {
	rng := rand7i.NewComplexRNG(0)
	b.SetBytes(toCopy)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < toCopy; j += 16 {
			rng.Complex128()
		}
	}
}

func Benchmark7iComplex128Go(b *testing.B) {
	rng := rand7i.NewComplexRNG(0)
	b.SetBytes(toCopy)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < toCopy; j += 16 {
			rng.Complex128Go()
		}
	}
}
