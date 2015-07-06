package onehop

import (
	"math/big"
	"testing"
)

func NewSlice() *Slice {

	block := big.NewInt(int64(64))

	slice := new(Slice)
	slice.Max = new(big.Int)
	slice.Min = new(big.Int)
	slice.Min.Add(block, slice.Min)
	slice.Max.Mul(block, big.NewInt(int64(8)))
	slice.Max.Add(slice.Max, slice.Min)
	slice.units = make([]*Unit, 0)

	for j := int64(0); j < int64(8); j++ {
		// Sorted units
		unit := new(Unit)
		unit.Min = new(big.Int)
		unit.Max = new(big.Int)

		unit.Min.Mul(block, big.NewInt(j))
		unit.Min.Add(unit.Min, slice.Min)

		unit.Max.Add(unit.Min, block)

		slice.units = append(slice.units, unit)
	}

	return slice
}

func TestGetUnitIndex(t *testing.T) {

	s := NewSlice()
	idx := s.GetUnitIndex(big.NewInt(int64(63)))
	if 8 != idx {
		t.Errorf("Error on Get Not existed %d", idx)
	}

	idx = s.GetUnitIndex(big.NewInt(int64(384)))
	if 4 != idx {
		t.Errorf("Error on Get Min:%d", idx)
	}

	idx = s.GetUnitIndex(big.NewInt(int64(385)))
	if 5 != idx {
		t.Errorf("Error on Get Inside:%d", idx)
	}
}
