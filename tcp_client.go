package mango

import (
	logger "TCPEx/Engine/Logger"
	"net"
	"strconv"
	"sync"
	"time"
)

const (
	_DefaultInterval = 3 * time.Second
)

//type ClientServerInfo struct {
//	_Kind int
//	_Idx  int
//}
//
//func (i ClientServerInfo) Kind() int {
//	return i._Kind
//}
//
//func (i ClientServerInfo) Idx() int {
//	return i._Idx
//}
//
//func (i ClientServerInfo) IsSame(a *ClientServerInfo) bool {
//	if i._Kind != a._Kind {
//		return false
//	}
//	if i._Idx != a._Idx {
//		return false
//	}
//	return true
//}

type TCPClient struct {
	_Addr            string
	_Port            int
	_PendingWriteNum int
	_ReadTimeOut     time.Duration
	_ConnInterval    time.Duration
	_IsClose         bool
	_Session         *TCPSession
	_Handler         TCPHandler
	//_ServerInfo      *ClientServerInfo
	_Observer interface{}
	_Wg       sync.WaitGroup
	sync.RWMutex
}

func NewClient(addr string, port int, connIntervalSec time.Duration, handler TCPHandler) *TCPClient {
	if 0 >= connIntervalSec {
		connIntervalSec = _DefaultInterval
	}

	c := TCPClient{
		_Addr:            addr,
		_Port:            port,
		_PendingWriteNum: defaultPendingWriteNum,
		_ReadTimeOut:     0,
		_ConnInterval:    connIntervalSec,
		_IsClose:         false,
		_Session:         nil,
		_Handler:         handler,
		_Observer:        nil,
		_Wg:              sync.WaitGroup{},
		RWMutex:          sync.RWMutex{},
	}

	//c._Addr = addr
	//c._Port = port
	//c._ConnInterval = connIntervalSec
	//c._ReadTimeOut = 0
	//c._PendingWriteNum = defaultPendingWriteNum
	//c._Handler = handler
	//c._IsClose = false
	////c._ServerInfo = &ClientServerInfo{}
	////c._ServerInfo.Init(-1, -1)

	return &c
}

func (c *TCPClient) SetPendingWriteNum(n int) {
	c.Lock()
	defer c.Unlock()
	c._PendingWriteNum = n
}

func (c *TCPClient) SetReadTimeOut(d time.Duration) {
	c.Lock()
	defer c.Unlock()
	c._ReadTimeOut = d
}

//func (c *TCPClient) SetServerInfo(kind, idx int) {
//	c._ServerInfo._Kind = kind
//	c._ServerInfo._Idx = idx
//}
//
//func (c *TCPClient) ServerInfo() *ClientServerInfo {
//	return c._ServerInfo
//}

func (c *TCPClient) SetObserver(ob interface{}) {
	c.Lock()
	defer c.Unlock()
	c._Observer = ob
}

func (c *TCPClient) Observer() interface{} {
	c.RLock()
	defer c.RUnlock()
	return c._Observer
}

func (c *TCPClient) Start() {
	go c.run()
}

func (c *TCPClient) dial() net.Conn {
	c.Lock()
	defer c.Unlock()

	if c._IsClose {
		return nil
	}

	// 접속 시도
	conn, err := net.Dial("tcp", c._Addr+":"+strconv.Itoa(c._Port))
	if err == nil {
		return conn
	}
	return nil
}

func (c *TCPClient) run() {
	c._Wg.Add(1)
	defer c._Wg.Done()

	// 주기적으로 재접속을 시도한다
	for {
		conn := c.dial()
		if conn == nil {
			c.Lock()
			if c._IsClose {
				c.Unlock()
				return
			}
			c.Unlock()
			//fmt.Println("re-connect to server!!!")
			time.Sleep(c._ConnInterval)
			continue
		}

		c.Lock()
		if c._IsClose {
			c.Unlock()
			if err := conn.Close(); nil != err {
				logger.Errorf("occur error(%v)", err)
			}
			return
		}
		// 세션 객체 생성
		c._Session = NewSession(conn, c._PendingWriteNum, c._ReadTimeOut)
		c._Session.SetObserver(c._Observer)
		c.Unlock()

		logger.Protect(func() {
			// 접속 처리 함수
			c._Handler.OnOpen(c._Session)
			//JobPush(OnOpen{c._Session, c._Handler})

			// Worker 실행
			doWorker(c._Session, c._Handler)

			// 접속 종료 처리
			//c._Handler.OnClose(c._Session)
			JobPush(OnClose{c._Session, c._Handler})
			<-c._Session._DoneChan
		})

		// cleanup
		c.Lock()
		c._Session.Close()
		c.Unlock()
	}
}

func (c *TCPClient) Session() *TCPSession {
	c.Lock()
	defer c.Unlock()
	return c._Session
}

func (c *TCPClient) IsConnected() bool {
	c.RLock()
	defer c.RUnlock()
	if nil == c._Session {
		return false
	}
	return c._Session.isConnected()
}

func (c *TCPClient) Write(pack *Packet) {
	c.Lock()
	defer c.Unlock()

	if nil == c._Session {
		DelPacket(pack)
		return
	}

	if !c._Session.isConnected() {
		//logger.Error("Session is disconnected!")
		DelPacket(pack)
		return
	}
	c._Session.Write(pack)
}

func (c *TCPClient) Close() {
	c.Lock()
	c._IsClose = true
	if nil != c._Session {
		c._Session.Close()
	}
	c.Unlock()

	// goroutine 종료할때 까지 대기
	c._Wg.Wait()

	logger.CInfo(logger.FG_LIGHT_YELLOW, "[TCPClient] Close.")
}
