package game

import "encoding/hex"

// RNG is a splitmix64 deterministic pseudo-random generator. It is implemented
// here (rather than using the standard library) so the authoritative stream is
// fully specified and stable across Go versions — replay must reproduce the
// exact same draws from the same seed (ADR-0001 D8.1).
type RNG struct {
	state uint64
}

// NewRNG seeds the generator with a uint64.
func NewRNG(seed uint64) *RNG { return &RNG{state: seed} }

// Next returns the next 64-bit value in the stream.
func (r *RNG) Next() uint64 {
	r.state += 0x9e3779b97f4a7c15
	z := r.state
	z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
	z = (z ^ (z >> 27)) * 0x94d049bb133111eb
	return z ^ (z >> 31)
}

// IntN returns a value in [0, n). n must be > 0.
func (r *RNG) IntN(n int) int {
	if n <= 0 {
		return 0
	}
	return int(r.Next() % uint64(n))
}

// SeedToUint64 folds a (hex) seed string deterministically into a uint64 using
// FNV-1a over its bytes. Hex seeds are decoded first so the full entropy of the
// server seed is used; non-hex strings are folded as raw bytes.
func SeedToUint64(seed string) uint64 {
	b, err := hex.DecodeString(seed)
	if err != nil || len(b) == 0 {
		b = []byte(seed)
	}
	var v uint64 = 1469598103934665603 // FNV-1a 64-bit offset basis
	for _, c := range b {
		v ^= uint64(c)
		v *= 1099511628211 // FNV-1a 64-bit prime
	}
	return v
}
