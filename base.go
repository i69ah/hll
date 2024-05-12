package hyperloglog

import "math"

type BaseSketch[V hllVersion] struct {
	p                     uint8   // точность / количество первых бит для адреса регистров
	m                     V       // количество регистров
	registers             []uint8 // массив регистров
	q                     uint8   // количество последних бит используемых для подсчёта ведущих нулей
	qMask                 V       // битовая маска для последних q бит
	alphaM2               float64 // коэффициент для рассчета количества уникальных элементов (a*m*m)
	lowEstimationBound    float64 // верхняя граница для "малых" кардинальностей
	middleEstimationBound float64 // верхняя граница для "средних" кардинальностей
}

func NewBaseSketch[V hllVersion](p uint8) Sketch[V] {
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
	alpha := .0

	switch p {
	case 4:
		alpha = alpha16
	case 5:
		alpha = alpha32
	case 6:
		alpha = alpha64
	default:
		alpha = .7213 / (1.0 + 1.079/float64(m))
	}

	return &BaseSketch[V]{
		m:                     V(m),
		registers:             make([]uint8, m),
		p:                     p,
		qMask:                 V(1<<q - 1),
		q:                     q,
		alphaM2:               alpha * float64(m) * float64(m),
		lowEstimationBound:    2.5 * float64(m),
		middleEstimationBound: twoPowOf32 / 30,
	}
}

func (bs *BaseSketch[V]) Add(value V) {
	idx := value >> bs.q
	rank := countLeadingZeros[V](value&bs.qMask, bs.p)

	if rank > bs.registers[idx] {
		bs.registers[idx] = rank
	}
}

func (bs *BaseSketch[V]) Cardinality() V {
	sum := .0
	m := float64(bs.m)
	emptyRegistersCount := 0

	for _, register := range bs.registers {
		if register == 0 {
			emptyRegistersCount++
			sum += 1.
		} else {
			sum += 1 / float64(int(1)<<register)
		}
	}

	estimate := bs.alphaM2 / sum

	if estimate <= bs.lowEstimationBound {
		if emptyRegistersCount != 0 {
			return V(m * math.Log(m/float64(emptyRegistersCount)))
		}

		return V(estimate)
	}

	if estimate <= bs.middleEstimationBound {
		return V(estimate)
	}

	return V(-twoPowOf32 * math.Log(1.-(estimate/twoPowOf32)))
}

func (bs *BaseSketch[V]) GetRegisters() []uint8 {
	return bs.registers
}

func (bs *BaseSketch[V]) GetP() uint8 {
	return bs.p
}

func (bs *BaseSketch[V]) SetRegisters(registers []uint8) {
	bs.registers = registers
}
