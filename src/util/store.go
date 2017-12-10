package util

import (
	"bufio"
	"encoding/binary"
	"os"
	"io"
	"path/filepath"
)

const storeName = "keepo.dat"

type Store struct {
	Path string
}

func getStorePath(store Store) string {
	return store.Path + string(filepath.Separator) + storeName
}

func (store Store) GetDataMap() (map[string][]byte) {

	dataMap := make(map[string][]byte)

	// open input file
	storePath := getStorePath(store)
	if fi, err := os.Open(storePath); err == nil {

		// close fi on exit and check for its returned error
		defer func() {
			if err := fi.Close(); err != nil {
				panic(err)
			}
		}()

		// make a read buffer
		r := bufio.NewReader(fi)

		for true {
			// read key
			keyLength := new(uint32)
			err = binary.Read(r, binary.LittleEndian, keyLength)
			if err == io.EOF {
				return dataMap
			}
			CheckError(err)

			keyLengthInt := int(*keyLength)
			CheckState(keyLengthInt > 0, "key length may not be negative")

			keyNameBytes := make([]byte, keyLengthInt)
			err = binary.Read(r, binary.LittleEndian, keyNameBytes)
			CheckError(err)

			keyName := string(keyNameBytes)

			// read value
			valueLength := new(uint32)
			err = binary.Read(r, binary.LittleEndian, valueLength)
			CheckError(err)

			valueLengthInt := int(*valueLength)
			CheckState(valueLengthInt > 0, "value Length may not be negative")

			valueBytes := make([]byte, valueLengthInt)
			err = binary.Read(r, binary.LittleEndian, valueBytes)
			CheckError(err)

			dataMap[keyName] = valueBytes
		}
	}

	return dataMap
}

func (store Store) SetDataMap(dataMap map[string][]byte) error {

	// open output file
	storePath := getStorePath(store)
	if fo, err := os.Create(storePath); err == nil {

		// close fo on exit and check for its returned error
		defer func() {
			if err := fo.Close(); err != nil {
				panic(err)
			}
		}()

		// make a write buffer
		w := bufio.NewWriter(fo)

		for k, v := range dataMap {

			kLen := len(k)
			err = binary.Write(w, binary.LittleEndian, uint32(kLen))
			CheckError(err)

			err = binary.Write(w, binary.LittleEndian, []byte(k))
			CheckError(err)

			vLen := len(v)
			err = binary.Write(w, binary.LittleEndian, uint32(vLen))
			CheckError(err)

			err = binary.Write(w, binary.LittleEndian, []byte(v))
			CheckError(err)

			w.Flush()
		}

	} else {
		return err
	}

	return nil
}



