package crypto

import (
	"testing"
	"fmt"
)

func TestCrypto(t *testing.T) {
	cases := []struct {
		in []byte
		want string
	}{
		{[]byte("Hello, world"), "Hello, world"},
		{[]byte(""), ""},
	}

	key := []byte("AES256Key-32Characters1234567890")
	for _, c := range cases {
		value, err := Encrypt(key, c.in)
		if err != nil {
			t.Errorf("Tried to encrypt but failed with %q", err)
		}

		fmt.Printf("Received encrypted value %q \n", value)

		got, err := Decrypt(key, value)
		if err != nil {
			t.Errorf("Tried to decrypt but failed with %q", err)
		}

		fmt.Printf("Received decrypted value %q \n", got)

		if string(got) != c.want {
			t.Errorf("Expected %q, received %q", c.want, got)
		}
	}
}