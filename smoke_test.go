package smartraiden

import (
	"testing"

	"context"
	"math/big"
	"math/rand"

	"time"

	"fmt"

	"github.com/SmartMeshFoundation/SmartRaiden/abi/bind"
	"github.com/SmartMeshFoundation/SmartRaiden/network/rpc"
	"github.com/SmartMeshFoundation/SmartRaiden/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/fatedier/frp/src/utils/log"
	assert2 "github.com/stretchr/testify/assert"
)

var big500 = big.NewInt(500)
var x = big.NewInt(0)

func assert(t *testing.T, expected, actual interface{}, msgAndArgs ...interface{}) bool {
	return assert2.EqualValues(t, expected, actual, msgAndArgs...)
}
func deployAToken(t *testing.T, raiden *RaidenService) (addr common.Address) {
	n := new(big.Int)
	n.SetBytes(raiden.NodeAddress[:])
	addr, tx, _, err := rpc.DeployHumanStandardToken(raiden.Chain.Auth, raiden.Chain.Client, n, "Contracts in Go!!!", 0, "Go!")
	if err != nil {
		t.Error(err)
		t.FailNow()
		return
	}

	_, err = bind.WaitDeployed(context.Background(), raiden.Chain.Client, tx)
	if err != nil {
		t.Error(err)
		t.FailNow()
		return
	}
	return
}
func testNewToken(t *testing.T, ra, rb, rc, rd *RaidenApi) (tokenAddr common.Address) {
	tokenAddr = deployAToken(t, ra.Raiden)
	token := ra.Raiden.Chain.Token(tokenAddr)
	assert(t, token.Transfer(rb.Raiden.NodeAddress, big500), nil)
	assert(t, token.Transfer(rc.Raiden.NodeAddress, big500), nil)
	assert(t, token.Transfer(rd.Raiden.NodeAddress, big500), nil)
	log.Info("step 2. register")
	_, err := ra.RegisterToken(tokenAddr)
	if err != nil {
		t.Error(err)
		t.FailNow()
		return
	}
	return
}
func testCreateChannel(t *testing.T, tokenAddr common.Address, contractBalance *big.Int, ra, rb, rc *RaidenApi) {
	var err error
	_, err = ra.Open(tokenAddr, rb.Raiden.NodeAddress, ra.Raiden.Config.SettleTimeout, ra.Raiden.Config.RevealTimeout)
	if err != nil {
		t.Error(err)
		t.FailNow()
		return
	}
	assert(t, ra.Deposit(tokenAddr, rb.Raiden.NodeAddress, contractBalance, time.Minute), nil)
	assert(t, rb.Deposit(tokenAddr, ra.Raiden.NodeAddress, contractBalance, time.Minute), nil)

	log.Info("step 3.2 channel B-C")
	_, err = rb.Open(tokenAddr, rc.Raiden.NodeAddress, ra.Raiden.Config.SettleTimeout, ra.Raiden.Config.RevealTimeout)
	if err != nil {
		t.Error(err)
		t.FailNow()
		return
	}
	assert(t, rb.Deposit(tokenAddr, rc.Raiden.NodeAddress, contractBalance, time.Minute), nil)
	assert(t, rc.Deposit(tokenAddr, rb.Raiden.NodeAddress, contractBalance, time.Minute), nil)
}
func newEnv(t *testing.T, ra, rb, rc, rd *RaidenApi) (addr1, addr2 common.Address) {
	var contractBalance = big.NewInt(100)
	tokenAddr := testNewToken(t, ra, rb, rc, rd)
	testCreateChannel(t, tokenAddr, contractBalance, ra, rb, rc)
	tokenAddr2 := testNewToken(t, ra, rb, rc, rd)
	testCreateChannel(t, tokenAddr2, contractBalance, ra, rb, rc)
	log.Info(fmt.Sprintf("newEnv tokenAddr1=%s,tokenAddr2=%s", tokenAddr.String(), tokenAddr2.String()))
	log.Info("create two tokens ,each token has tow channels a-b and b-c , each channel has 100 balance")
	return tokenAddr, tokenAddr2
}
func TestSmoke(t *testing.T) {
	var err error
	ra, rb, rc, rd := makeTestRaidenApis()
	log.Info("step 1. build env for test")
	var tokenAddr, tokenAddr2 common.Address
	var contractBalance = big.NewInt(100)
	var tAmount = big.NewInt(1)
	if true {
		tokenAddr, tokenAddr2 = newEnv(t, ra, rb, rc, rd)
	} else {
		tokenAddr = common.HexToAddress("0x088015E873D8C94ac1bf3731198309E25683Cc9E")
		tokenAddr2 = common.HexToAddress("0xF3AdEde8030D33d6B360e7d0FE08E5e4c1425c8C")
		time.Sleep(time.Second) //let ra,rb,rc,rd udpate channel info
		log.Info("channels about token1")
		ra.Raiden.Token2ChannelGraph[tokenAddr].PrintGraph()
		log.Info("channels about token2")
		ra.Raiden.Token2ChannelGraph[tokenAddr].PrintGraph()
	}

	log.Info("step 2 transfer from A to B")
	err = ra.Transfer(tokenAddr, tAmount, utils.BigInt0, rb.Raiden.NodeAddress, rand.New(rand.NewSource(time.Now().UnixNano())).Uint64(), time.Minute)
	if err != nil {
		t.Error(err)
		return
	}
	//let rb finish transfer
	time.Sleep(time.Second * 5)
	//channel a-b of tokenaddr
	assert(t, ra.Raiden.GetChannel(tokenAddr, rb.Raiden.NodeAddress).Balance(), x.Sub(contractBalance, tAmount))
	assert(t, rb.Raiden.GetChannel(tokenAddr, ra.Raiden.NodeAddress).Balance(), x.Add(contractBalance, tAmount))

	log.Info("step 3 transfer from A to C")
	err = ra.Transfer(tokenAddr, tAmount, utils.BigInt0, rc.Raiden.NodeAddress, rand.New(rand.NewSource(time.Now().UnixNano())).Uint64(), time.Minute)
	if err != nil {
		t.Error(err)
		return
	}
	time.Sleep(time.Second * 5) //let rb,rc to update
	//channel a-b of tokenaddr
	assert(t, ra.Raiden.GetChannel(tokenAddr, rb.Raiden.NodeAddress).Balance(), x.Sub(contractBalance, tAmount).Sub(x, tAmount))
	assert(t, rb.Raiden.GetChannel(tokenAddr, ra.Raiden.NodeAddress).Balance(), x.Add(contractBalance, tAmount).Add(x, tAmount))
	//channel b-c of tokenaddr
	assert(t, rb.Raiden.GetChannel(tokenAddr, rc.Raiden.NodeAddress).Balance(), x.Sub(contractBalance, tAmount))
	assert(t, rc.Raiden.GetChannel(tokenAddr, rb.Raiden.NodeAddress).Balance(), x.Add(contractBalance, tAmount))

	log.Info("step 4 D connect to this token network")
	if false {
		err = rd.ConnectTokenNetwork(tokenAddr, big.NewInt(300), 3, 0.4)
		if err != nil {
			t.Error(err)
			return
		}
	}
	log.Info(" step 5 make a token swap between A and B")
	log.Info(fmt.Sprintf("a:a-b token1=%d,token2=%d", ra.Raiden.GetChannel(tokenAddr, rb.Raiden.NodeAddress).Balance(), ra.Raiden.GetChannel(tokenAddr2, rb.Raiden.NodeAddress).Balance()))
	log.Info(fmt.Sprintf("b:a-b token1=%d,token2=%d", rb.Raiden.GetChannel(tokenAddr, ra.Raiden.NodeAddress).Balance(), rb.Raiden.GetChannel(tokenAddr2, ra.Raiden.NodeAddress).Balance()))
	err = rb.ExpectTokenSwap(32, tokenAddr, tokenAddr2, ra.Raiden.NodeAddress, rb.Raiden.NodeAddress, tAmount, x.Add(tAmount, tAmount))
	if err != nil {
		t.Error(err)
		return
	}
	err = ra.TokenSwapAndWait(32, tokenAddr, tokenAddr2, ra.Raiden.NodeAddress, rb.Raiden.NodeAddress, tAmount, x.Add(tAmount, tAmount))
	if err != nil {
		t.Error(err)
		return
	}
	//how to know finish of taker?
	time.Sleep(time.Second * 12) //let ra,rb udpate data ,short time will error

	//channel a-b of tokenaddr a-amount b+amount
	assert(t, ra.Raiden.GetChannel(tokenAddr, rb.Raiden.NodeAddress).Balance(), x.Sub(contractBalance, x.Mul(tAmount, big.NewInt(3))))
	assert(t, rb.Raiden.GetChannel(tokenAddr, ra.Raiden.NodeAddress).Balance(), x.Add(contractBalance, x.Mul(tAmount, big.NewInt(3))))

	//channel a-b of tokenadd4 a+amount*2 b-amount*2
	assert(t, ra.Raiden.GetChannel(tokenAddr2, rb.Raiden.NodeAddress).Balance(), x.Add(contractBalance, x.Mul(tAmount, big.NewInt(2))))
	assert(t, rb.Raiden.GetChannel(tokenAddr2, ra.Raiden.NodeAddress).Balance(), x.Sub(contractBalance, x.Mul(tAmount, big.NewInt(2))))

	log.Info(" step 6 make a token swap between A and c through b")
	err = rc.ExpectTokenSwap(33, tokenAddr, tokenAddr2, ra.Raiden.NodeAddress, rc.Raiden.NodeAddress, tAmount, x.Add(tAmount, tAmount))
	if err != nil {
		t.Error(err)
		return
	}
	err = ra.TokenSwapAndWait(33, tokenAddr, tokenAddr2, ra.Raiden.NodeAddress, rc.Raiden.NodeAddress, tAmount, x.Add(tAmount, tAmount))
	if err != nil {
		t.Error(err)
		return
	}
	//how to know finish of taker?
	time.Sleep(time.Second * 12) //let ra,rb ,rcudpate data ,short time will error

	//channel a-b of tokenaddr a-amount b+amount
	assert(t, ra.Raiden.GetChannel(tokenAddr, rb.Raiden.NodeAddress).Balance(), x.Sub(contractBalance, x.Mul(tAmount, big.NewInt(4))))
	assert(t, rb.Raiden.GetChannel(tokenAddr, ra.Raiden.NodeAddress).Balance(), x.Add(contractBalance, x.Mul(tAmount, big.NewInt(4))))
	//channel b-c of tokenaddr b-amount c+amount
	assert(t, rb.Raiden.GetChannel(tokenAddr, rc.Raiden.NodeAddress).Balance(), x.Sub(contractBalance, x.Mul(tAmount, big.NewInt(2))))
	assert(t, rc.Raiden.GetChannel(tokenAddr, rb.Raiden.NodeAddress).Balance(), x.Add(contractBalance, x.Mul(tAmount, big.NewInt(2))))

	//channel a-b of tokenaddr2 a+2*amount b-2*amount
	assert(t, ra.Raiden.GetChannel(tokenAddr2, rb.Raiden.NodeAddress).Balance(), x.Add(contractBalance, x.Mul(tAmount, big.NewInt(4))))
	assert(t, rb.Raiden.GetChannel(tokenAddr2, ra.Raiden.NodeAddress).Balance(), x.Sub(contractBalance, x.Mul(tAmount, big.NewInt(4))))
	//channel b-c of tokenaddr2 b+2amount c-2*amount
	assert(t, rb.Raiden.GetChannel(tokenAddr2, rc.Raiden.NodeAddress).Balance(), x.Add(contractBalance, x.Mul(tAmount, big.NewInt(2))))
	assert(t, rc.Raiden.GetChannel(tokenAddr2, rb.Raiden.NodeAddress).Balance(), x.Sub(contractBalance, x.Mul(tAmount, big.NewInt(2))))
	log.Info(" step 8 test leave network take a long long time")
	if false {
		_, err = rd.LeaveTokenNetwork(tokenAddr, true)
		if err != nil {
			t.Error(err)
		}
	}
}

func TestFeeCharger(t *testing.T) {
	var err error
	policy := &ConstantFeePolicy{}
	ra, rb, rc, rd := makeTestRaidenApisWithFee(policy)
	log.Info("step 1. build env for test")
	var tokenAddr, tokenAddr2 common.Address
	var contractBalance = big.NewInt(100)
	var tAmount = big.NewInt(1)
	if false {
		tokenAddr, tokenAddr2 = newEnv(t, ra, rb, rc, rd)
	} else {
		tokenAddr = common.HexToAddress("0x883FF6D87eB3f0b6f9122E96cE01d9b508bEC2C9")
		tokenAddr2 = common.HexToAddress("0xd319EBa3d8237c8b72759f0BB368Fb0A31De7CcA")
		time.Sleep(time.Second) //let ra,rb,rc,rd udpate channel info
		log.Info("channels about token1")
		ra.Raiden.Token2ChannelGraph[tokenAddr].PrintGraph()
		log.Info("channels about token2")
		ra.Raiden.Token2ChannelGraph[tokenAddr].PrintGraph()
	}
	log.Info("tokenAddr=%s,tokenaddr2=%s", tokenAddr.String(), tokenAddr2.String())
	log.Info("transfer from A to C")
	err = ra.Transfer(tokenAddr, tAmount, utils.BigInt0, rc.Raiden.NodeAddress, rand.New(rand.NewSource(time.Now().UnixNano())).Uint64(), time.Minute)
	if err != nil {
		t.Error(err)
		return
	}
	//let rb finish transfer
	time.Sleep(time.Second * 3)
	abAmount := new(big.Int).Add(tAmount, policy.GetNodeChargeFee(rb.Raiden.NodeAddress, tokenAddr, tAmount))
	//channel a-b of tokenaddr
	assert(t, ra.Raiden.GetChannel(tokenAddr, rb.Raiden.NodeAddress).Balance(), x.Sub(contractBalance, abAmount))
	assert(t, rb.Raiden.GetChannel(tokenAddr, ra.Raiden.NodeAddress).Balance(), x.Add(contractBalance, abAmount))
	bcAmount := tAmount
	assert(t, rb.Raiden.GetChannel(tokenAddr, rc.Raiden.NodeAddress).Balance(), x.Sub(contractBalance, bcAmount))
	assert(t, rc.Raiden.GetChannel(tokenAddr, rb.Raiden.NodeAddress).Balance(), x.Add(contractBalance, bcAmount))

	//specifed a  wrong fee,
	err = ra.Transfer(tokenAddr, tAmount, big.NewInt(1), rc.Raiden.NodeAddress, rand.New(rand.NewSource(time.Now().UnixNano())).Uint64(), time.Minute)
	if err == nil {
		t.Error(fmt.Sprintf("should fail because of not engough fee."))
		return
	}
}
