package smartraiden

import (
	"bytes"
	"fmt"
	"path/filepath"

	"io/ioutil"

	"strings"

	"errors"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

var noSuchAddress = errors.New("can not found this address")

/*
List All Accounts in directory KeyPath
*/
type AccountManager struct {
	KeyPath  string
	Accounts []accounts.Account
}

func NewAccountManager(keyPath string) (mgr *AccountManager) {
	mgr = &AccountManager{
		KeyPath: keyPath,
	}
	ks := keystore.NewKeyStore(keyPath, keystore.StandardScryptN, keystore.StandardScryptP)
	mgr.Accounts = ks.Accounts()
	return
}

func (this *AccountManager) AddressInKeyStore(addr common.Address) bool {
	for _, acc := range this.Accounts {
		if bytes.Equal(acc.Address[:], addr[:]) {
			return true
		}
	}
	return false
}

/*
Find the keystore file for an account, unlock it and get the private key
   addr: The Ethereum address for which to find the keyfile in the system
	password: Mostly for testing purposes. A password can be provided
			  as the function argument here. If it's not then the
              user is interactively queried for one.
    return The private key associated with the address
*/
func (this *AccountManager) GetPrivateKey(addr common.Address, password string) (privKeyBin []byte, err error) {
	if !this.AddressInKeyStore(addr) {
		err = noSuchAddress
		return
	}
	addrhex := strings.ToLower(addr.Hex())
	filename := fmt.Sprintf("UTC--*%s", addrhex[2:]) //skip 0x
	path := filepath.Join(this.KeyPath, filename)
	files, err := filepath.Glob(path)
	if err != nil {
		return
	}
	keyjson, _ := ioutil.ReadFile(files[0])
	key, err := keystore.DecryptKey(keyjson, password)
	if err != nil {
		return
	}
	privKeyBin = crypto.FromECDSA(key.PrivateKey)
	return
}
