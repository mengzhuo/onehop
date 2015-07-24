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

	k := fmt.Sprintf("%x", key)

	glog.V(3).Infof("Get Key %s", k)
	s.mu.RLock()
	defer s.mu.RUnlock()
	var ok bool
	reply, ok = s.db[k]

	if !ok {
		return fmt.Errorf("key %s not existed", k)
	}
	return nil
}

func (s *Storage) Put(args *PutArgs, reply *bool) (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := fmt.Sprintf("%x", args.Key)
	glog.V(3).Infof("Put Item %s %x", key, args.Item)

	ditem, ok := s.db[key]
	if !ok {
		// Override
		s.db[key] = args.Item
		*reply = true
		return
	}

	if ditem.Ver > args.Item.Ver {
		*reply = false
		return fmt.Errorf("Invaild Id %d", args.Item.Ver)
	}

	s.db[key] = args.Item
	*reply = true
	return nil
}

type DeleteArgs struct {
	Key []byte
	Ver uint64
}

func (s *Storage) Delete(args *DeleteArgs, reply *bool) (err error) {

	s.mu.Lock()
	defer s.mu.Unlock()

	key := fmt.Sprintf("%x", args.Key)
	ditem, ok := s.db[key]
	if !ok {
		return fmt.Errorf("key %s not existed", args.Key)
	}

	if ditem.Ver > args.Ver {
		*reply = false
		return fmt.Errorf("Invaild Id %d", args.Ver)
	}
	delete(s.db, key)
	*reply = true
	return nil
}

func (s *Storage) Replicate(from string, reply *map[string]*Item) (err error) {

	s.mu.RLock()
	defer s.mu.RUnlock()

	glog.Infof("Node %s Replicate from us, items:%d", from, len(s.db))

	reply = &s.db
	return
}
