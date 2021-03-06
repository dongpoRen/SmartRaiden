package network

import (
	"testing"

	"time"

	"github.com/ethereum/go-ethereum/crypto"
)

func init() {
	InitIceTransporter("182.254.155.208:3478", "bai", "bai", "182.254.155.208:5222")
}

type testreceiver struct {
	data chan []byte
}

func (this *testreceiver) Receive(data []byte, host string, port int) {
	this.data <- data
}
func TestNewIceTransporter(t *testing.T) {
	var err error
	k1, _ := crypto.GenerateKey()
	//addr1 := crypto.PubkeyToAddress(k1.PublicKey)
	k2, _ := crypto.GenerateKey()
	addr2 := crypto.PubkeyToAddress(k2.PublicKey)
	it1, _ := NewIceTransporter(k1, "client1")
	it1.Start()

	it2, _ := NewIceTransporter(k2, "client2")
	tr2 := testreceiver{make(chan []byte, 1)}
	it2.RegisterProtocol(&tr2)
	it2.Start()

	err = it1.Send(addr2, "", 0, []byte("data from addr2 to addr1"))
	if err != nil {
		t.Logf("send to addr2 err:%s", err)
		return
	}

	select {
	case <-time.After(time.Second * 50):
		t.Error("receive timeout")
	case data := <-tr2.data:
		t.Log("addr2 received ", string(data))
	}
	it2.Stop()
	it1.Stop()
}
