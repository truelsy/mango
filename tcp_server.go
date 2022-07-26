package mango

import (
	logger "TCPEx/Engine/Logger"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	_DefaultConnNum = 3000
)

type ConnSet map[net.Conn]struct{}

type TCPServer struct {
	_Addr            string
	_Port            int
	_MaxConnNum      int
	_PendingWriteNum int
	_ReadTimeOut     time.Duration
	_Ln              *net.TCPListener
	_Conns           ConnSet
	_WgLn            sync.WaitGroup
	_WgConns         sync.WaitGroup
	_Handler         TCPHandler
	sync.Mutex
}

func NewServer(addr string, port int, maxConnNum int, handler TCPHandler) *TCPServer {
	tcpAddr, _ := net.ResolveTCPAddr("tcp", addr+":"+strconv.Itoa(port))
	ln, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Listen Error. cause(", err, ")")
		return nil
	}

	if 0 >= maxConnNum {
		maxConnNum = _DefaultConnNum
	}

	s := TCPServer{
		_Addr:            addr,
		_Port:            port,
		_MaxConnNum:      maxConnNum,
		_PendingWriteNum: defaultPendingWriteNum,
		_ReadTimeOut:     0,
		_Ln:              ln,
		_Conns:           make(ConnSet),
		_WgLn:            sync.WaitGroup{},
		_WgConns:         sync.WaitGroup{},
		_Handler:         handler,
		Mutex:            sync.Mutex{},
	}

	return &s
}

func (t *TCPServer) SetPendingWriteNum(n int) {
	t._PendingWriteNum = n
}

func (t *TCPServer) SetReadTimeOut(d time.Duration) {
	t._ReadTimeOut = d
}

func (t *TCPServer) Port() int {
	return t._Port
}

func (t *TCPServer) Listen() {
	go t.run()
}

func (t *TCPServer) run() {
	t._WgLn.Add(1)
	defer t._WgLn.Done()

	var tempDelay time.Duration

	for {
		conn, err := t._Ln.AcceptTCP()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if 0 == tempDelay {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}

				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}

				logger.Error("[TCPServer] Accept error:", err, ", retrying in ", tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return
		}

		tempDelay = 0

		if t.isMaxConn(conn) {
			logger.Error("[TCPServer] Too many connections. cur:", len(t._Conns), ", max:", t._MaxConnNum)
			continue
		}

		t.registConns(conn)
		t._WgConns.Add(1)

		s := NewSession(conn, t._PendingWriteNum, t._ReadTimeOut)

		go func() {
			logger.Protect(func() {
				// Connect
				t._Handler.OnOpen(s)
				//JobPush(OnOpen{s, t._Handler})
				// run Worker
				doWorker(s, t._Handler)
				// Closed
				// 직접 호출 하는 경우 One Dispather규칙이 깨짐.
				//t._Handler.OnClose(s)
				JobPush(OnClose{s, t._Handler})
				<-s._DoneChan
			})

			// CleanUp
			s.Close()
			t.removeConns(conn)

			t._WgConns.Done()
		}()
	}
}

func (t *TCPServer) isMaxConn(conn *net.TCPConn) bool {
	t.Lock()
	defer t.Unlock()
	if len(t._Conns) >= t._MaxConnNum {
		conn.Close()
		return true
	}
	return false
}

func (t *TCPServer) registConns(conn *net.TCPConn) {
	t.Lock()
	defer t.Unlock()
	t._Conns[conn] = struct{}{}
	//fmt.Println("Connected!! Addr(", conn.RemoteAddr(), ") Count(", len(tcpServer.Conns), ")")
}

func (t *TCPServer) removeConns(conn *net.TCPConn) {
	t.Lock()
	defer t.Unlock()
	delete(t._Conns, conn)
	//fmt.Println("Closed!! Addr(", conn.RemoteAddr(), ") Count(", len(tcpServer.Conns), ")")
}

func (t *TCPServer) destroyConns() {
	t.Lock()
	defer t.Unlock()
	for conn := range t._Conns {
		conn.Close()
	}
	t._Conns = nil
}

func (t *TCPServer) Close() {
	t._Ln.Close()
	t._WgLn.Wait()

	t.destroyConns()
	t._WgConns.Wait()

	logger.CInfo(logger.FG_LIGHT_YELLOW, "[TCPServer] Close.")
}
