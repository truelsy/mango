package mango

import (
	logger "TCPEx/Engine/Logger"
	"reflect"
	"time"
)

type Ticker struct {
	_Tick     *time.Ticker
	_DirectFn func()
	_Observer TimerHandler
	_Done     chan bool
	//_Pause    bool
	//sync.RWMutex
}

func NewTicker(dur time.Duration, arg interface{}) *Ticker {
	ticker := new(Ticker)
	ticker._Tick = time.NewTicker(dur)
	ticker._Done = make(chan bool)
	//ticker._Pause = false
	//ticker.RWMutex = sync.RWMutex{}

	t := reflect.TypeOf(arg)
	switch t.Kind() {
	case reflect.Func:
		ticker._DirectFn = reflect.ValueOf(arg).Interface().(func())
	case reflect.Struct, reflect.Ptr:
		// TimerHandler 를 구현했는지 확인
		if t.Implements(reflect.TypeOf((*TimerHandler)(nil)).Elem()) {
			ticker._Observer = reflect.ValueOf(arg).Interface().(TimerHandler)
		} else {
			logger.Errorf("can't implements struct or pointer! name(%v)", t)
			return nil
		}
	default:
		logger.Errorf("not supported argument! type(%v)", t.Kind())
		return nil
	}

	return ticker
}

// Run dur 주기로 계속 실행
func (t *Ticker) Run() {
	go func() {
		for {
			select {
			case <-t._Tick.C:
				//if !t.isPause() {
				JobPush(OnTimer{CallBack: t._DirectFn, TimerHandler: t._Observer})
				//}
			case <-t._Done:
				return
			}
		}
	}()
}

// RunOnce 한번 실행 후 종료
func (t *Ticker) RunOnce() {
	go func() {
		select {
		case <-t._Tick.C:
			JobPush(OnTimer{CallBack: t._DirectFn, TimerHandler: t._Observer})
			t.Stop()
			return
		case <-t._Done:
			return
		}
	}()
}

func (t *Ticker) Stop() {
	t._Tick.Stop()
	close(t._Done)
}

//func (t *Ticker) Pause() {
//	t.Lock()
//	defer t.Unlock()
//	t._Pause = true
//}
//
//func (t *Ticker) Resume() {
//	t.Lock()
//	defer t.Unlock()
//	t._Pause = false
//}
//
//func (t *Ticker) isPause() bool {
//	t.RLock()
//	defer t.RUnlock()
//	if t._Pause {
//		return true
//	}
//	return false
//}

//func SetTimerOnce(dur time.Duration, fn func()) *time.Timer {
//	// Timer (dur 후 한번실행)
//	return time.AfterFunc(dur, fn)
//}
//
//func SetTimerNextMinute(baseTime time.Time, dur time.Duration, fn func()) *time.Timer {
//	// Timer (baseTime 기준으로 dur 후에 실행)
//	goalTime := baseTime.Add(dur)
//	//goalTime = goalTime.Add(-1 * time.Second * time.Duration(goalTime.Second()))
//	//goalTime = goalTime.Add(-1 * time.Duration(goalTime.Nanosecond()))
//
//	// 실행해야 할 시간이 이미 지났다.
//	if goalTime.Before(baseTime) {
//		return nil
//	}
//
//	return time.AfterFunc(goalTime.Sub(baseTime), fn)
//}

//func SetTicker(dur time.Duration, fn func()) *time.Ticker {
//	// Ticker (dur 주기로 계속 실행)
//	ticker := time.NewTicker(dur)
//	go func() {
//		for range ticker.C {
//			fn()
//		}
//		fmt.Println("Stop goroutine!")
//	}()
//	return ticker
//}
