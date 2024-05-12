package hyperloglog

type hllVersion interface {
	uint32 | uint64
}

type Sketch[V hllVersion] interface {
	Add(V)
	Cardinality() V
	GetRegisters() []uint8
	GetP() uint8
	SetRegisters([]uint8)
}
