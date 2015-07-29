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
	Key  []byte
	Item *Item
}

func (s *Storage) Get(key []byte, reply *Item) error {

	k := string(key)

	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.db[k]

	if !ok {
		return fmt.Errorf("key %x not existed", key)
	}
	// TODO wired pointer
	*reply = *r
	glog.V(3).Infof("Get Key %x %v", key, reply)
	return nil
}

func (s *Storage) Put(args *PutArgs, reply *bool) (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := string(args.Key)
	glog.V(3).Infof("Put Item %x %s", args.Key, args.Item)

	ditem, ok := s.db[key]
	if !ok {
		// Override
		s.db[key] = args.Item
		*reply = true
		return
	}

	if ditem.Ver >= args.Item.Ver {
		*reply = false
		err = fmt.Errorf("Invaild put request %x %s", args.Key, args.Item)
		glog.Error(err)
		return err
	}

	s.db[key] = args.Item
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
