package onehop

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/golang/glog"
)

type Service struct {
	conn *net.UDPConn

	route *Route

	id string

	exchangeEvent map[string]*Event
	notifyEvent   map[string]*Event

	selfSlice *Slice
	W, R      int
	bytePool  *LeakyBuffer

	InBuffer  *bytes.Buffer
	InDecoder *json.Decoder
	InLock    *sync.Mutex

	OutBuffer  *bytes.Buffer
	OutEncoder *json.Encoder
	OutLock    *sync.Mutex
}

func NewUDPDecoder(buf *bytes.Buffer) *json.Decoder {
	return json.NewDecoder(buf)
}

// NetType, Address for UDP connection
// k for OneHop slice number
func NewService(netType, address string, k, w, r int) *Service {

	glog.Infof("Write:%d, Read:%d", k, r)
	listener, err := newListener(netType, address)
	if err != nil {
		panic(err)
	}
	if err != nil {
		panic(err)
	}
	glog.Infof("Listening to :%s %s", netType, address)

	route = NewRoute(k)
	vid := make([]byte, 16)
	rand.Read(vid)
	id := hex.EncodeToString(vid)

	n := &Node{ID: id,
		Addr: listener.LocalAddr().(*net.UDPAddr)}
	route.Add(n)

	slice := route.slices[route.GetIndex(id)]
	glog.V(1).Info("Self Slice:", slice.Max)

	inBuffer := bytes.NewBuffer(make([]byte, 0))
	outBuffer := bytes.NewBuffer(make([]byte, 0))
	service = &Service{
		listener, route,
		id,
		make(map[string]*Event, 0),
		make(map[string]*Event, 0),
		slice,
		w, r,
		NewLeakyBuffer(1024, MSG_MAX_SIZE),
		inBuffer,
		json.NewDecoder(inBuffer),
		&sync.Mutex{},
		outBuffer,
		json.NewEncoder(outBuffer),
		&sync.Mutex{}}

	glog.V(3).Infof("RPC Listener Accepted")
	go service.Start()
	return service
}

/*
func (s *Service) Get(key []byte) *Item {

	items := make([]*Item, 0)
	var first_node *Node
	id := BytesToId(key)

	for i := 0; i < s.R; i++ {
		node := s.route.SuccessorOf(id)

		glog.V(3).Infof("Get Found SuccessorOf %x is %s", id, node)
		if node == nil {
			// no more to read...
			continue
		}
		if first_node == nil {
			first_node = node
		} else {
			// Loopbacked
			if node == first_node {
				break
			}
		}

		if node == s.selfNode {
			// it's ourself
			if item, ok := s.db.db[string(key)]; ok {
				items = append(items, item)
			}
			id = node.ID
			continue
		}

		client, err := s.RPCPool.Get(node.Addr.String())

		if err != nil {
			glog.Error(err)
			s.NotifySliceLeader(node, LEAVE)
			s.route.Delete(node.ID)
			continue
		}
		var reply *Item
		err = client.Call("Storage.Get", key, &reply)
		glog.V(1).Infof("Get %x From:%s", key, node.Addr)
		if err == nil && reply != nil {
			glog.V(1).Infof("Get reply %s", reply)
			items = append(items, reply)
		}
		id = node.ID
	}

	if len(items) == 0 {
		return nil
	}

	max_item := items[0]

	if len(items) > 1 {
		for _, item := range items[1:] {
			if item.Ver > max_item.Ver {
				max_item = item
			}
		}
	}

	return max_item
}

func (s *Service) PutByString(key string, item *Item) int {
	k, err := hex.DecodeString(key)
	if err != nil {
		glog.Error(err)
		return 0
	}
	return s.Put(k, item)
}

func (s *Service) Put(key []byte, item *Item) (count int) {

	id := BytesToId(key)

	var first_node *Node
	count = 0

	for i := 0; i < s.W; i++ {

		node := s.route.SuccessorOf(id)
		glog.V(3).Infof("Put Found SuccessorOf %x is %s", id, node)

		if node == nil {
			// no more to write...
			continue
		}

		if first_node == nil {
			first_node = node
		} else {
			// Loopbacked
			if node == first_node {
				break
			}
		}
		if node == s.selfNode {
			// it's ourself
			k := string(key)
			if selfItem, ok := s.db.db[k]; !ok {
				s.db.db[k] = item
				count += 1
			} else {
				if selfItem.Ver < item.Ver {
					s.db.db[k] = item
					count += 1
				}
			}
			id = node.ID
			continue
		}
		glog.V(1).Infof("Put %x to %s", key, node.Addr.String())
		id = node.ID

		client, err := s.RPCPool.Get(node.Addr.String())
		if s.RPCError(err, node) {
			continue
		}

		args := &PutArgs{key, item}
		var reply *bool
		err = client.Call("Storage.Put", args, &reply)
		if !s.RPCError(err, node) {
			count += 1
		}
	}
	return
}
*/
func (s *Service) RPCError(err error, node *Node) bool {
	if err != nil {
		glog.Errorf("Node:%x %#v", node.ID, err)
		switch err.(type) {

		case *net.OpError:
			glog.Error("RPC call failed")
			//s.NotifySliceLeader(node, LEAVE)
			s.route.Delete(node.ID)
		}
		return true
	}
	return false
}

func (s *Service) BootStrapFrom(address string) {

	glog.Infof("BootStrap From :%s", address)
	addr, _ := net.ResolveUDPAddr("udp", address)
	msg := new(Msg)
	msg.Type = BOOTSTRAP
	msg.From = s.id
	s.SendMsg(addr, msg)
}

func BytesToId(p []byte) string {
	return fmt.Sprintf("%032x", p)
}

func (s *Service) Listen() {

	for {

		p := s.bytePool.Get()
		n, addr, err := s.conn.ReadFromUDP(p)
		if err != nil || n < 5 {
			glog.Errorf("insufficient data from %s:%x", addr, p[:n])
			continue
		}
		glog.V(10).Infof("Recv From %s with %d", addr, n)

		msg := new(Msg)
		s.InBuffer.Truncate(0)
		nn, err := s.InBuffer.Write(p[:n])
		if err != nil || nn != n {
			glog.Error(err)
			continue
		}

		err = s.InDecoder.Decode(msg)
		s.InBuffer.Truncate(0)
		if err != nil {
			glog.Error(err, s.InBuffer.Len())
			continue
		}

		glog.V(9).Infof("Msg:%#v", msg)
		go s.handle(addr, msg)
		s.bytePool.Put(p)
	}
}

func (s *Service) Send(dstAddr *net.UDPAddr, p []byte) {

	defer func() {
		if r := recover(); r != nil {
			glog.ErrorDepth(0, r)
		}
	}()
	if dstAddr == s.conn.LocalAddr() {
		return
	}
	s.conn.WriteToUDP(p, dstAddr)
}

func (s *Service) SendMsg(dstAddr *net.UDPAddr, msg *Msg) {

	s.OutLock.Lock()
	s.OutEncoder.Encode(msg)
	p := s.OutBuffer.Bytes()
	s.OutBuffer.Reset()

	s.OutLock.Unlock()

	glog.V(10).Infof("SEND %s with %x", dstAddr, p)
	go s.Send(dstAddr, p)
}

func (s *Service) Start() {

	glog.Info("Checker on")
	for {
		time.Sleep(1 * time.Second)
		s.check()
	}
}

func (s *Service) check() {

	// Check Timeout node/leader
	for _, slice := range s.route.slices {
		if slice == s.selfSlice {
			continue
		}

		var leader *Node
		if leader = slice.Leader(); leader == nil {
			continue
		}
		if time.Since(leader.updateAt).Seconds() > SLICE_LEADER_TIMEOUT {
			s.exchangeEvent[leader.ID] = &Event{LEAVE, leader.Addr}
			s.route.Delete(leader.ID)
			continue
		}
	}

	if s.selfSlice.Leader().ID == s.id {
		s.exchange()
		//s.keepOtherAlive(s.exchangeEvent)
	}
}
