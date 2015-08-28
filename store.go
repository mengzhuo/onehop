// Package onehop provides ...
package onehop

import (
	"fmt"
	"sync"

	"github.com/golang/glog"
)

type Item struct {
	Ver  uint64
	Data []byte
}

func (i *Item) String() string {
	return fmt.Sprintf("Ver:%d Data:%x", i.Ver, i.Data)
}

// Thread safe storage
type Storage struct {
	db map[string]*Item
	mu *sync.RWMutex
}

func NewStorage() *Storage {

	mu := new(sync.RWMutex)
	db := make(map[string]*Item, 0)
	return &Storage{db, mu}
}

type PutArgs struct {
	Key  string
	Item *Item
}

func (s *Storage) Get(key string, reply *Item) error {

	s.mu.RLock()
	defer s.mu.RUnlock()
	if r, ok := s.db[key]; ok {
		*reply = *r
		glog.V(3).Infof("Get Key %s %v", key, reply)
	}
	return nil
}

func (s *Storage) put(key string, item *Item) (ok bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if d_item, ok := s.db[key]; ok {
		if d_item.Ver > item.Ver {
			return false
		}
	}
	s.db[key] = item
	return true
}

func (s *Storage) Put(args *PutArgs, reply *bool) (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	glog.V(3).Infof("Put Item %s %s", args.Key, args.Item)

	ditem, ok := s.db[args.Key]
	if !ok {
		// Override
		s.db[args.Key] = args.Item
		*reply = true
		return
	}

	if ditem.Ver >= args.Item.Ver {
		*reply = false
		err = fmt.Errorf("Invaild put request %x %s", args.Key, args.Item)
		glog.Error(err)
		return err
	}

	s.db[args.Key] = args.Item
	*reply = true
	return nil
}

func (s *Storage) Replicate(from string, reply *map[string]*Item) (err error) {

	s.mu.RLock()
	defer s.mu.RUnlock()

	glog.Infof("Node %s Replicate from us, items:%d", from, len(s.db))
	if len(s.db) == 0 {
		return fmt.Errorf("Nothing to replicated")
	}
	*reply = s.db
	return
}
