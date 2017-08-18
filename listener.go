package main

import "net"
import "log"
import "syscall"

type Listener net.TCPListener

const (
	tcpFastOpen  = 0x17
	fastOpenQlen = 16 * 1024
)

func NewListener(raddr string) (*Listener, error) {
	addr, err := net.ResolveTCPAddr("tcp4", raddr)
	if err != nil {
		return nil, err
	}
	ln, err := net.ListenTCP("tcp4", addr)
	if err != nil {
		return nil, err
	}
	f, _ := ln.File()
	fd := int(f.Fd())
	if err := syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_DEFER_ACCEPT, 1); err != nil {
		log.Fatal("cannot enable TCP_DEFER_ACCEPT", err)
	}
	if err := syscall.SetsockoptInt(fd, syscall.SOL_TCP, tcpFastOpen, fastOpenQlen); err != nil {
		log.Fatalf("cannot enable TCP_FASTOPEN(qlen=%d): %s", fastOpenQlen, err)
	}
	return (*Listener)(ln), err
}

func (cl *Listener) Accept() (net.Conn, error) {
	conn, err := (*net.TCPListener)(cl).AcceptTCP()
	if err != nil {
		return conn, err
	}
	return conn, err
}

func (cl *Listener) Addr() net.Addr {
	return (*net.TCPListener)(cl).Addr()
}

func (cl *Listener) Close() error {
	return (*net.TCPListener)(cl).Close()
}
