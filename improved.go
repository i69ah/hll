package hyperloglog

import (
	"math"
)

type ImprovedSketch[V hllVersion] struct {
	p         uint8   // точность / количество первых бит для адреса регистров
	m         V       // количество регистров
	registers []uint8 // массив регистров
	q         uint8   // количество последних бит используемых для подсчёта ведущих нулей
	qMask     V       // битовая маска для последних q бит
}

func NewImprovedSketch[V hllVersion](p uint8) Sketch[V] {
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

	return &ImprovedSketch[V]{
		m:         V(m),
		registers: make([]uint8, m),
		p:         p,
		qMask:     V(1<<q - 1),
		q:         q,
	}
}

func (is *ImprovedSketch[V]) Add(value V) {
	idx := value >> is.q
	rank := countLeadingZeros[V](value&is.qMask, is.p)

	if rank > is.registers[idx] {
		is.registers[idx] = rank
	}
}

func (is *ImprovedSketch[V]) getMultiplicityVector() []V {
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

	c := make([]V, bits-is.p+2)
	for i := V(0); i < is.m; i++ {
		c[is.registers[i]]++
	}

	return c
}

func (is *ImprovedSketch[V]) getPByMultiplicityVector() float64 {
	sum := V(0)
	for _, ci := range is.getMultiplicityVector() {
		sum += ci
	}

	return math.Log2(float64(sum))
}

func (is *ImprovedSketch[V]) Cardinality() V {
	c := is.getMultiplicityVector()
	z := float64(is.m) * tau(1.-float64(c[len(c)-1])/float64(is.m))
	for k := is.q; k > 0; k-- {
		z = .5 * (z + float64(c[k]))
	}

	z = z + float64(is.m)*sigma(float64(c[0])/float64(is.m))

	aInf := 1 / (2 * math.Log(2))

	return V(math.Round(aInf * math.Pow(float64(is.m), 2) / z))
}

func (is *ImprovedSketch[V]) GetRegisters() []uint8 {
	return is.registers
}

func (is *ImprovedSketch[V]) GetP() uint8 {
	return is.p
}
func (is *ImprovedSketch[V]) SetRegisters(registers []uint8) {
	is.registers = registers
}
