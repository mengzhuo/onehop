// Package main provides ...
package main

import (
	"bufio"
	"encoding/hex"
	"flag"
	"fmt"
	"net"
	"onehop"
	"os"
	"strconv"
	"strings"
)

var (
	K       = flag.Int("k", 2, "K value")
	R       = flag.Int("r", 2, "R value")
	W       = flag.Int("w", 2, "W value")
	bind_ep = flag.String("b", "", "bind address")
	boot_ep = flag.String("t", "", "BOOTSTRAP address")
)

func main() {
	flag.Parse()
	s := onehop.NewService("udp", *bind_ep, *K, *W, *R)
	if *boot_ep != "" {
		addr, _ := net.ResolveUDPAddr("udp", *boot_ep)
		msg := &onehop.Msg{}
		msg.Type = onehop.BOOTSTRAP
		msg.NewID()
		s.SendMsg(addr, msg)
	}
	go s.Listen()
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		t := scanner.Text()
		data := strings.Split(t, " ")
		cmd := data[0]
		key, err := hex.DecodeString(data[1])
		if err != nil {
			fmt.Println(err)
			continue
		}
		switch cmd {
		case "GET":
			fmt.Printf(">>> %s\n", s.Get(key))
		case "PUT":
			version, _ := strconv.Atoi(data[2])
			data := data[3]
			a := &onehop.Item{uint64(version), []byte(data)}
			fmt.Printf(">>> %s\n", s.Put(key, a))
		}
	}
}
