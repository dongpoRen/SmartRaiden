package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
)

// key for test
const passkey = "838e2Bf510eC7Ff49CC607b718Ce8401"

func PasswordEncrypt(pass string) (encstr string, err error) {
	key, _ := hex.DecodeString(passkey)
	encdata, err := Encrypt([]byte(pass), key)
	if err != nil {
		return
	}
	encstr = base64.RawStdEncoding.EncodeToString(encdata)
	return
}
func PasswordDecrypt(encpass string) (pass string, err error) {
	key, _ := hex.DecodeString(passkey)
	decdata, err := base64.RawStdEncoding.DecodeString(encpass)
	if err != nil {
		return
	}
	plaindata, err := Decrypt(decdata, key)
	if err != nil {
		return
	}
	pass = string(plaindata)
	return
}

// AES加密
func Encrypt(src, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic("aes.NewCipher: " + err.Error())
	}
	encryptText := make([]byte, aes.BlockSize+len(src))
	iv := encryptText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(encryptText[aes.BlockSize:], src)

	return encryptText, nil

}

// AES解密
func Decrypt(src, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic("aes.NewCipher: " + err.Error())
	}
	// 长度不能小于aes.Blocksize
	if len(src) < aes.BlockSize {
		return nil, errors.New("crypto/cipher: ciphertext too short")
	}

	iv := src[:aes.BlockSize]
	src = src[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(src, src)
	return src, nil
}
