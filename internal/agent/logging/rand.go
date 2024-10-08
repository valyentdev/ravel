// MIT License

// Copyright (c) 2020 Milan Pavlik

package logging

import "math/rand"

const (
	lowercase = "abcdefghijklmnopqrstuvwxyz"
	uppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digits    = "0123456789"

	chars = lowercase + uppercase + digits
)

const (
	letterIdxBits = 6                        // 6 bits to represent a letter index
	letterIdxMask = (1 << letterIdxBits) - 1 // All 1-bits, as many as letterIdxBits
)

// https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
func RandomHash(length int) string {
	b := make([]byte, length)
	for i := 0; i < length; {
		if idx := int(rand.Int63() & letterIdxMask); idx < len(chars) {
			b[i] = chars[idx]
			i++
		}
	}
	return string(b)
}
