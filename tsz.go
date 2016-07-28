// Package tsz implement time-series compression
/*

http://www.vldb.org/pvldb/vol8/p1816-teller.pdf

*/
package tsz

import (
	"encoding/gob"
	"log"
	"math"
	"os"
	"sync"

	"github.com/dgryski/go-bits"
)

// Series is the basic series primitive
// you can concurrently put values, finish the stream, and create iterators
type Series struct {
	mutex sync.Mutex

	// TODO(dgryski): timestamps in the paper are uint64
	T0  uint32
	T   uint32
	Val float64

	BW       bstream
	Leading  uint8
	Trailing uint8
	Finished bool

	TDelta uint32
}

// Lock the series
func (s *Series) Lock() {
	s.mutex.Lock()
}

// Unlock the series
func (s *Series) Unlock() {
	s.mutex.Unlock()
}

// New series with t0 as start
func New(t0 uint32) *Series {
	s := Series{
		T0:      t0,
		Leading: ^uint8(0),
	}

	// block header
	s.BW.writeBits(uint64(t0), 32)

	return &s

}

// Bytes of the series
func (s *Series) Bytes() []byte {
	s.Lock()
	defer s.Unlock()
	return s.BW.bytes()
}

func finish(w *bstream) {
	// write an end-of-stream record
	w.writeBits(0x0f, 4)
	w.writeBits(0xffffffff, 32)
	w.writeBit(zero)
}

// Finish the series
func (s *Series) Finish() {
	s.Lock()
	if !s.Finished {
		finish(&s.BW)
		s.Finished = true
	}
	s.Unlock()
}

// Push timestamp and value to the series
func (s *Series) Push(t uint32, v float64) {
	s.Lock()
	defer s.Unlock()

	if s.T == 0 {
		// first point
		s.T = t
		s.Val = v
		s.TDelta = t - s.T0
		s.BW.writeBits(uint64(s.TDelta), 14)
		s.BW.writeBits(math.Float64bits(v), 64)
		return
	}

	tDelta := t - s.T
	dod := int32(tDelta - s.TDelta)

	switch {
	case dod == 0:
		s.BW.writeBit(zero)
	case -63 <= dod && dod <= 64:
		s.BW.writeBits(0x02, 2) // '10'
		s.BW.writeBits(uint64(dod), 7)
	case -255 <= dod && dod <= 256:
		s.BW.writeBits(0x06, 3) // '110'
		s.BW.writeBits(uint64(dod), 9)
	case -2047 <= dod && dod <= 2048:
		s.BW.writeBits(0x0e, 4) // '1110'
		s.BW.writeBits(uint64(dod), 12)
	default:
		s.BW.writeBits(0x0f, 4) // '1111'
		s.BW.writeBits(uint64(dod), 32)
	}

	vDelta := math.Float64bits(v) ^ math.Float64bits(s.Val)

	if vDelta == 0 {
		s.BW.writeBit(zero)
	} else {
		s.BW.writeBit(one)

		leading := uint8(bits.Clz(vDelta))
		trailing := uint8(bits.Ctz(vDelta))

		// clamp number of leading zeros to avoid overflow when encoding
		if leading >= 32 {
			leading = 31
		}

		// TODO(dgryski): check if it's 'cheaper' to reset the leading/trailing bits instead
		if s.Leading != ^uint8(0) && leading >= s.Leading && trailing >= s.Trailing {
			s.BW.writeBit(zero)
			s.BW.writeBits(vDelta>>s.Trailing, 64-int(s.Leading)-int(s.Trailing))
		} else {
			s.Leading, s.Trailing = leading, trailing

			s.BW.writeBit(one)
			s.BW.writeBits(uint64(leading), 5)

			// Note that if leading == trailing == 0, then sigbits == 64.  But that value doesn't actually fit into the 6 bits we have.
			// Luckily, we never need to encode 0 significant bits, since that would put us in the other case (vdelta == 0).
			// So instead we write out a 0 and adjust it back to 64 on unpacking.
			sigbits := 64 - leading - trailing
			s.BW.writeBits(uint64(sigbits), 6)
			s.BW.writeBits(vDelta>>trailing, int(sigbits))
		}
	}

	s.TDelta = tDelta
	s.T = t
	s.Val = v

}

// Iter of key/val of the series
func (s *Series) Iter() *Iter {
	s.Lock()
	w := s.BW.clone()
	s.Unlock()

	finish(w)
	iter, _ := bstreamIterator(w)
	return iter
}

// Iter lets you iterate over a series.  It is not concurrency-safe.
type Iter struct {
	T0 uint32

	t   uint32
	val float64

	br       bstream
	leading  uint8
	trailing uint8

	finished bool

	tDelta uint32
	err    error
}

func bstreamIterator(br *bstream) (*Iter, error) {

	br.Count = 8

	t0, err := br.readBits(32)
	if err != nil {
		return nil, err
	}

	return &Iter{
		T0: uint32(t0),
		br: *br,
	}, nil
}

// NewIterator on the series
func NewIterator(b []byte) (*Iter, error) {
	return bstreamIterator(newBReader(b))
}

// Next item from the iterator
func (it *Iter) Next() bool {

	if it.err != nil || it.finished {
		return false
	}

	if it.t == 0 {
		// read first t and v
		tDelta, err := it.br.readBits(14)
		if err != nil {
			it.err = err
			return false
		}
		it.tDelta = uint32(tDelta)
		it.t = it.T0 + it.tDelta
		v, err := it.br.readBits(64)
		if err != nil {
			it.err = err
			return false
		}

		it.val = math.Float64frombits(v)

		return true
	}

	// read delta-of-delta
	var d byte
	for i := 0; i < 4; i++ {
		d <<= 1
		bit, err := it.br.readBit()
		if err != nil {
			it.err = err
			return false
		}
		if bit == zero {
			break
		}
		d |= 1
	}

	var dod int32
	var sz uint
	switch d {
	case 0x00:
		// dod == 0
	case 0x02:
		sz = 7
	case 0x06:
		sz = 9
	case 0x0e:
		sz = 12
	case 0x0f:
		bits, err := it.br.readBits(32)
		if err != nil {
			it.err = err
			return false
		}

		// end of stream
		if bits == 0xffffffff {
			it.finished = true
			return false
		}

		dod = int32(bits)
	}

	if sz != 0 {
		bits, err := it.br.readBits(int(sz))
		if err != nil {
			it.err = err
			return false
		}
		if bits > (1 << (sz - 1)) {
			// or something
			bits = bits - (1 << sz)
		}
		dod = int32(bits)
	}

	tDelta := it.tDelta + uint32(dod)

	it.tDelta = tDelta
	it.t = it.t + it.tDelta

	// read compressed value
	bit, err := it.br.readBit()
	if err != nil {
		it.err = err
		return false
	}

	if bit == zero {
		// it.val = it.val
	} else {
		bit, err := it.br.readBit()
		if err != nil {
			it.err = err
			return false
		}
		if bit == zero {
			// reuse leading/trailing zero bits
			// it.leading, it.trailing = it.leading, it.trailing
		} else {
			bits, err := it.br.readBits(5)
			if err != nil {
				it.err = err
				return false
			}
			it.leading = uint8(bits)

			bits, err = it.br.readBits(6)
			if err != nil {
				it.err = err
				return false
			}
			mbits := uint8(bits)
			// 0 significant bits here means we overflowed and we actually need 64; see comment in encoder
			if mbits == 0 {
				mbits = 64
			}
			it.trailing = 64 - it.leading - mbits
		}

		mbits := int(64 - it.leading - it.trailing)
		bits, err := it.br.readBits(mbits)
		if err != nil {
			it.err = err
			return false
		}
		vbits := math.Float64bits(it.val)
		vbits ^= (bits << it.trailing)
		it.val = math.Float64frombits(vbits)
	}

	return true
}

// Values from the iterator
func (it *Iter) Values() (uint32, float64) {
	return it.t, it.val
}

// Err from the iterator
func (it *Iter) Err() error {
	return it.err
}

// Save the series to a file
func (s *Series) Save(path string) error {
	file, err := os.Create(path)
	defer file.Close()
	if err != nil {
		return err
	}
	encoder := gob.NewEncoder(file)
	err = encoder.Encode(s)
	if err != nil {
		return err
	}
	return nil
}

// Load a series from a file
func (s *Series) Load(path string) error {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return err
	}
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&s)
	if err != nil {
		log.Println("failed decoding file")
		return err
	}
	return nil
}
