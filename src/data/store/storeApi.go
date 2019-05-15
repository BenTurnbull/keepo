package store

import (
	"golang.org/x/crypto/nacl/secretbox"
	"io"
	"keepo/src/crypto"
	"keepo/src/util"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const storeName = "keepo.dat"

func GetStorePath(path string) string {
	return path + string(filepath.Separator) + storeName
}

func GetMapKeys(path string) (keys []string) {
	storePath := GetStorePath(path)
	_, dataMap, err := getIndex(storePath)
	if err == io.EOF {
		log.Println("empty data store")
		os.Exit(0)
	}
	util.CheckError(err, "could not access data")

	keys = make([]string, 0, len(dataMap))
	for key := range dataMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func GetMapValue(path, dataKey, secret string) (value []byte, err error) {
	storePath := GetStorePath(path)
	sealedSecret, dataIndex, err := getIndex(storePath)
	if _, ok := err.(*os.PathError); ok {
		return nil, ValueAbsentState
	}

	if err != nil {
		return nil, err
	}

	unsealedSecret, err := unsealSecret(secret, sealedSecret)
	if err != nil {
		return nil, err
	}

	dataOffset := dataIndex[dataKey]
	if dataOffset == 0 {
		return nil, ValueAbsentState
	}

	data, err := getData(storePath, dataOffset)
	util.CheckError(err, "could not read data")
	unsealedData := unsealData(data, unsealedSecret)

	return unsealedData, nil
}

func SetMapValue(path, dataKey, dataValue, secret string) (err error) {
	storePath := GetStorePath(path)
	sealedSecret, dataIndex, err := getIndex(storePath)

	if _, ok := err.(*os.PathError); ok {
		dataIndex = make(map[string]uint64)
		err = nil
		log.Println("starting new data store")
	}

	if err != nil {
		return err
	}

	// authenticate
	unsealedSecret, err := unsealSecret(secret, sealedSecret)
	if err != nil {
		return err
	}

	dataMap := make(map[string][]byte, len(dataIndex))
	for k, v := range dataIndex {
		if strings.Compare(k ,dataKey) == 0 {
			continue
		}

		data, err := getData(storePath, v)
		util.CheckError(err, "could not read data for entry: " + k)
		dataMap[k] = data
	}

	dataMap[dataKey] = sealData([]byte(dataValue), unsealedSecret)

	if sealedSecret == nil {
		sealedSecret = sealData(unsealedSecret[:], crypto.GetHash([]byte(secret)))
	}

	return set(storePath, sealedSecret, dataMap)
}

func ClearMapValue(path, dataKey, secret string) (err error) {

	storePath := GetStorePath(path)
	sealedSecret, dataIndex, err := getIndex(storePath)
	if _, ok := err.(*os.PathError); ok {
		return ValueAbsentState
	}

	if err != nil {
		return err
	}

	_, err = unsealSecret(secret, sealedSecret)
	if err != nil {
		return err
	}

	if _, ok := dataIndex[dataKey]; ok {
		delete(dataIndex, dataKey)

		dataMap := make(map[string][]byte, len(dataIndex))
		for k, v := range dataIndex {
			if strings.Compare(k ,dataKey) == 0 {
				continue
			}

			data, err := getData(storePath, v)
			util.CheckError(err, "could not read data for entry: " + k)
			dataMap[k] = data
		}

		// time to re-pack
		err := set(storePath, sealedSecret[:], dataMap)
		util.CheckError(err, "could not write data")
		return nil
	} else {
		return ValueAbsentState
	}
}

func unsealSecret(secret string, sealedSecret []byte) (unsealedSecret [crypto.SecretSize]byte, err error) {
	if sealedSecret == nil {
		return crypto.GenerateSecret(), nil
	}

	// verify the secret
	hashedSecret := crypto.GetHash([]byte(secret))
	var nonce [crypto.NonceSize]byte
	copy(nonce[:], sealedSecret[:crypto.NonceSize])

	out, ok := secretbox.Open(nil, sealedSecret[crypto.NonceSize:], &nonce, &hashedSecret)
	if !ok {
		return unsealedSecret, AuthenticationFailedState
	}

	util.CheckState(len(out) == crypto.SecretSize, "unsealed secret was not of key length")
	copy(unsealedSecret[:], out)
	return unsealedSecret, nil
}

func unsealData(sealedData []byte, secret [crypto.SecretSize]byte) (unsealedData []byte){
	var nonce [crypto.NonceSize]byte
	copy(nonce[:], sealedData[:crypto.NonceSize])
	unsealedData, ok := secretbox.Open(nil, sealedData[crypto.NonceSize:], &nonce, &secret)
	if !ok {
		log.Println("authentication failed")
		os.Exit(0)
	}
	return unsealedData
}

func sealData(data []byte, secret [crypto.SecretSize]byte) (sealedData []byte){
	var nonce = crypto.GenerateNonce()
	return secretbox.Seal(nonce[:], data, &nonce, &secret)
}

/* // conversion code for v1
func Convert(path, secret string) {
	keyMap := GetMapKeysV1(path)
	keyValueMap := make(map[string][]byte)

	for _, v := range keyMap {
		value, err := GetMapValueV1(path, v, secret)
		util.CheckError(err, "could not get data value")
		keyValueMap[v] = value
	}

	for k, v := range keyValueMap {
		log.Printf("%s - %s", k, v)
	}

	fixedSecret := crypto.GenerateSecret()
	fixedSecret = crypto.GetHash(fixedSecret[:])
	log.Printf("fixed key hash %s", hex.EncodeToString(fixedSecret[:]))

	var nonce [crypto.NonceSize]byte
	for k, v := range keyValueMap {
		nonce = crypto.GenerateNonce()
		out := secretbox.Seal(nonce[:], v, &nonce, &fixedSecret)
		keyValueMap[k] = out
	}

	log.Println("encrypted values")
	for k, v := range keyValueMap {
		log.Printf("%s - %s", k, hex.EncodeToString(v))
	}

	log.Println("decrypted values")
	for k, v := range keyValueMap {

		copy(nonce[:], v[:crypto.NonceSize])
		out, ok := secretbox.Open(nil, v[crypto.NonceSize:], &nonce, &fixedSecret)
		if !ok {
			log.Println("could not decrypt message")
		} else {
			log.Printf("%s - %s", k, string(out))
		}
	}

	nonce = crypto.GenerateNonce()
	newHashedSecret := crypto.GetHash([]byte(secret))
	sealedSecret := secretbox.Seal(nonce[:], fixedSecret[:], &nonce, &newHashedSecret)

	convertedPath := GetStorePath(path) + ".upgrade"
	err := set(convertedPath, sealedSecret[:], keyValueMap)
	util.CheckError(err, "could not write converted data")
}*/