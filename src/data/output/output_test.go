package output

import (
	"testing"
)

func TestCopyToClipboard(t *testing.T) {

	err := CopyToClipboard([]byte("test"))
	if err != nil {
		t.Errorf("Tried to copy to clipboard but failed with %q", err)
	}

}