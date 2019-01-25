package store

import (
	"bufio"
	"encoding/binary"
	"io"
	"keepo/src/util"
	"os"
)

func getV1(path string) map[string][]byte {

	dataMap := make(map[string][]byte)

	// open input file
	if fi, err := os.Open(path); err == nil {

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
				break
			}
			util.CheckError(err, "could not read key length")
			keyLengthInt := uint32(*keyLength)

			keyNameBytes := make([]byte, keyLengthInt)
			err = binary.Read(r, binary.LittleEndian, keyNameBytes)
			util.CheckError(err, "could not read key")

			keyName := string(keyNameBytes)

			// read value
			valueLength := new(uint32)
			err = binary.Read(r, binary.LittleEndian, valueLength)
			util.CheckError(err, "could not read value length")
			valueLengthInt := uint32(*valueLength)

			valueBytes := make([]byte, valueLengthInt)
			err = binary.Read(r, binary.LittleEndian, valueBytes)
			util.CheckError(err, "could not read value")

			dataMap[keyName] = valueBytes
		}
	}

	return dataMap
}

func setV1(path string, dataMap map[string][]byte) error {

	// open output file
	if fo, err := os.Create(path); err == nil {

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
			util.CheckError(err, "could not write key length")

			err = binary.Write(w, binary.LittleEndian, []byte(k))
			util.CheckError(err, "could not write key")

			vLen := len(v)
			err = binary.Write(w, binary.LittleEndian, uint32(vLen))
			util.CheckError(err, "could not write value length")

			err = binary.Write(w, binary.LittleEndian, []byte(v))
			util.CheckError(err, "could not write value")

			err = w.Flush()
			util.CheckError(err, "could not flush writes")
		}

	} else {
		return err
	}

	return nil
}