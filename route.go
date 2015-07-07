// Package onehop provides ...
package onehop

import "math/big"

var (
	fullByte = []byte{0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff}
)

type Route struct {
	slices []*Slice
	k      int // number of slices the ring is divided into
	u      int
	block  *big.Int
}

func (r *Route) Len() int {
	return len(r.slices)
}

func (r *Route) GetIndex(id *big.Int) (sliceidx, unitidex int) {

	if id.Int64() == int64(0) {
		return 0, 0
	}

	block_idx := int(new(big.Int).Div(id, r.block).Int64())
	slice_idx := block_idx / r.u

	unitidex = block_idx - slice_idx*r.u

	return slice_idx, unitidex
}

func (r *Route) GetNode(id *big.Int) (n *Node) {

	slice_idx, unit_idx := r.GetIndex(id)

	slice := r.slices[slice_idx]
	unit := slice.units[unit_idx]

	return unit.Get(id)
}

func (r *Route) SuccessorOf(id *big.Int) (n *Node) {

	slice_idx, unit_idx := r.GetIndex(id)

	slice := r.slices[slice_idx]
	unit := slice.units[unit_idx]

	n = unit.Get(id)
	if n != nil {
		return n
	}

	n = unit.SuccessorOf(id)

	if n != nil {
		return n
	}

	start_slice, start_unit := slice_idx, unit_idx

	if unit_idx != r.u-1 {
		unit_idx++
	} else {
		unit_idx = 0
		slice_idx++
	}

	// We don't want recycle
	for unit_idx != start_unit && slice_idx != start_slice {

		slice = r.slices[slice_idx]
		n = slice.units[unit_idx].SuccessorOf(id)
		if n != nil {
			return n
		}
		if unit_idx != r.u-1 {
			unit_idx++
		} else {
			unit_idx = 0
			slice_idx++
		}
	}

	return nil
}

func (r *Route) Add(n *Node) (ok bool) {

	slice_idx, unit_idx := r.GetIndex(n.ID)
	slice := r.slices[slice_idx]
	if slice.Len() == 0 {
		slice.Leader = n
	}

	unit := slice.units[unit_idx]
	unit.Add(n)
	return true
}
func (r *Route) Delete(id string) {

}

func NewRoute(k int, u int) *Route {

	if k < 2 || u < 2 {
		panic("K or U can't not less than 2")
	}

	block := new(big.Int)
	block.SetBytes(fullByte)
	block.Div(block, big.NewInt(int64(k*u)))
	// TODO wired length issues on divied number
	block.Add(block, big.NewInt(1))

	max_num := new(big.Int)
	max_num.SetBytes(fullByte)

	l := make([]*Slice, 0)

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

		l = append(l, slice)

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

	return &Route{l, k, u, block}
}
