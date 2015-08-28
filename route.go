// Package onehop provides ...
package onehop

import (
	"math"
	"math/big"
	"strconv"

	"github.com/golang/glog"
)

const (
	ZERO_ID = "00000000000000000000000000000000"
	FULL_ID = "ffffffffffffffffffffffffffffffff"
)

var (
	FullID = []byte{0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff}
	zeroID = big.NewInt(int64(0))
)

type Route struct {
	slices []*Slice
	k      int // number of slices the ring is divided into
	div    int
	block  int
}

func (r *Route) Len() int {
	count := 0
	for _, s := range r.slices {
		count += s.Len()
	}
	return count
}

func (r *Route) GetIndex(id string) (sliceidx int) {

	idx, err := strconv.ParseUint(id[:r.div], 16, 0)
	if err != nil {
		glog.Error(err)
	}
	if id[:r.div] != FULL_ID[:r.div] {
		return int(idx) / r.block
	} else {
		return len(r.slices) - 1
	}
}

func (r *Route) GetSlice(id string) (slice *Slice) {

	slice_idx := r.GetIndex(id)
	slice = r.slices[slice_idx]
	return
}

func (r *Route) GetNode(id string) (n *Node) {

	slice_idx := r.GetIndex(id)
	slice := r.slices[slice_idx]

	return slice.Get(id)
}

func (r *Route) SuccessorOf(id string) (n *Node) {

	slice_idx := r.GetIndex(id)
	for i := 0; i < r.k; i++ {
		if i != 0 {
			id = ZERO_ID
		}
		slice := r.slices[slice_idx]
		n = slice.successorOf(id)
		if n != nil {
			return
		}
		slice_idx = (slice_idx + 1) % r.k
	}
	return
}

func (r *Route) Add(n *Node) (ok bool) {

	slice_idx := r.GetIndex(n.ID)
	slice := r.slices[slice_idx]
	result := slice.Add(n)

	return result
}
func (r *Route) Delete(id string) (ok bool) {

	slice_idx := r.GetIndex(id)
	slice := r.slices[slice_idx]

	result := slice.Delete(id)
	return result
}

func NewRoute(k int) *Route {

	if k < 2 {
		panic("K  can't not less than 2")
	}
	div := int(math.Ceil(math.Log2(float64(k))/8)) * 2
	glog.Infof("starting route k=%d", k)
	block := new(big.Int)
	block.SetBytes(FullID)

	block.Div(block, big.NewInt(int64(k)))
	block.Add(block, big.NewInt(1))

	max_num := new(big.Int)
	max_num.SetBytes(FullID)

	l := make([]*Slice, 0)

	for i := int64(0); i < int64(k); i++ {
		max := new(big.Int)
		min := new(big.Int)
		min.Mul(block, big.NewInt(i))
		max.Mul(block, big.NewInt((i + 1)))
		if max.Cmp(max_num) > 0 {
			max.SetBytes(FullID)
		} else {
			max.Sub(max, big.NewInt(int64(1)))
		}

		slice := NewSlice(
			BytesToId(min.Bytes()),
			BytesToId(max.Bytes()))
		l = append(l, slice)
	}

	r := &Route{l, k, div,
		int(math.Pow(float64(16), float64(div))) / k}
	return r
}
