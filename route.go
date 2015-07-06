// Package onehop provides ...
package onehop

import (
	"container/list"
	_ "fmt"
	"math/big"
)

type Route struct {
	slices *list.List
	k      int // number of slices the ring is divided into
}

func NewRoute(k int, u int) *Route {

	fullByte := []byte{0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff}

	block := new(big.Int)
	block.SetBytes(fullByte)
	block.Div(block, big.NewInt(int64(k*u)))
	// TODO wired length issues on divied number
	block.Add(block, big.NewInt(1))

	max_num := new(big.Int)
	max_num.SetBytes(fullByte)

	l := list.New()

	for i := int64(0); i < int64(k); i++ {
		slice := new(Slice)
		slice.Max = new(big.Int)
		slice.Min = new(big.Int)
		slice.Min.Mul(block, big.NewInt(i*int64(u)))
		slice.Max.Mul(block, big.NewInt((i+1)*int64(u)))
		if slice.Max.Cmp(max_num) > 0 {
			slice.Max.SetBytes(fullByte)
		} else {
			slice.Max.Sub(slice.Max, big.NewInt(int64(1)))
		}
		l.PushBack(slice)
		slice.units = make([]*Unit, 0)

		for j := int64(0); j < int64(u); j++ {
			// Sorted units
			unit := new(Unit)
			unit.Min = new(big.Int)
			unit.Max = new(big.Int)

			unit.Min.Mul(block, big.NewInt(j))
			unit.Min.Add(unit.Min, slice.Min)

			unit.Max.Add(unit.Min, block)
			if unit.Max.Cmp(max_num) > 0 {
				unit.Max.SetBytes(fullByte)
			} else {
				unit.Max.Sub(unit.Max, big.NewInt(int64(1)))
			}

			slice.units = append(slice.units, unit)
		}
	}

	return &Route{l, k}
}
