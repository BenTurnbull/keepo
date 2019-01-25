package store

import (
	"keepo/src/crypto"
	"keepo/src/util"
	"path/filepath"
	"sort"
)

const storeName = "keepo.dat"

func GetStorePath(path string) string {
	return path + string(filepath.Separator) + storeName
}

func GetMapKeysV1(path string) (keys []string) {
	storePath := GetStorePath(path)
	dataMap := getV1(storePath)
	keys = make([]string, 0, len(dataMap))
	for key := range dataMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func GetMapValueV1(path, dataKey, password string) (value []byte, err error) {
	storePath := GetStorePath(path)
	dataMap := getV1(storePath)

	dataBytes := dataMap[dataKey]
	util.CheckState(len(dataBytes) > 0, "name not found")

	passwordHash := crypto.GetHash([]byte(password))
	return crypto.DecryptCFB(passwordHash[:], dataBytes)
}

func SetMapValueV1(path, dataKey, dataValue, password string) (err error) {
	storePath := GetStorePath(path)
	dataMap := getV1(storePath)

	passwordHash := crypto.GetHash([]byte(password))
	dataBytes, err := crypto.EncryptCFB(passwordHash[:], []byte(dataValue))
	util.CheckError(err, "could not encrypt data value")
	dataMap[dataKey] = dataBytes

	return setV1(storePath, dataMap)
}

/*func convert(path, existingKey, newKey string) {
	storePath := GetStorePath(path)
	dataMap := getV1(storePath)
	convertedDataMap := make(map[string][]byte)
	key := crypto.GetHash([]byte(existingKey))

	for k, v := range dataMap {
		value, err := crypto.DecryptCFB(key[:], v)
		util.CheckError(err, "could not decrypt store value")
		convertedDataMap[k] = value
	}

	fixedKey := crypto.GenerateKey()
	log.Printf("fixed key %s", hex.EncodeToString(fixedKey[:]))

	var nonce [crypto.NonceSize]byte
	for k, v := range dataMap {
		nonce = crypto.GenerateNonce()
		out := make([]byte, len(nonce))
		copy(out, nonce[:])

		out = secretbox.Seal(out, v, &nonce, &fixedKey)
		convertedDataMap[k] = out
	}

	nonce = crypto.GenerateNonce()
	out := secretbox.Seal(nonce[:], fixedKey[:], &nonce, &key)
	err := set(out, convertedDataMap)
	util.CheckError(err, "could not save converted store")
}*/

/*func GetMapValueV2(path, dataKey, password string) (value []byte, err error) {
	storePath := GetStorePath(path)
	fixedKey, data, err := get(storePath)
	util.CheckError(err, "could not read store")

	var nonce [crypto.NonceSize]byte
	copy(nonce[:], fixedKey[:crypto.NonceSize])

	key := crypto.GetHash([]byte(input.ReadPassword()))
	out, ok := secretbox.Open(nil, fixedKey, &nonce, &key)
	util.CheckState(ok, "could not decrypt fixed key")

	log.Printf("fixed key %s", hex.EncodeToString(out[:]))

	dataValue := data[dataKey]
	util.CheckState(len(dataValue) > 0, "name not found")

	//todo
	return nil, nil
}*/
