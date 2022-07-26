package mango

type TCPHandler interface {
	OnOpen(*TCPSession)
	OnClose(*TCPSession)
	OnPacket(*TCPSession, *Packet)
}

type TimerHandler interface {
	OnTime()
}
