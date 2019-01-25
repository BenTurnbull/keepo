package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"keepo/src/util"
)

const(
	HashSize = 32
	KeySize = 32
	NonceSize = 24
)

func GenerateKey() (key [KeySize]byte) {
	_, err := io.ReadFull(rand.Reader, key[:])
	util.CheckError(err, "could not generate key")
	return key
}

func GenerateNonce() (nonce [NonceSize]byte) {
	_, err := io.ReadFull(rand.Reader, nonce[:])
	util.CheckError(err, "could not generate nonce")
	return nonce
}

func GetHash(content []byte) (hash [HashSize]byte) {
	sha256Function := sha256.New()
	sha256Function.Write(content)
	hashBytes := sha256Function.Sum(nil)
	util.CheckState(len(hashBytes) == HashSize, "hash output was not hash length")
	copy(hash[:], hashBytes)
	return hash
}

func EncryptCFB(key, text []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	b := base64.StdEncoding.EncodeToString(text)
	cipherText := make([]byte, aes.BlockSize+len(b))

	// randomly generated initialization vector
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(cipherText[aes.BlockSize:], []byte(b))
	return cipherText, nil
}

func DecryptCFB(key, text []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(text) < aes.BlockSize {
		return nil, errors.New("text too short")
	}
	iv := text[:aes.BlockSize]
	text = text[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(text, text)
	data, err := base64.StdEncoding.DecodeString(string(text))
	if err != nil {
		return nil, err
	}
	return data, nil
}