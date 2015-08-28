package tsz

import (
	"log"
	"testing"
	"time"
)

func TestEncode(t *testing.T) {

	// Example from the paper
	t0, _ := time.ParseInLocation("Jan _2 2006 15:04:05", "Mar 24 2015 02:00:00", time.Local)
	tunix := uint32(t0.Unix())

	s := New(tunix)

	tunix += 62
	s.Push(tunix, 12)

	tunix += 60
	s.Push(tunix, 12)

	tunix += 60
	s.Push(tunix, 24)

	s.Finish()

	log.Printf("len(s.Bytes())=%+v\n", len(s.Bytes()))

	it := s.Iter()

	for it.Next() {
		t, v := it.Values()
		log.Printf("it.Values()=(%+v, %+v)\n", time.Unix(int64(t), 0), v)
	}

	if err := it.Err(); err != nil {
		log.Printf("err=%+v\n", err)
	}
}
