package mango

import (
	logger "TCPEx/Engine/Logger"
	"errors"
	"io"
	"net"
	"runtime"
	"sync"
	"time"
)

type TCPSession struct {
	_Conn        net.Conn
	_WriteChan   chan *Packet
	_IsClose     bool
	_ReadTimeOut time.Duration
	_Observer    interface{}
	_DoneChan    chan struct{}
	_IsChanFull  bool
	sync.RWMutex
}

func NewSession(conn net.Conn, pendingWriteNum int, readTimeOut time.Duration) *TCPSession {
	s := new(TCPSession)
	s._Conn = conn
	s._WriteChan = make(chan *Packet, pendingWriteNum)
	s._IsClose = false
	s._ReadTimeOut = readTimeOut
	s._Observer = nil
	s._DoneChan = make(chan struct{}, 0)
	s._IsChanFull = false
	s.doWork()
	return s
}

func (s *TCPSession) doWork() {
	// Write goroutine 시작
	go func() {
		for packet := range s._WriteChan {
			if packet == nil {
				break
			}

			isWriteFailed := false
			writeRetryCnt := 0
			totalSendByte := 0

			// Send Buffer 가 다 차서 일부데이터만 전송되는 경우가 발생할 수 있으므로
			// 전체 전송 바이트를 체크하면서 전송한다.
			for totalSendByte < packet.Len() {
				sendByte, err := s._Conn.Write(packet.FinishedBytes()[totalSendByte:packet.Len()])
				if err != nil {
					// 일시적으로 네트워크 상태에 이상이 생겨 Write 에 에러가 발생할 수 있음.
					// 잠시 지연 후 다시 시도
					if nErr, ok := err.(net.Error); ok && nErr.Temporary() {
						runtime.Gosched() // 다른 고루틴에 양보
						writeRetryCnt++
						// N번 재시도후 실패인 경우
						if writeRetryCnt >= maxWriteRetryCount {
							// return to pool (N번 전송 에러된 패킷 반환)
							//DelPacket(packet)
							isWriteFailed = true
							break
						}
						// 전송 재시도
						continue
					}
					// return to pool (전송 에러된 패킷 반환)
					//DelPacket(packet)
					isWriteFailed = true
					break
				}
				// 전송된 바이트 수 누적
				totalSendByte += sendByte
			}
			// 패킷 반환
			DelPacket(packet)

			// 전송 실패 접속 종료
			if isWriteFailed {
				break
			}
		}

		//fmt.Println("closed write goroutine. Addr(", tcpSession.RemoteAddr(), ")")
		_ = s._Conn.(*net.TCPConn).SetLinger(0)
		_ = s._Conn.Close()

		s.Lock()
		s._IsClose = true
		//logger.Debug("Observer = nil")
		//s._Observer = nil

		//remainPacketCount := len(s._WriteChan)
		//for i := 0; i < remainPacketCount; i++ {
		//	packet := <-s._WriteChan
		//	if packet != nil {
		//		DelPacket(packet) // return to pool
		//	} else {
		//		logger.Warnf("Oops! packet is nil! Addr(%v)", s.RemoteAddr())
		//	}
		//}
		close(s._WriteChan)
		for packet := range s._WriteChan {
			if packet != nil {
				DelPacket(packet) // return to pool
			}
		}
		s.Unlock()
	}()
}

func (s *TCPSession) SetObserver(ob interface{}) {
	s.Lock()
	defer s.Unlock()
	s._Observer = ob
}

func (s *TCPSession) Observer() interface{} {
	s.RLock()
	defer s.RUnlock()
	return s._Observer
}

// func (tcpSession *TCPSession) doDestroy() {
// 	tcpSession.Conn.(*net.TCPConn).SetLinger(0)
// 	tcpSession.Conn.Close()

// 	if !tcpSession.isClose {
// 		close(tcpSession.WriteChan)
// 		tcpSession.isClose = true
// 		tcpSession.User = nil
// 	}
// }

// func (tcpSession *TCPSession) Destroy() {
// 	tcpSession.Lock()
// 	defer tcpSession.Unlock()

// 	tcpSession.doDestroy()
// }

func (s *TCPSession) isConnected() bool {
	s.Lock()
	defer s.Unlock()
	if s._IsClose {
		return false
	}
	return true
}

func (s *TCPSession) Close() {
	s.Lock()
	defer s.Unlock()
	if s._IsClose {
		return
	}
	s.doWrite(nil)
	s._IsClose = true
}

func (s *TCPSession) doWrite(packet *Packet) {
	chanCap := cap(s._WriteChan) - 1
	if len(s._WriteChan) >= chanCap {
		logger.Errorf("[TCPSession] WriteChan is Full! Addr(%v) Len(%v) Cap(%v)", s.RemoteAddr(), len(s._WriteChan), chanCap)
		// 유저의 세션을 끊어야 할까?
		if !s._IsChanFull {
			s._WriteChan <- nil
			s._IsChanFull = true
		}

		if nil != packet {
			DelPacket(packet)
		}
		return
	}
	s._WriteChan <- packet
}

// packet must not be modified by the others goroutines
func (s *TCPSession) Write(packet *Packet) {
	s.Lock()
	defer s.Unlock()

	if nil == packet {
		return
	}

	if s._IsClose {
		DelPacket(packet)
		return
	}

	s.doWrite(packet)
}

func (s *TCPSession) Read(b []byte) (int, error) {
	if nil == s._Conn {
		return 0, errors.New("nil conn")
	}
	return io.ReadFull(s._Conn, b)

}

func (s *TCPSession) LocalAddr() net.Addr {
	return s._Conn.LocalAddr()
}

func (s *TCPSession) RemoteAddr() net.Addr {
	return s._Conn.RemoteAddr()
}

func (s *TCPSession) Done() {
	close(s._DoneChan)
}
