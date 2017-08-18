package main

import "net"

type NoDelayListener net.TCPListener

func NewNoDelayListener(raddr string) (*NoDelayListener, error) {
	addr, err := net.ResolveTCPAddr("tcp4", raddr)
	if err != nil {
		return nil, err
	}
	ln, err := net.ListenTCP("tcp4", addr)
	return (*NoDelayListener)(ln), err
}

func (cl *NoDelayListener) Accept() (net.Conn, error) {
	conn, err := (*net.TCPListener)(cl).AcceptTCP()
	if err != nil {
		return conn, err
	}
	conn.SetNoDelay(true)
	return conn, err
}

func (cl *NoDelayListener) Addr() net.Addr {
	return (*net.TCPListener)(cl).Addr()
}

func (cl *NoDelayListener) Close() error {
	return (*net.TCPListener)(cl).Close()
}
