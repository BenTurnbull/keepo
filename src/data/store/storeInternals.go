package store

import (
	"encoding/binary"
	"os"
)

/**
 * Store layout:
 *
 * header:
 * secret-length 		- uint32
 * secret-value  		- secret-length bytes
 *
 * index:
 * data-key-count		- uint32
 *
 * data-key-length		- uint32
 * data-key-value  		- data-key-length bytes
 * data-value-offset	- offset in bytes
 * ...
 *
 * data:
 * data-values:
 * data-value-length	- uint32
 * data-value  			- data-value-length bytes
 * ...
 *
 */

func getIndex(path string) (sealedSecret []byte, dataIndex map[string]uint64, err error) {

	// open input file
	if fi, err := os.Open(path); err == nil {

		// close fi on exit and check for its returned error
		defer func() {
			if err := fi.Close(); err != nil {
				panic(err)
			}
		}()

		uint32Bytes := make([]byte, 4)
		uint64Bytes := make([]byte, 8)

		// read the secret
		_, err := fi.Read(uint32Bytes)
		if err != nil {
			return nil, nil, InvalidFormatError("could not read secret length")
		}

		secretLength := binary.LittleEndian.Uint32(uint32Bytes)
		sealedSecret = make([]byte, secretLength)
		_, err = fi.Read(sealedSecret)
		if err != nil {
			return nil, nil, InvalidFormatError("could not read secret")
		}

		// read the index
		_, err = fi.Read(uint32Bytes)
		if err != nil {
			return nil, nil, InvalidFormatError("could not read index count")
		}

		indexCount := int(binary.LittleEndian.Uint32(uint32Bytes))
		dataIndex = make(map[string]uint64, indexCount)

		// read index entries
		for i := 0; i < indexCount; i++ {
			_, err = fi.Read(uint32Bytes)
			if err != nil {
				return sealedSecret, dataIndex, InvalidFormatError("could not read an index key length")
			}

			keyLength := int(binary.LittleEndian.Uint32(uint32Bytes))

			keyBytes := make([]byte, keyLength)
			_, err = fi.Read(keyBytes)
			if err != nil {
				return sealedSecret, dataIndex, InvalidFormatError("could not read an index key")
			}

			_, err = fi.Read(uint64Bytes)
			if err != nil {
				return sealedSecret, dataIndex, InvalidFormatError("could not read an index data offset")
			}

			dataOffset := binary.LittleEndian.Uint64(uint64Bytes)
			dataIndex[string(keyBytes)] = dataOffset
		}

		return sealedSecret, dataIndex, err
	} else {
		return nil, nil, err
	}
}

func getData(path string, dataOffset uint64) (data []byte, err error) {

	// open input file
	if fi, err := os.Open(path); err == nil {

		// close fi on exit and check for its returned error
		defer func() {
			if err := fi.Close(); err != nil {
				panic(err)
			}
		}()

		uint32Bytes := make([]byte, 4)

		_, err := fi.Seek(int64(dataOffset), 0)
		if err != nil {
			return nil, InvalidFormatError("could not seek to data offset")
		}

		_, err = fi.Read(uint32Bytes)
		if err != nil {
			return nil, InvalidFormatError("could not read data length")
		}
		dataLength := int(binary.LittleEndian.Uint32(uint32Bytes))

		data := make([]byte, dataLength)
		_, err = fi.Read(data)
		if err != nil {
			return nil, InvalidFormatError("could not read data")
		}

		return data, nil
	} else {
		return nil, err
	}
}

func set(path string, sealedSecret []byte, dataMap map[string][]byte) (err error) {

	// open output file
	if fo, err := os.Create(path); err == nil {

		// close fo on exit and check for its returned error
		defer func() {
			if err := fo.Close(); err != nil {
				panic(err)
			}
		}()

		uint32Bytes := make([]byte, 4)
		uint64Bytes := make([]byte, 8)

		// write the sealedSecret
		binary.LittleEndian.PutUint32(uint32Bytes, uint32(len(sealedSecret)))
		_, err = fo.Write(uint32Bytes)
		if err != nil {
			return InvalidFormatError("could not write secret length")
		}

		_, err = fo.Write(sealedSecret)
		if err != nil {
			return InvalidFormatError("could not write secret")
		}

		// write the header count
		entryCount := uint32(len(dataMap))
		binary.LittleEndian.PutUint32(uint32Bytes, entryCount)
		_, err = fo.Write(uint32Bytes)
		if err != nil {
			return InvalidFormatError("could not write key count")
		}

		// write the header
		headerMap := make(map[string]uint64, 0)
		for k := range dataMap {

			binary.LittleEndian.PutUint32(uint32Bytes, uint32(len(k)))
			_, err = fo.Write(uint32Bytes)
			if err != nil {
				return InvalidFormatError("could not write header entry length for: " + k)
			}

			_, err = fo.Write([]byte(k))
			if err != nil {
				return InvalidFormatError("could not write header entry for: " + k)
			}

			currentPosition, err := fo.Seek(0, 1)
			if err != nil {
				return InvalidFormatError("could not get current position of file")
			}

			headerMap[k] = uint64(currentPosition)

			binary.LittleEndian.PutUint64(uint64Bytes, uint64(0))
			_, err = fo.Write(uint64Bytes)
			if err != nil {
				return InvalidFormatError("could not write value position placeholder")
			}
		}

		// write the data and update the header value offsets
		for k, v := range dataMap {

			// track current value position
			currentPosition, err := fo.Seek(0, 1)
			if err != nil {
				return InvalidFormatError("could not get current position of file")
			}

			_, err = fo.Seek(int64(headerMap[k]), 0)
			binary.LittleEndian.PutUint64(uint64Bytes, uint64(currentPosition))
			_, err = fo.Write(uint64Bytes)
			if err != nil {
				return InvalidFormatError("could not write value position")
			}

			_, err = fo.Seek(currentPosition, 0)

			// write the value
			binary.LittleEndian.PutUint32(uint32Bytes, uint32(len(v)))
			_, err = fo.Write(uint32Bytes)
			if err != nil {
				return InvalidFormatError("could not write entry length for value of: " + k)
			}

			_, err = fo.Write(v)
			if err != nil {
				return InvalidFormatError("could not write entry value for: " + k)
			}
		}

		return nil
	} else {
		return err
	}
}