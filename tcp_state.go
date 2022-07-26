package mango

import "fmt"

type TCPState struct {
	funcMap map[uint16]func(*TCPSession, *Packet)
	name    string
}

func NewState(name string) *TCPState {
	return &TCPState{
		funcMap: make(map[uint16]func(*TCPSession, *Packet)),
		name:    name,
	}
}

func (s *TCPState) AddCommand(cmd uint16, f func(*TCPSession, *Packet)) {
	if _, ok := s.funcMap[cmd]; ok {
		return
	}
	s.funcMap[cmd] = f
}

func (s *TCPState) GetCommand(cmd uint16) (func(*TCPSession, *Packet), error) {
	if f, ok := s.funcMap[cmd]; ok {
		return f, nil
	}
	return nil, fmt.Errorf("NOT Exists Command(%v)", cmd)
}

func (s *TCPState) GetName() string {
	return s.name
}
