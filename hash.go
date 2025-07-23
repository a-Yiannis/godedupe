package main

import (
	"io"
	"os"
	"sync"

	"github.com/cespare/xxhash/v2"
)

const sampleSize = 16 << 10 // 16 KB for partial-hash sampling

var bufPool = sync.Pool{
	New: func() interface{} {
		// allocate a fixed-size array, return its address
		var b [sampleSize]byte
		return &b
	},
}

// partialHash reads 3 small chunks from the file (start, middle, end).
// Buffers are reused from bufPool to avoid repeated allocations.
func partialHash(path string) (uint64, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return 0, err
	}
	size := info.Size()

	// get a buffer from pool
	// new: pull a *[sampleSize]byte, slice it, and return the pointer
	bufPtr := bufPool.Get().(*[sampleSize]byte)
	defer bufPool.Put(bufPtr)
	buf := bufPtr[:] // []byte view of entire array

	h := xxhash.New()

	// 1) start
	n, _ := f.Read(buf)
	h.Write(buf[:n])

	// 2) middle
	if size > int64(len(buf)) {
		mid := size/2 - int64(len(buf))/2
		if _, err := f.Seek(mid, io.SeekStart); err == nil {
			n, _ = f.Read(buf)
			h.Write(buf[:n])
		}
	}

	// 3) end
	if size > int64(len(buf)) {
		end := size - int64(len(buf))
		if _, err := f.Seek(end, io.SeekStart); err == nil {
			n, _ = f.Read(buf)
			h.Write(buf[:n])
		}
	}

	return h.Sum64(), nil
}

// fullHash computes the full 64-bit xxHash of a file.
// This is used to confirm duplicates after a partial hash collision.
func fullHash(path string) (uint64, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	h := xxhash.New()
	if _, err := io.Copy(h, f); err != nil {
		return 0, err
	}
	return h.Sum64(), nil
}
