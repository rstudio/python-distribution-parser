package packages

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"io"
	"log"
	"os"

	"golang.org/x/crypto/blake2b"
)

// Hexdigest struct to hold the hashes
type Hexdigest struct {
	md5    string
	sha2   string
	blake2 string
}

// HashManager struct to manage hash values
type HashManager struct {
	filename string
	md5      *md5Hasher
	sha2     sha256Hasher
	blake2   *blake2Hasher
}

type md5Hasher struct {
	hasher hash.Hash
}
type sha256Hasher struct {
	hasher hash.Hash
}
type blake2Hasher struct {
	hasher hash.Hash
}

func (h *md5Hasher) update(content []byte) {
	if h != nil {
		h.hasher.Write(content)
	}
}
func (h *md5Hasher) hexdigest() string {
	if h == nil {
		return ""
	}
	hash := h.hasher.Sum(nil)
	hashStr := hex.EncodeToString(hash)
	return hashStr
}
func (h *sha256Hasher) update(content []byte) {
	h.hasher.Write(content)
}
func (h *sha256Hasher) hexdigest() string {
	hash := h.hasher.Sum(nil)
	hashStr := hex.EncodeToString(hash)
	return hashStr
}
func (h *blake2Hasher) update(content []byte) {
	if h != nil {
		h.hasher.Write(content)
	}
}
func (h *blake2Hasher) hexdigest() string {
	if h == nil {
		return ""
	}
	hash := h.hasher.Sum(nil)
	hashStr := hex.EncodeToString(hash)
	return hashStr
}

// NewHashManager creates a new HashManager
func NewHashManager(filename string) (*HashManager, error) {
	m := &md5Hasher{hasher: md5.New()}
	s := sha256Hasher{hasher: sha256.New()}
	blake2, err := blake2b.New256(nil)
	if err != nil {
		return nil, err
	}
	b := &blake2Hasher{hasher: blake2}
	return &HashManager{filename: filename, md5: m, sha2: s, blake2: b}, nil
}

func (hm *HashManager) Hash() error {
	file, err := os.Open(hm.filename)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Printf("error closing file: %v", err)
		}
	}(file)

	buffer := make([]byte, 64*1024)
	for {
		bytesRead, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return err
		}
		if bytesRead == 0 {
			break
		}
		content := buffer[:bytesRead]
		hm.md5.update(content)
		hm.sha2.update(content)
		hm.blake2.update(content)
	}

	return nil
}

func (hm *HashManager) HexDigest() *Hexdigest {
	return &Hexdigest{
		md5:    hm.md5.hexdigest(),
		sha2:   hm.sha2.hexdigest(),
		blake2: hm.blake2.hexdigest(),
	}
}
