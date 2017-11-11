package tsz

import (
	"io"
	"testing"
)

func TestNewBWriter(t *testing.T) {
	b := newBWriter(1)
	if b.count != 0 {
		t.Errorf("Unexpected value: %v\n", b.count)
	}
}

func TestReadBitEOF1(t *testing.T) {
	b := newBWriter(1)
	_, err := b.readBit()
	if err != io.EOF {
		t.Errorf("Unexpected value: %v\n", err)
	}
}

func TestReadBitEOF2(t *testing.T) {
	b := newBReader([]byte{1})
	b.count = 0
	_, err := b.readBit()
	if err != io.EOF {
		t.Errorf("Unexpected value: %v\n", err)
	}
}

func TestReadByteEOF1(t *testing.T) {
	b := newBWriter(1)
	_, err := b.readByte()
	if err != io.EOF {
		t.Errorf("Unexpected value: %v\n", err)
	}
}

func TestReadByteEOF2(t *testing.T) {
	b := newBReader([]byte{1})
	b.count = 0
	_, err := b.readByte()
	if err != io.EOF {
		t.Errorf("Unexpected value: %v\n", err)
	}
}

func TestReadByteEOF3(t *testing.T) {
	b := newBReader([]byte{1})
	b.count = 16
	_, err := b.readByte()
	if err != io.EOF {
		t.Errorf("Unexpected value: %v\n", err)
	}
}

func TestReadBitsEOF(t *testing.T) {
	b := newBReader([]byte{1})
	_, err := b.readBits(9)
	if err != io.EOF {
		t.Errorf("Unexpected value: %v\n", err)
	}
}

func TestUnmarshalBinaryErr(t *testing.T) {
	b := &bstream{}
	err := b.UnmarshalBinary([]byte{})
	if err == nil {
		t.Errorf("An error was expected\n")
	}
}
