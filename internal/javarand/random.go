package javarand

import "math"

const (
	multiplier = uint64(0x5DEECE66D)
	addend     = uint64(0xB)
	mask       = (uint64(1) << 48) - 1
)

// Random replicates java.util.Random for deterministic parity.
type Random struct {
	seed                 uint64
	haveNextNextGaussian bool
	nextNextGaussian     float64
}

func New(seed int64) *Random {
	r := &Random{}
	r.SetSeed(seed)
	return r
}

func (r *Random) SetSeed(seed int64) {
	r.seed = (uint64(seed) ^ multiplier) & mask
	r.haveNextNextGaussian = false
}

func (r *Random) next(bits int) int32 {
	r.seed = (r.seed*multiplier + addend) & mask
	return int32(r.seed >> (48 - bits))
}

func (r *Random) Float64() float64 {
	a := int64(r.next(26))
	b := int64(r.next(27))
	return float64((a<<27)+b) / (1 << 53)
}

func (r *Random) Intn(n int) int {
	if n <= 0 {
		panic("n must be positive")
	}
	if n&(n-1) == 0 {
		return int((int64(n) * int64(r.next(31))) >> 31)
	}
	for {
		bits := int(r.next(31))
		val := bits % n
		if bits-val+(n-1) >= 0 {
			return val
		}
	}
}

func (r *Random) NormFloat64() float64 {
	if r.haveNextNextGaussian {
		r.haveNextNextGaussian = false
		return r.nextNextGaussian
	}
	var v1, v2, s float64
	for {
		v1 = 2*r.Float64() - 1
		v2 = 2*r.Float64() - 1
		s = v1*v1 + v2*v2
		if s < 1 && s != 0 {
			break
		}
	}
	mult := math.Sqrt(-2 * math.Log(s) / s)
	r.nextNextGaussian = v2 * mult
	r.haveNextNextGaussian = true
	return v1 * mult
}

func (r *Random) Shuffle(n int, swap func(i, j int)) {
	for i := n; i > 1; i-- {
		j := r.Intn(i)
		swap(i-1, j)
	}
}
