// Package tsz implement time-series compression
/*

http://www.vldb.org/pvldb/vol8/p1816-teller.pdf

*/
package tsz

import (
	"bytes"
	"encoding/gob"
	"github.com/dgryski/go-bits"
	"github.com/dgryski/go-bitstream"

	"math"
	"sync"
)

// Series is the basic series primitive
// you can concurrently put values, finish the stream, and create iterators
type Series struct {
	sync.Mutex

	// TODO(dgryski): timestamps in the paper are uint64

	t0     uint32
	tDelta uint32
	t      uint32
	val    float64

	leading  uint64
	trailing uint64

	buf bytes.Buffer
	bw  *bitstream.BitWriter

	finished bool
}

// Data structure for serializing a series.
type seriesOnDisk struct {
	T0           uint32
	TDelta       uint32
	T            uint32
	Val          float64
	Leading      uint64
	Trailing     uint64
	B            []byte
	PendingBits  byte
	PendingCount uint8
	Finished     bool
}

// Implimentation GobEncoder interface
// https://golang.org/pkg/encoding/gob/#GobEncoder
func (s *Series) GobEncode() ([]byte, error) {
	s.Lock()
	defer s.Unlock()
	pendingBits, pendingCount := s.bw.Pending()
	sOnDisk := seriesOnDisk{
		T0:           s.t0,
		TDelta:       s.tDelta,
		T:            s.t,
		Val:          s.val,
		Trailing:     s.trailing,
		Leading:      s.leading,
		B:            s.buf.Bytes(),
		PendingBits:  pendingBits,
		PendingCount: pendingCount,
		Finished:     s.finished,
	}
	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	err := enc.Encode(sOnDisk)
	return b.Bytes(), err
}

// Implimentation GobDecode interface
// https://golang.org/pkg/encoding/gob/#GobDecoder
func (s *Series) GobDecode(data []byte) error {
	s.Lock()
	defer s.Unlock()
	r := bytes.NewReader(data)
	dec := gob.NewDecoder(r)
	sOnDisk := &seriesOnDisk{}
	err := dec.Decode(sOnDisk)
	if err != nil {
		return err
	}
	s.t0 = sOnDisk.T0
	s.tDelta = sOnDisk.TDelta
	s.t = sOnDisk.T
	s.val = sOnDisk.Val
	s.leading = sOnDisk.Leading
	s.trailing = sOnDisk.Trailing
	s.buf.Write(sOnDisk.B)
	s.bw = bitstream.NewWriter(&s.buf)
	s.bw.Resume(sOnDisk.PendingBits, sOnDisk.PendingCount)
	return nil
}

func New(t0 uint32) *Series {
	s := Series{
		t0:      t0,
		leading: ^uint64(0),
	}

	s.bw = bitstream.NewWriter(&s.buf)

	// block header
	s.bw.WriteBits(uint64(t0), 32)

	return &s

}

func (s *Series) Bytes() []byte {
	s.Lock()
	defer s.Unlock()
	return s.buf.Bytes()
}

func finish(w *bitstream.BitWriter) {
	// write an end-of-stream record
	w.WriteBits(0x0f, 4)
	w.WriteBits(0xffffffff, 32)
	w.WriteBit(bitstream.Zero)
	w.Flush(bitstream.Zero)
}

func (s *Series) Finish() {
	s.Lock()
	if !s.finished {
		finish(s.bw)
		s.finished = true
	}
	s.Unlock()
}

func (s *Series) Push(t uint32, v float64) {
	s.Lock()
	defer s.Unlock()

	if s.t == 0 {
		// first point
		s.t = t
		s.val = v
		s.tDelta = t - s.t0
		s.bw.WriteBits(uint64(s.tDelta), 14)
		s.bw.WriteBits(math.Float64bits(v), 64)
		return
	}

	tDelta := t - s.t
	dod := int32(tDelta - s.tDelta)

	switch {
	case dod == 0:
		s.bw.WriteBit(bitstream.Zero)
	case -63 <= dod && dod <= 64:
		s.bw.WriteBits(0x02, 2) // '10'
		s.bw.WriteBits(uint64(dod), 7)
	case -255 <= dod && dod <= 256:
		s.bw.WriteBits(0x06, 3) // '110'
		s.bw.WriteBits(uint64(dod), 9)
	case -2047 <= dod && dod <= 2048:
		s.bw.WriteBits(0x0e, 4) // '1110'
		s.bw.WriteBits(uint64(dod), 12)
	default:
		s.bw.WriteBits(0x0f, 4) // '1111'
		s.bw.WriteBits(uint64(dod), 32)
	}

	vDelta := math.Float64bits(v) ^ math.Float64bits(s.val)

	if vDelta == 0 {
		s.bw.WriteBit(bitstream.Zero)
	} else {
		s.bw.WriteBit(bitstream.One)

		leading := bits.Clz(vDelta)
		trailing := bits.Ctz(vDelta)

		// clamp number of leading zeros to avoid overflow when encoding
		if leading >= 32 {
			leading = 31
		}

		// TODO(dgryski): check if it's 'cheaper' to reset the leading/trailing bits instead
		if s.leading != ^uint64(0) && leading >= s.leading && trailing >= s.trailing {
			s.bw.WriteBit(bitstream.Zero)
			s.bw.WriteBits(vDelta>>s.trailing, 64-int(s.leading)-int(s.trailing))
		} else {
			s.leading, s.trailing = leading, trailing

			s.bw.WriteBit(bitstream.One)
			s.bw.WriteBits(leading, 5)

			// Note that if leading == trailing == 0, then sigbits == 64.  But that value doesn't actually fit into the 6 bits we have.
			// Luckily, we never need to encode 0 significant bits, since that would put us in the other case (vdelta == 0).
			// So instead we write out a 0 and adjust it back to 64 on unpacking.
			sigbits := 64 - leading - trailing
			s.bw.WriteBits(sigbits, 6)
			s.bw.WriteBits(vDelta>>trailing, int(sigbits))
		}
	}

	s.tDelta = tDelta
	s.t = t
	s.val = v

}

func (s *Series) Iter() *Iter {
	s.Lock()
	data := s.buf.Bytes()
	newData := make([]byte, len(data), len(data)+1)
	copy(newData, data)
	byt, count := s.bw.Pending()
	s.Unlock()
	buf := bytes.NewBuffer(newData)
	w := bitstream.NewWriter(buf)
	w.Resume(byt, count)
	finish(w)
	iter, _ := NewIterator(buf.Bytes())
	return iter
}

// Iter lets you iterate over a series.  It is not concurrency-safe.
type Iter struct {
	t0 uint32

	tDelta uint32
	t      uint32
	val    float64

	leading  uint64
	trailing uint64

	br *bitstream.BitReader

	b []byte

	finished bool

	err error
}

func NewIterator(b []byte) (*Iter, error) {
	br := bitstream.NewReader(bytes.NewReader(b))

	t0, err := br.ReadBits(32)
	if err != nil {
		return nil, err
	}

	return &Iter{
		t0: uint32(t0),
		br: br,
		b:  b,
	}, nil
}

func (it *Iter) Next() bool {

	if it.err != nil || it.finished {
		return false
	}

	if it.t == 0 {
		// read first t and v
		tDelta, err := it.br.ReadBits(14)
		if err != nil {
			it.err = err
			return false
		}
		it.tDelta = uint32(tDelta)
		it.t = it.t0 + it.tDelta
		v, err := it.br.ReadBits(64)
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
		bit, err := it.br.ReadBit()
		if err != nil {
			it.err = err
			return false
		}
		if bit == bitstream.Zero {
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
		bits, err := it.br.ReadBits(32)
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
		bits, err := it.br.ReadBits(int(sz))
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
	d = 0
	bit, err := it.br.ReadBit()
	if err != nil {
		it.err = err
		return false
	}

	if bit == bitstream.Zero {
		// it.val = it.val
	} else {
		bit, err := it.br.ReadBit()
		if err != nil {
			it.err = err
			return false
		}
		if bit == bitstream.Zero {
			// reuse leading/trailing zero bits
			// it.leading, it.trailing = it.leading, it.trailing
		} else {
			bits, err := it.br.ReadBits(5)
			if err != nil {
				it.err = err
				return false
			}
			it.leading = bits

			bits, err = it.br.ReadBits(6)
			if err != nil {
				it.err = err
				return false
			}
			mbits := bits
			// 0 significant bits here means we overflowed and we actually need 64; see comment in encoder
			if mbits == 0 {
				mbits = 64
			}
			it.trailing = 64 - it.leading - mbits
		}

		mbits := int(64 - it.leading - it.trailing)
		bits, err := it.br.ReadBits(mbits)
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

func (it *Iter) Values() (uint32, float64) {
	return it.t, it.val
}

func (it *Iter) Err() error {
	return it.err
}
