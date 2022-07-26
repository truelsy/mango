package mango_test

import (
	"TCPEx/Engine/mango"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	mango.Init(10000)
	defer mango.Close()
	os.Exit(m.Run())
}

func BenchmarkPacketPool(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StartTimer()
		pack := mango.NewPacket()
		mango.DelPacket(pack)
		b.StopTimer()
	}
}
