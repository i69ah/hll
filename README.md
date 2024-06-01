# HyperLogLog algorithms package
***


## Installation
```shell
go get github.com/i69ah/hll
```

## Usage
1. Import package to your project
```go
package main

import (
    "github.com/i69ah/hll"
)
```
2. Create Sketch instance using one of three factory methods
```go
baseSketch := hll.NewBaseSketch[uint32](12)
improvedSketch := hll.NewImprovedSketch[uint32](12)
maximumLikelihoodSketch := hll.NewMaximumLikelihoodSketch[uint32](12)
```
Each sketch implements Sketch interface:
```go
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
```

## Examples

```go
package main

import (
    "fmt"
    "github.com/i69ah/hll"
)

func main() {
    baseSketch := hll.NewBaseSketch[uint32](12)

    for i := 0; i < 1_000_000; i++ {
        baseSketch.Add(uint32(i))
    }

    fmt.Printf("Estimated cardinality - %v", baseSketch.Cardinality())
}
```