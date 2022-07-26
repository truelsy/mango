package mango

var gDispatcher = new(Dispatcher)
var gPacketPool = new(PacketPool)

func Init(packetPoolCnt int) {
	gDispatcher.init()
	gPacketPool.init(packetPoolCnt)
}

func Close() {
	gPacketPool.Close()

	// Dispatcher를 제일 마지막에  Close!
	gDispatcher.Close()
}

func NewPacket() *Packet {
	pack := gPacketPool.Get().(*Packet)
	pack.NewTick()
	return pack
}

func DelPacket(pack *Packet) {
	if pack != nil {
		gPacketPool.Put(pack)
	}
}

func PacketPoolLen() (int32, int32) {
	return gPacketPool.Len()
}

func JobPush(job interface{}) {
	if nil == job {
		return
	}
	gDispatcher._JobChan <- job
}
