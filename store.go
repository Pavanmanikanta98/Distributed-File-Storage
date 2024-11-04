package main

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

const defaultRootFoldername = "glnetwork"

func CASPathTransformFunc(key string) PathKey {
	//[20]byte => []byte -> [:](convert into slice)
	hash := sha1.Sum([]byte(key))
	hashStr := hex.EncodeToString(hash[:])

	blockSize := 5
	sliceLen := len(hashStr) / blockSize
	paths := make([]string, sliceLen)
	for i := 0; i < sliceLen; i++ {
		from, to := i*blockSize, (i*blockSize)+blockSize
		paths[i] = hashStr[from:to]
	}

	return PathKey{
		Pathname: strings.Join(paths, "/"),
		Filename: hashStr,
	}
	// return strings.Join(paths, "/")

}

type PathTransformFunc func(string) PathKey

type PathKey struct {
	Pathname string
	Filename string
}

func (p PathKey) FirstPathname() string {
	paths := strings.Split(p.Pathname, "/")
	if len(paths) == 0 {
		return ""
	}
	return paths[0]
}

func (p PathKey) FullPath() string {
	return fmt.Sprintf("%s/%s", p.Pathname, p.Filename)
}

type StoreOpts struct {
	//Root is the folder name of the root directory, containing the folders/files of the system.
	Root              string
	PathTransformFunc PathTransformFunc
	//ID of the owner of the storage, which will be used to store all files at that locations
	//so we can sync all the files if needed.
	ID string
}

var DefaultPathTransformFunc = func(key string) PathKey {

	return PathKey{
		Pathname: key,
		Filename: key,
	}

}

type store struct {
	StoreOpts
}

func NewStore(opts StoreOpts) *store {

	if opts.PathTransformFunc == nil {
		opts.PathTransformFunc = DefaultPathTransformFunc
	}
	if len(opts.Root) == 0 {
		opts.Root = defaultRootFoldername
	}
	if len(opts.ID) == 0 {
		opts.ID = generateID()
	}
	return &store{
		StoreOpts: opts,
	}

}

func (s *store) Has(key string) bool {

	pathKey := s.PathTransformFunc(key)
	fullPathWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, s.ID, pathKey.FullPath())
	_, err := os.Stat(fullPathWithRoot)

	return !errors.Is(err, os.ErrNotExist)
}

func (s *store) clear() error {

	return os.RemoveAll(s.Root)
}

func (s *store) Delete(key string) error {

	pathkey := s.PathTransformFunc(key)

	defer func() {
		log.Printf("Deleting %s from disk", pathkey.Filename)
	}()

	firstPAthnameWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, s.ID, pathkey.FirstPathname())

	return os.RemoveAll(firstPAthnameWithRoot)
}

func (s *store) Write(key string, data io.Reader) (int64, error) {
	// f, err := s.writeStream(key, data)

	// if err!= nil {
	//     return err
	// }

	// defer f.Close()

	// _, err = io.Copy(f, data)

	// return err
	return s.writeStream(key, data)
}

func (s *store) WriteDecrypt(enckey []byte, key string, r io.Reader) (int64, error) {
	f, err := s.openFileForWriting(key)

	if err != nil {
		return 0, err
	}

	n, err := copyDecrypt(enckey, r, f)

	return int64(n), err
}

func (s *store) openFileForWriting(key string) (*os.File, error) {
	pathkey := s.PathTransformFunc(key)
	pathWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, s.ID, pathkey.Pathname)

	if err := os.MkdirAll(pathWithRoot, os.ModePerm); err != nil {
		return nil, err
	}
	// fullPath := pathkey.FullPath()
	fullPathWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, s.ID, pathkey.FullPath())
	return os.Create(fullPathWithRoot)
}
func (s *store) writeStream(key string, r io.Reader) (int64, error) {

	f, err := s.openFileForWriting(key)
	if err != nil {
		return 0, err
	}
	return io.Copy(f, r)

}

// FIXME: Done
func (s *store) Read(key string) (int64, io.Reader, error) {

	return s.readStream(key)

}
func (s *store) readStream(key string) (int64, io.ReadCloser, error) {

	pathKey := s.PathTransformFunc(key)
	fullPathWithRoot := fmt.Sprintf("%s/%s/%s", s.Root, s.ID, pathKey.FullPath())

	file, err := os.Open(fullPathWithRoot)

	if err != nil {
		return 0, nil, err
	}

	fi, err := file.Stat()

	if err != nil {
		return 0, nil, err
	}
	return fi.Size(), file, nil
}
