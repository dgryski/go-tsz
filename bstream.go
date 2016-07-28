package tsz

import (
	"io"
)

// bstream is a stream of bits
type bstream struct {
	// the data stream
	Stream []byte

	// how many bits are valid in current byte
	Count uint8
}

func newBReader(b []byte) *bstream {
	return &bstream{Stream: b, Count: 8}
}

func newBWriter(size int) *bstream {
	return &bstream{Stream: make([]byte, 0, size), Count: 0}
}

func (b *bstream) clone() *bstream {
	d := make([]byte, len(b.Stream))
	copy(d, b.Stream)
	return &bstream{Stream: d, Count: b.Count}
}

func (b *bstream) bytes() []byte {
	return b.Stream
}

type bit bool

const (
	zero bit = false
	one  bit = true
)

func (b *bstream) writeBit(bit bit) {

	if b.Count == 0 {
		b.Stream = append(b.Stream, 0)
		b.Count = 8
	}

	i := len(b.Stream) - 1

	if bit {
		b.Stream[i] |= 1 << (b.Count - 1)
	}

	b.Count--
}

func (b *bstream) writeByte(byt byte) {

	if b.Count == 0 {
		b.Stream = append(b.Stream, 0)
		b.Count = 8
	}

	i := len(b.Stream) - 1

	// fill up b.b with b.Count bits from byt
	b.Stream[i] |= byt >> (8 - b.Count)

	b.Stream = append(b.Stream, 0)
	i++
	b.Stream[i] = byt << b.Count
}

func (b *bstream) writeBits(u uint64, nbits int) {
	u <<= (64 - uint(nbits))
	for nbits >= 8 {
		byt := byte(u >> 56)
		b.writeByte(byt)
		u <<= 8
		nbits -= 8
	}

	for nbits > 0 {
		b.writeBit((u >> 63) == 1)
		u <<= 1
		nbits--
	}
}

func (b *bstream) readBit() (bit, error) {

	if len(b.Stream) == 0 {
		return false, io.EOF
	}

	if b.Count == 0 {
		b.Stream = b.Stream[1:]
		// did we just run out of stuff to read?
		if len(b.Stream) == 0 {
			return false, io.EOF
		}
		b.Count = 8
	}

	b.Count--
	d := b.Stream[0] & 0x80
	b.Stream[0] <<= 1
	return d != 0, nil
}

func (b *bstream) readByte() (byte, error) {

	if len(b.Stream) == 0 {
		return 0, io.EOF
	}

	if b.Count == 0 {
		b.Stream = b.Stream[1:]

		if len(b.Stream) == 0 {
			return 0, io.EOF
		}

		b.Count = 8
	}

	if b.Count == 8 {
		b.Count = 0
		return b.Stream[0], nil
	}

	byt := b.Stream[0]
	b.Stream = b.Stream[1:]

	if len(b.Stream) == 0 {
		return 0, io.EOF
	}

	byt |= b.Stream[0] >> b.Count
	b.Stream[0] <<= (8 - b.Count)

	return byt, nil
}

func (b *bstream) readBits(nbits int) (uint64, error) {

	var u uint64

	for nbits >= 8 {
		byt, err := b.readByte()
		if err != nil {
			return 0, err
		}

		u = (u << 8) | uint64(byt)
		nbits -= 8
	}

	var err error
	for nbits > 0 && err != io.EOF {
		byt, err := b.readBit()
		if err != nil {
			return 0, err
		}
		u <<= 1
		if byt {
			u |= 1
		}
		nbits--
	}

	return u, nil
}
