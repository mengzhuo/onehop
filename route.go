// Package onehop provides ...
package onehop

import (
	"container/list"
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

	b := new(big.Int)
	b.SetBytes(fullByte)

	block := new(big.Int)
	block.Div(b, big.NewInt(int64(k)))

	l := list.New()

	for i := int64(0); i < int64(k); i++ {

		slice := new(Slice)
		slice.Min = new(big.Int)
		slice.Min.Mul(block, big.NewInt(i))

		slice.Max = new(big.Int)
		slice.Max.Mul(block, big.NewInt(i+1))

		ublock := new(big.Int)
		ublock.Sub(slice.Max, slice.Min)
		ublock.Div(ublock, big.NewInt(int64(k)))

		for j := int64(1); j < int64(u); j++ {
			umin := big.NewInt(0)
			umin.Add(umin, ublock)
			umin.Mul(umin, big.NewInt(j))

			umax := new(big.Int)
			umax.Add(umin, ublock)
			unit := NewUnit(umin, umax)
			slice.units = append(slice.units, unit)
		}
		l.PushBack(slice)
	}

	return &Route{l, k}
}
