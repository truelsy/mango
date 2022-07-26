package mango

import (
	logger "TCPEx/Engine/Logger"
	"reflect"
	"sync"
)

//type OnOpen struct {
//	*TCPSession
//	Handler
//}
//
//type OnClose struct {
//	*TCPSession
//	Handler
//}

type OnPacket struct {
	*TCPSession
	*Packet
	TCPHandler
}

type OnTimer struct {
	CallBack func()
	TimerHandler
}

type OnClose struct {
	*TCPSession
	TCPHandler
}

type Dispatcher struct {
	_JobChan chan interface{}
	_Wg      sync.WaitGroup
}

func (d *Dispatcher) init() {
	d._JobChan = make(chan interface{})

	// Dispatcher 시작!
	d._Wg.Add(1)
	go func() {
		d.doRun()
		d._Wg.Done()
	}()
}

func (d *Dispatcher) doRun() {
	for job := range d._JobChan {
		o := reflect.ValueOf(job)
		v := reflect.Indirect(o)
		logger.Protect(func() {
			switch t := v.Interface().(type) {
			//case OnOpen:
			//	t.Handler.OnOpen(t.TCPSession)
			case OnClose:
				t.TCPHandler.OnClose(t.TCPSession)
			case OnPacket:
				t.TCPHandler.OnPacket(t.TCPSession, t.Packet)
				//DelPacket(t.Packet)
			case OnTimer:
				if nil != t.CallBack {
					t.CallBack()
				} else {
					t.TimerHandler.OnTime()
				}
			}
		})

		// 패닉이 발생하는 경우 Packet이 반환되지 않으므로
		// 여기서 반환해 준다.
		switch t := v.Interface().(type) {
		case OnPacket:
			DelPacket(t.Packet)
		}
	}
}

func (d *Dispatcher) Close() {
	close(d._JobChan)
	d._Wg.Wait()
	logger.CInfo(logger.FG_LIGHT_YELLOW, "[Dispatcher] Close.")
}
