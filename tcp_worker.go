package mango

import (
	logger "TCPEx/Engine/Logger"
	"encoding/binary"
	"time"
)

func doWorker(s *TCPSession, handler interface{}) {
	readBuffer := make([]byte, bufferSize)
	msgHeaderLen := GetHeaderLen()

	for {
		if 0 < s._ReadTimeOut {
			// timeout after 3min
			//logger.CDebugf(logger.FG_LIGHT_BLUE, "SetReadDeadline : %v", s._ReadTimeOut)
			err := s._Conn.SetReadDeadline(time.Now().Add(s._ReadTimeOut))
			if err != nil {
				logger.Errorf("failed to SetReadDeadline. err(%v)", err)
			}
		}

		if _, err := s.Read(readBuffer[:msgHeaderLen]); err != nil {
			//fmt.Fprintln(os.Stderr, "failed read header. err(", err, ") Addr(", tcpSession.RemoteAddr(), ")")
			//logger.Debugf("[TCPWorker] Error Read Head. Err(%v) Addr(%v)", err, s.RemoteAddr())
			break
		}

		if readBuffer[0] != Stx {
			logger.Error("[TCPWorker] Invalid Stx(", readBuffer[0], ") Addr(", s.RemoteAddr(), ")")
			break
		}

		msgLen := binary.LittleEndian.Uint16(readBuffer[1:])
		if 0 > msgLen {
			logger.Error("[TCPWorker] Invalid packet msgLen(", msgLen, ") Addr(", s.RemoteAddr(), ")")
			break
		}

		msgTotalLen := msgHeaderLen + msgLen
		if bufferSize < int(msgTotalLen) {
			logger.Error("[TCPWorker] Buffer overflow! msgTotalLen(", msgTotalLen, ") Addr(", s.RemoteAddr(), ")")
			break
		}

		//fmt.Println("msgTotalLen(", msgTotalLen, ") msgHeaderLen(", msgHeaderLen, ") msgLen(", msgLen, ")")

		// remain data
		if msgTotalLen > msgHeaderLen {
			if _, err := s.Read(readBuffer[msgHeaderLen:msgTotalLen]); err != nil {
				//fmt.Fprintln(os.Stderr, "failed read body. err(", err, ") Addr(", tcpSession.RemoteAddr(), ")")
				logger.Debugf("[TCPWorker] failed read body. Err(%v) Addr(%v)", err, s.RemoteAddr())
				break
			}
		}

		// Packet 객체 생성
		pack := NewPacket()

		// Copy하는데 3-10 나노초 정도 소요된다.
		pack.Copy(readBuffer[:msgTotalLen])

		//switch workType {
		//case eServerWork:
		//  JobPush(OnServer{s, pack, handler.(ServerHandler)})
		//case eClientWork:
		//  JobPush(OnClient{s, pack, handler.(ClientHandler)})
		//}
		JobPush(OnPacket{s, pack, handler.(TCPHandler)})
	}
}

//// doWorker Echo Test Worker
//func doWorker(s *TCPSession, handler interface{}) {
//	readBuffer := make([]byte, bufferSize)
//	msgHeaderLen := GetHeaderLen()
//
//	const (
//		StxOffset  = 0
//		SizeOffset = 1
//	)
//
//	for {
//		if 0 < s._ReadTimeOut {
//			// timeout after 3min
//			//logger.CDebugf(logger.FG_LIGHT_BLUE, "SetReadDeadline : %v", s._ReadTimeOut)
//			err := s._Conn.SetReadDeadline(time.Now().Add(s._ReadTimeOut))
//			if nil != err {
//				logger.Error("occur error(%v)", err)
//			}
//		}
//
//		if _, err := s.Read(readBuffer[:4]); err != nil {
//			//logger.Debugf("[TCPWorker] Error Read Head. Err(%v) Addr(%v)", err, s.RemoteAddr())
//			break
//		}
//
//		//if readBuffer[StxOffset] != Stx {
//		//   logger.Error("[TCPWorker] Invalid Stx(", readBuffer[StxOffset], ") Addr(", s.RemoteAddr(), ")")
//		//   break
//		//}
//		msgTotalLen := uint16(0)
//		if readBuffer[StxOffset] == Stx {
//			if _, err := s.Read(readBuffer[4:6]); err != nil {
//				//logger.Debugf("[TCPWorker] Error Read Head. Err(%v) Addr(%v)", err, s.RemoteAddr())
//				break
//			}
//
//			msgLen := binary.LittleEndian.Uint16(readBuffer[SizeOffset:])
//			//logger.Debug("msgLen:", msgLen)
//			if 0 > msgLen {
//				logger.Error("[TCPWorker] Invalid packet msgLen(", msgLen, ") Addr(", s.RemoteAddr(), ")")
//				break
//			}
//
//			msgTotalLen = msgHeaderLen + msgLen
//			// test 용
//			//msgTotalLen := msgLen
//			if bufferSize < int(msgTotalLen) {
//				logger.Error("[TCPWorker] Buffer overflow! msgTotalLen(", msgTotalLen, ") Addr(", s.RemoteAddr(), ")")
//				break
//			}
//
//			//logger.Debug("msgTotalLen(", msgTotalLen, ") msgHeaderLen(", msgHeaderLen, ") msgLen(", msgLen, ")")
//
//			// remain data
//			if msgTotalLen > msgHeaderLen {
//				// test 용
//				//if _, err := s.Read(readBuffer[4:msgTotalLen]); err != nil {
//				if _, err := s.Read(readBuffer[msgHeaderLen:msgTotalLen]); err != nil {
//					logger.Debugf("[TCPWorker] failed read body. Err(%v) Addr(%v)", err, s.RemoteAddr())
//					break
//				}
//			}
//		} else {
//			msgLen := binary.LittleEndian.Uint32(readBuffer[0:])
//			//logger.Debug("msgLen:", msgLen)
//			if 0 > msgLen {
//				logger.Error("[TCPWorker] Invalid packet msgLen(", msgLen, ") Addr(", s.RemoteAddr(), ")")
//				break
//			}
//
//			// test 용
//			msgTotalLen = uint16(msgLen)
//			if bufferSize < int(msgTotalLen) {
//				logger.Error("[TCPWorker] Buffer overflow! msgTotalLen(", msgTotalLen, ") Addr(", s.RemoteAddr(), ")")
//				break
//			}
//
//			//logger.Debug("msgTotalLen(", msgTotalLen, ") msgHeaderLen(", msgHeaderLen, ") msgLen(", msgLen, ")")
//
//			// remain data
//			if msgTotalLen > 4 {
//				// test 용
//				if _, err := s.Read(readBuffer[4:msgTotalLen]); err != nil {
//					//if _, err := s.Read(readBuffer[msgHeaderLen:msgTotalLen]); err != nil {
//					logger.Debugf("[TCPWorker] failed read body. Err(%v) Addr(%v)", err, s.RemoteAddr())
//					break
//				}
//			}
//		}
//
//		// Packet 객체 생성
//		pack := NewPacket()
//
//		// Copy 하는데 3-10 나노초 정도 소요된다.
//		pack.Copy(readBuffer[:msgTotalLen])
//
//		//switch workType {
//		//case eServerWork:
//		//   JobPush(OnServer{s, pack, handler.(ServerHandler)})
//		//case eClientWork:
//		//   JobPush(OnClient{s, pack, handler.(ClientHandler)})
//		//}
//		JobPush(OnPacket{s, pack, handler.(TCPHandler)})
//	}
//}
