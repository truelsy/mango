package mango

import (
	logger "TCPEx/Engine/Logger"
	"TCPEx/Engine/Utils"
	"time"
)

type PacketPool struct {
	//_ReturnChan chan *Packet
	//_Wg         sync.WaitGroup
	*Utils.SyncPool
}

func (p *PacketPool) init(poolCnt int) {
	//p._ReturnChan = make(chan *Packet)
	//p._Wg = sync.WaitGroup{}
	p.SyncPool = Utils.CreatePool("PacketPool", poolCnt, func() Utils.Pooler {
		pack := new(Packet)
		pack._Buf = make([]byte, bufferSize)
		pack._ReadOffset = GetHeaderLen()
		pack._WriteOffset = GetHeaderLen()
		pack._RefCount = 0
		pack._Tick = time.Time{}
		return pack
	})

	//p._Wg.Add(1)
	//go func() {
	//	p.doReturn()
	//	p._Wg.Done()
	//}()
}

//func (p *PacketPool) doReturn() {
//	for pack := range p._ReturnChan {
//		if nil == pack {
//			continue
//		}
//		// Pool 에 반환
//		p.Put(pack)
//	}
//}

func (p *PacketPool) Close() {
	//close(p._ReturnChan)
	//p._Wg.Wait()
	p.SyncPool.Close()
	logger.CInfo(logger.FG_LIGHT_YELLOW, "[PacketPool] Close.")
}
