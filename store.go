// Package onehop provides ...
package onehop

import (
	"fmt"
	"sync"
)

type Item struct {
	Id   uint64
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
	Key  string
	Item *Item
}

func (s *Storage) Get(key string, reply *Item) error {

	s.mu.RLock()
	defer s.mu.RUnlock()
	reply, ok := s.db[key]

	if !ok {
		return fmt.Errorf("key %s not existed", key)
	}
	return nil
}

func (s *Storage) Put(args *PutArgs, reply *bool) (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ditem, ok := s.db[args.Key]
	if !ok {
		// Override
		s.db[args.Key] = args.Item
		*reply = true
		return
	}

	if ditem.Id > args.Item.Id {
		*reply = false
		return fmt.Errorf("Invaild Id %d", args.Item.Id)
	}

	s.db[args.Key] = args.Item
	*reply = true
	return nil
}

type DeleteArgs struct {
	Key string
	Id  uint64
}

func (s *Storage) Delete(args *DeleteArgs, reply *bool) (err error) {

	s.mu.Lock()
	defer s.mu.Unlock()

	ditem, ok := s.db[args.Key]
	if !ok {
		return fmt.Errorf("key %s not existed", args.Key)
	}

	if ditem.Id > args.Id {
		*reply = false
		return fmt.Errorf("Invaild Id %d", args.Id)
	}
	delete(s.db, args.Key)
	*reply = true
	return nil
}
