package hyperloglog

import (
	"math"
)

type MaximumLikelihoodSketch[V hllVersion] struct {
	p         uint8   // точность / количество первых бит для адреса регистров
	m         V       // количество регистров
	registers []uint8 // массив регистров
	q         uint8   // количество последних бит используемых для подсчёта ведущих нулей
	qMask     V       // битовая маска для последних q бит
}

func NewMaximumLikelihoodSketch[V hllVersion](p uint8) Sketch[V] {
	var version V
	var q uint8

	switch any(version).(type) {
	case uint32:
		if p > 32 {
			panic("can not use p > 32")
		}
		q = 32 - p
	case uint64:
		if p > 64 {
			panic("can not use p > 64")
		}
		q = 64 - p
	}

	m := 1 << p

	return &MaximumLikelihoodSketch[V]{
		m:         V(m),
		registers: make([]uint8, m),
		p:         p,
		qMask:     V(1<<q - 1),
		q:         q,
	}
}

func (mls *MaximumLikelihoodSketch[V]) Add(value V) {
	idx := value >> mls.q
	rank := countLeadingZeros[V](value&mls.qMask, mls.p)

	if rank > mls.registers[idx] {
		mls.registers[idx] = rank
	}
}

func (mls *MaximumLikelihoodSketch[V]) getMultiplicityVector() []V {
	var sketchVersion V
	var bits uint8

	switch any(sketchVersion).(type) {
	case uint32:
		bits = 32
	case uint64:
		bits = 64
	default:
		panic("unsupported sketch version")
	}

	c := make([]V, bits-mls.p+2)
	for i := V(0); i < mls.m; i++ {
		c[mls.registers[i]]++
	}

	return c
}

func (mls *MaximumLikelihoodSketch[V]) Cardinality() V {
	q := mls.q
	C := mls.getMultiplicityVector()

	if C[q+1] == mls.m {
		var res V
		switch any(res).(type) {
		case uint32:
			return V(1)<<32 - 1
		case uint64:
			return V(1)<<64 - 1
		default:
			panic("unsupported sketch version")
		}
	}

	Kmin := mls.getKmin(C)
	Kmax := mls.getKmax(C)

	z := float64(0)
	for k := Kmax; k >= Kmin; k-- {
		z = z/2 + float64(C[k])
	}

	z = z / float64(int(1)<<Kmin)

	c := C[q+1]
	if q >= 1 {
		c = c + C[Kmax]
	}

	gPrev := .0
	a := z + float64(C[0])
	b := z + float64(C[q+1])/float64(uint64(1)<<q)
	m1 := float64(mls.m - C[0])
	var x float64

	if b <= 1.5*a {
		x = m1 / (b/2 + a)
	} else {
		x = (m1 / b) * math.Log(1+b/a)
	}

	deltaX := x
	sigmaVar := .01 / math.Sqrt(float64(mls.m))

	for deltaX > x*sigmaVar {
		ks := float64(2) + math.Floor(math.Log2(x))
		degree := float64(Kmax)
		if ks > degree {
			degree = ks
		}

		x1 := x / float64(int(1)<<int(degree+1))
		x11 := x1 * x1
		h := x1 - x11/3 + (x11*x11)*(1/45-x11/472.5)
		for k := ks - 1; k >= float64(Kmax); k -= 1. {
			h = (x1 + h*(1-h)) / (x1 + 1 - h)
			x1 *= 2
		}

		g := float64(c) * h
		for k := Kmax - 1; k >= Kmin; k-- {
			h = (x1 + h*(1-h)) / (x1 + 1 - h)
			g += float64(C[k]) * h
			x1 *= 2
		}

		g += x * a
		if g > gPrev && m1 >= g {
			deltaX = deltaX * (m1 - g) / (g - gPrev)
		} else {
			deltaX = 0
		}

		x = x + deltaX
		gPrev = g
	}

	return V(math.Round(float64(mls.m) * x))
}

func (mls *MaximumLikelihoodSketch[V]) getKmin(c []V) uint8 {
	for k, Ck := range c {
		if Ck > 0 {
			if k > 1 {
				return uint8(k)
			}

			return 1
		}
	}

	return 1
}

func (mls *MaximumLikelihoodSketch[V]) getKmax(c []V) uint8 {
	Kmax := uint8(0)
	for k, Ck := range c {
		if Ck > 0 {
			Kmax = uint8(k)
		}
	}

	if Kmax > mls.q {
		return mls.q
	}

	return Kmax
}

func (mls *MaximumLikelihoodSketch[V]) GetRegisters() []uint8 {
	return mls.registers
}

func (mls *MaximumLikelihoodSketch[V]) GetP() uint8 {
	return mls.p
}

func (mls *MaximumLikelihoodSketch[V]) SetRegisters(registers []uint8) {
	mls.registers = registers
}
