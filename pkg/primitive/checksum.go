package primitive

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"
	"hash/crc32"
	"hash/fnv"
	"io"

	"github.com/zeebo/xxh3"
	"golang.org/x/crypto/blake2b"
)

// A ChecksumAlg is the name of a hash function.
type ChecksumAlg string

// List of supported hash algs.
const (
	ChecksumCRC32      ChecksumAlg = "crc32"
	ChecksumFNV1A64    ChecksumAlg = "fnv1a64"
	ChecksumXXH3       ChecksumAlg = "xxh3" // xxHash 64bit
	ChecksumMD5        ChecksumAlg = "md5"
	ChecksumSHA1       ChecksumAlg = "sha1"
	ChecksumSHA256     ChecksumAlg = "sha256"
	ChecksumSHA512     ChecksumAlg = "sha512"
	ChecksumBLAKE2B    ChecksumAlg = "blake2b"
	ChecksumBLAKE2B512 ChecksumAlg = "blake2b512"
)

// Checksum performs the sum of the given r.
func Checksum(r io.Reader, algs ...ChecksumAlg) (map[ChecksumAlg]hash.Hash, error) {
	var err error
	hashes := []io.Writer{}
	mhashes := map[ChecksumAlg]hash.Hash{}

	for _, alg := range algs {
		var h hash.Hash
		switch alg {
		case ChecksumCRC32:
			h = crc32.New(crc32.IEEETable)
		case ChecksumFNV1A64:
			h = fnv.New64a()
		case ChecksumXXH3:
			h = xxh3.New()
		case ChecksumMD5:
			h = md5.New()
		case ChecksumSHA1:
			h = sha1.New()
		case ChecksumSHA256:
			h = sha256.New()
		case ChecksumSHA512:
			h = sha512.New()
		case ChecksumBLAKE2B:
			h, err = blake2b.New256(nil)
			if err != nil {
				return nil, fmt.Errorf("blake2b: %w", err)
			}
		case ChecksumBLAKE2B512:
			h, err = blake2b.New512(nil)
			if err != nil {
				return nil, fmt.Errorf("blake2b: %w", err)
			}
		default:
			return nil, fmt.Errorf("Unsuported algorithm: %s", alg)
		}
		hashes = append(hashes, h)
		mhashes[alg] = h
	}

	w := io.MultiWriter(hashes...)
	_, err = io.Copy(w, r)
	if err != nil {
		return nil, fmt.Errorf("checksum: %w", err)
	}
	return mhashes, nil
}
