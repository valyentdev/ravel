package id

import (
	"crypto/rand"
	"fmt"
	"math"

	"github.com/nrednav/cuid2"
)

func init() {
	generate, _ = cuid2.Init(cuid2.WithRandomFunc(RandomFloat))
}

var generate func() string

func GeneratePrefixed(prefix string) string {
	return fmt.Sprintf("%s_%s", prefix, Generate())
}

func Generate() string {
	return generate()
}

func RandomFloat() float64 {
	var randomBytes [8]byte
	_, err := rand.Read(randomBytes[:])
	if err != nil {
		panic(err)
	}

	float := float64(uint64(randomBytes[0])<<56 | uint64(randomBytes[1])<<48 | uint64(randomBytes[2])<<40 | uint64(randomBytes[3])<<32 | uint64(randomBytes[4])<<24 | uint64(randomBytes[5])<<16 | uint64(randomBytes[6])<<8 | uint64(randomBytes[7]))
	return math.Abs(float / math.MaxUint64)
}
