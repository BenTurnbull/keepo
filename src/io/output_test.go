package io

import (
	"testing"
)

func TestCopyToClipboard(t *testing.T) {

	CopyToClipboard([]byte("test"))

}