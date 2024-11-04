package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
)

func generateID() string {
	buf := make([]byte, 32)
	io.ReadFull(rand.Reader, buf)

	return hex.EncodeToString(buf)
}
func hashKey(key string) string {
	hash := md5.Sum([]byte(key))
	return hex.EncodeToString(hash[:])
}
func newEncryptionkey() []byte {
	keyBuf := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, keyBuf); err != nil {
		fmt.Println(err)
	}
	return keyBuf
}

func processStream(key []byte, src io.Reader, dst io.Writer, encrypt bool) (int, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return 0, err
	}

	iv := make([]byte, block.BlockSize())

	// For encryption, generate a new IV and write it to the destination.
	// For decryption, read the IV from the source.
	if encrypt {
		if _, err := io.ReadFull(rand.Reader, iv); err != nil {
			return 0, err
		}
		if _, err := dst.Write(iv); err != nil {
			return 0, err
		}
	} else {
		if _, err := src.Read(iv); err != nil {
			return 0, err
		}
	}

	stream := cipher.NewCTR(block, iv)
	var (
		buf = make([]byte, 32*1024)
		nw  = block.BlockSize()
	)

	for {
		n, err := src.Read(buf)
		if n > 0 {
			stream.XORKeyStream(buf[:n], buf[:n])
			nn, writeErr := dst.Write(buf[:n])
			if writeErr != nil {
				return 0, writeErr
			}
			nw += nn
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}
	}

	return nw, nil
}

func copyEncrypt(key []byte, src io.Reader, dst io.Writer) (int, error) {
	return processStream(key, src, dst, true)
}

func copyDecrypt(key []byte, src io.Reader, dst io.Writer) (int, error) {
	return processStream(key, src, dst, false)
}
