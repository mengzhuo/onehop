package onehop

import (
	"fmt"
	"math/big"
)

type Slice struct {
	Leader *Node
	Min    *big.Int
	Max    *big.Int
	units  []*Unit
}

func NewUnit(min, max *big.Int) *Unit {

	fmt.Printf("Min  %x\n", min.Bytes())
	fmt.Printf("max %x\n", max.Bytes())
	l := make([]*Node, 0)
	return &Unit{Min: min, Max: max, NodeList: l}
}

type Unit struct {
	Leader *Node
	Min    *big.Int
	Max    *big.Int

	NodeList []*Node
}
