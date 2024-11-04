package main

import (
	"bytes"
	"testing"
)

func TestCopyEncryptDecrypt(t *testing.T) {
	payload := "is the test working"
	src := bytes.NewBuffer([]byte(payload))
	dst := new(bytes.Buffer)
	key := newEncryptionkey()

	_, err := copyEncrypt(key, src, dst)

	if err != nil {
		t.Error(err)
	}

	// fmt.Println(dst.String())

	out := new(bytes.Buffer)

	nw, err := copyDecrypt(key, dst, out)
	if err != nil {
		t.Error(err)
	}

	if nw != 16+len(payload) {
		t.Fail()

	}

	if out.String() != payload {
		t.Error("Decryption failed")
	}

	// fmt.Println(out.String())

}
