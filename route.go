// Package onehop provides ...
package onehop

import (
	"math/big"
	"sync"
	"time"

	"github.com/golang/glog"
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
	slices []*Slice
	k      int // number of slices the ring is divided into
	block  *big.Int
	mu     *sync.RWMutex
}

func (r *Route) Len() int {
	return len(r.slices)
}

func (r *Route) GetIndex(id *big.Int) (sliceidx int) {

	if id.Int64() == int64(0) {
		return 0
	}
	sliceidx = int(new(big.Int).Div(id, r.block).Int64())
	return
}

func (r *Route) GetNode(id *big.Int) (n *Node) {

	r.mu.RLock()
	defer r.mu.RUnlock()

	slice_idx := r.GetIndex(id)
	slice := r.slices[slice_idx]

	return slice.Get(id)
}

func (r *Route) forward(slice_idx int) (sidx int) {

	sidx = slice_idx
	if sidx == r.k-1 {
		sidx = 0
	} else {
		sidx++
	}
	return
}

func (r *Route) SuccessorOf(id *big.Int) (n *Node) {

	slice_idx := r.GetIndex(id)

	for i := 0; i < r.k; i++ {
		slice := r.slices[slice_idx]

		n = slice.successorOf(id)
		if n != nil {
			break
		}
		slice_idx = r.forward(slice_idx)
		// Reset to 0 for loop back
		id = zeroID
	}

	return
}

func (r *Route) Add(n *Node) (ok bool) {

	r.mu.Lock()
	defer r.mu.Unlock()
	slice_idx := r.GetIndex(n.ID)
	slice := r.slices[slice_idx]

	result := slice.add(n)
	slice.updateLeader()

	return result
}
func (r *Route) Delete(id *big.Int) (ok bool) {

	r.mu.Lock()
	defer r.mu.Unlock()

	slice_idx := r.GetIndex(id)
	slice := r.slices[slice_idx]

	result := slice.Delete(id)
	return result
}

func NewRoute(k int) *Route {

	if k < 2 {
		panic("K  can't not less than 2")
	}
	glog.Infof("starting route k=%d", k)
	block := new(big.Int)
	block.SetBytes(FullID)

	block.Div(block, big.NewInt(int64(k)))
	// TODO wired length issues on divied number

	block.Add(block, big.NewInt(1))

	max_num := new(big.Int)
	max_num.SetBytes(FullID)

	l := make([]*Slice, 0)

	for i := int64(0); i < int64(k); i++ {
		slice := new(Slice)
		slice.Max = new(big.Int)
		slice.Min = new(big.Int)
		slice.Min.Mul(block, big.NewInt(i))
		slice.Max.Mul(block, big.NewInt((i + 1)))
		if slice.Max.Cmp(max_num) > 0 {
			slice.Max.SetBytes(FullID)
		} else {
			slice.Max.Sub(slice.Max, big.NewInt(int64(1)))
		}

		l = append(l, slice)
	}
	r := &Route{l, k, block, new(sync.RWMutex)}
	route = r
	return r
}

func (r *Route) ServeTimeout(self *Slice) {
	ticker := time.NewTicker(1 * time.Second)
	other := make([]*Slice, 0)
	for _, s := range r.slices {
		if s != self {
			other = append(other, s)
		}
	}
	// Node timeout checker
	for {
		<-ticker.C
		for _, s := range other {
			if s.Leader != nil &&
				s.Leader.updateAt.Add(NODE_TIMEOUT).Before(time.Now()) {
				glog.Errorf("Slice Leader Node %x timeout", s.Leader.ID)
				r.Delete(s.Leader.ID)
			}
		}
		for _, n := range self.nodes {
			if n.ID.String() == service.id.String() {
				continue
			}

			if n.updateAt.Add(NODE_TIMEOUT).Before(time.Now()) {
				glog.Errorf("Node %x timeout", n.ID)
				r.Delete(n.ID)
			}
		}
	}
}
