// Package onehop provides ...
package onehop

import (
	"log"
	"math/big"
	"sync"
	"time"
)

var (
	FullID = []byte{0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff}
	zeroID = big.NewInt(int64(0))
)

const NODE_TIMEOUT = 10 * time.Second

type Route struct {
	slices      []*Slice
	k           int // number of slices the ring is divided into
	u           int
	block       *big.Int
	mu          *sync.RWMutex
	timeoutNode chan *Node
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

	r.mu.RLock()
	defer r.mu.RUnlock()

	slice_idx, unit_idx := r.GetIndex(id)

	slice := r.slices[slice_idx]
	unit := slice.units[unit_idx]

	return unit.Get(id)
}

func (r *Route) forward(slice_idx, unit_idx int) (sidx, uidx int) {

	sidx, uidx = slice_idx, unit_idx

	// Uidx not full
	if uidx != r.u-1 {
		uidx++
		return
	}

	// Uidx fulled goto next slice
	uidx = 0
	sidx++
	if sidx == r.k-1 {
		sidx = 0
	}
	return
}

func (r *Route) SuccessorOf(id *big.Int) (n *Node) {

	slice_idx, unit_idx := r.GetIndex(id)

	for i := 0; i < r.u*r.k; i++ {
		slice := r.slices[slice_idx]
		unit := slice.units[unit_idx]

		n = unit.successorOf(id)
		if n != nil {
			break
		}
		slice_idx, unit_idx = r.forward(slice_idx, unit_idx)
		// Reset to 0 for loop back
		id = zeroID
	}

	return
}

func (r *Route) Refresh(id *big.Int) (ok bool) {

	r.mu.Lock()
	defer r.mu.Unlock()

	slice_idx, unit_idx := r.GetIndex(id)
	slice := r.slices[slice_idx]

	unit := slice.units[unit_idx]

	n := unit.Get(id)
	if n != nil {
		n.resetTimer()
		return true
	}
	return false
}

func (r *Route) Add(n *Node) (ok bool) {

	r.mu.Lock()
	defer r.mu.Unlock()

	slice_idx, unit_idx := r.GetIndex(n.ID)
	slice := r.slices[slice_idx]

	unit := slice.units[unit_idx]

	result := unit.add(n)
	if result {
		n.resetTimer()
	}
	slice.updateLeader()

	return result
}
func (r *Route) Delete(id *big.Int) (ok bool) {

	r.mu.Lock()
	defer r.mu.Unlock()

	slice_idx, unit_idx := r.GetIndex(id)
	slice := r.slices[slice_idx]

	unit := slice.units[unit_idx]
	result := unit.Delete(id)
	slice.updateLeader()
	return result
}

func NewRoute(k int, u int) *Route {

	if k < 2 || u < 2 {
		panic("K or U can't not less than 2")
	}
	log.Printf("starting route k=%d, u=%d", k, u)
	block := new(big.Int)
	block.SetBytes(FullID)
	block.Div(block, big.NewInt(int64(k*u)))
	// TODO wired length issues on divied number
	block.Add(block, big.NewInt(1))

	max_num := new(big.Int)
	max_num.SetBytes(FullID)

	l := make([]*Slice, 0)

	for i := int64(0); i < int64(k); i++ {
		slice := new(Slice)
		slice.Max = new(big.Int)
		slice.Min = new(big.Int)
		slice.Min.Mul(block, big.NewInt(i*int64(u)))
		slice.Max.Mul(block, big.NewInt((i+1)*int64(u)))
		if slice.Max.Cmp(max_num) > 0 {
			slice.Max.SetBytes(FullID)
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
				unit.Max.SetBytes(FullID)
			} else {
				unit.Max.Sub(unit.Max, big.NewInt(int64(1)))
			}

			slice.units = append(slice.units, unit)
		}
	}

	r := &Route{l, k, u, block, new(sync.RWMutex),
		make(chan *Node, 16)}
	go r.Serve()
	return r
}

func (r *Route) Serve() {
	for n := range r.timeoutNode {
		log.Printf("Node:%x timeouted", n.ID)
		r.Delete(n.ID)
	}
}
