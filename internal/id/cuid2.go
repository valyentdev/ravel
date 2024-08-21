package id

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"

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
	random, _ := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	float, _ := random.Float64()

	return float / math.MaxFloat64
}
