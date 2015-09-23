package tsz

import (
	"io"
	"sync"
	"testing"
	"time"
)

func TestExampleEncoding(t *testing.T) {

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

	// extra tests

	// floating point masking/shifting bug
	tunix += 60
	s.Push(tunix, 13)

	tunix += 60
	s.Push(tunix, 24)

	// delta-of-delta sizes
	tunix += 300 // == delta-of-delta of 240
	s.Push(tunix, 24)

	tunix += 900 // == delta-of-delta of 600
	s.Push(tunix, 24)

	tunix += 900 + 2050 // == delta-of-delta of 600
	s.Push(tunix, 24)

	it := s.Iter()

	tunix = uint32(t0.Unix())
	want := []struct {
		t uint32
		v float64
	}{
		{tunix + 62, 12},
		{tunix + 122, 12},
		{tunix + 182, 24},

		{tunix + 242, 13},
		{tunix + 302, 24},

		{tunix + 602, 24},
		{tunix + 1502, 24},
		{tunix + 4452, 24},
	}

	for _, w := range want {
		if !it.Next() {
			t.Fatalf("Next()=false, want true")
		}
		tt, vv := it.Values()
		if w.t != tt || w.v != vv {
			t.Errorf("Values()=(%v,%v), want (%v,%v)\n", tt, vv, w.t, w.v)
		}
	}

	if it.Next() {
		t.Fatalf("Next()=true, want false")
	}

	if err := it.Err(); err != nil {
		t.Errorf("it.Err()=%v, want nil", err)
	}
}

var TwoHoursData = []struct {
	v float64
	t uint32
}{
	// 2h of data
	{761, 1440583200}, {727, 1440583260}, {765, 1440583320}, {706, 1440583380}, {700, 1440583440},
	{679, 1440583500}, {757, 1440583560}, {708, 1440583620}, {739, 1440583680}, {707, 1440583740},
	{699, 1440583800}, {740, 1440583860}, {729, 1440583920}, {766, 1440583980}, {730, 1440584040},
	{715, 1440584100}, {705, 1440584160}, {693, 1440584220}, {765, 1440584280}, {724, 1440584340},
	{799, 1440584400}, {761, 1440584460}, {737, 1440584520}, {766, 1440584580}, {756, 1440584640},
	{719, 1440584700}, {722, 1440584760}, {801, 1440584820}, {747, 1440584880}, {731, 1440584940},
	{742, 1440585000}, {744, 1440585060}, {791, 1440585120}, {750, 1440585180}, {759, 1440585240},
	{809, 1440585300}, {751, 1440585360}, {705, 1440585420}, {770, 1440585480}, {792, 1440585540},
	{727, 1440585600}, {762, 1440585660}, {772, 1440585720}, {721, 1440585780}, {748, 1440585840},
	{753, 1440585900}, {744, 1440585960}, {716, 1440586020}, {776, 1440586080}, {659, 1440586140},
	{789, 1440586200}, {766, 1440586260}, {758, 1440586320}, {690, 1440586380}, {795, 1440586440},
	{770, 1440586500}, {758, 1440586560}, {723, 1440586620}, {767, 1440586680}, {765, 1440586740},
	{693, 1440586800}, {706, 1440586860}, {681, 1440586920}, {727, 1440586980}, {724, 1440587040},
	{780, 1440587100}, {678, 1440587160}, {696, 1440587220}, {758, 1440587280}, {740, 1440587340},
	{735, 1440587400}, {700, 1440587460}, {742, 1440587520}, {747, 1440587580}, {752, 1440587640},
	{734, 1440587700}, {743, 1440587760}, {732, 1440587820}, {746, 1440587880}, {770, 1440587940},
	{780, 1440588000}, {710, 1440588060}, {731, 1440588120}, {712, 1440588180}, {712, 1440588240},
	{741, 1440588300}, {770, 1440588360}, {770, 1440588420}, {754, 1440588480}, {718, 1440588540},
	{670, 1440588600}, {775, 1440588660}, {749, 1440588720}, {795, 1440588780}, {756, 1440588840},
	{741, 1440588900}, {787, 1440588960}, {721, 1440589020}, {745, 1440589080}, {782, 1440589140},
	{765, 1440589200}, {780, 1440589260}, {811, 1440589320}, {790, 1440589380}, {836, 1440589440},
	{743, 1440589500}, {858, 1440589560}, {739, 1440589620}, {762, 1440589680}, {770, 1440589740},
	{752, 1440589800}, {763, 1440589860}, {795, 1440589920}, {792, 1440589980}, {746, 1440590040},
	{786, 1440590100}, {785, 1440590160}, {774, 1440590220}, {786, 1440590280}, {718, 1440590340},
}

func TestRoundtrip(t *testing.T) {

	s := New(TwoHoursData[0].t)
	for _, p := range TwoHoursData {
		s.Push(p.t, p.v)
	}

	it := s.Iter()
	for _, w := range TwoHoursData {
		if !it.Next() {
			t.Fatalf("Next()=false, want true")
		}
		tt, vv := it.Values()
		// t.Logf("it.Values()=(%+v, %+v)\n", time.Unix(int64(tt), 0), vv)
		if w.t != tt || w.v != vv {
			t.Errorf("Values()=(%v,%v), want (%v,%v)\n", tt, vv, w.t, w.v)
		}
	}

	if it.Next() {
		t.Fatalf("Next()=true, want false")
	}

	if err := it.Err(); err != nil {
		t.Errorf("it.Err()=%v, want nil", err)
	}
}

func writer(s *Series, done chan struct{}, sleep time.Duration) {
	for _, p := range TwoHoursData {
		s.Push(p.t, p.v)
		time.Sleep(sleep)
	}
	close(done)
}

func reader(s *Series, done chan struct{}, errors chan string, wg *sync.WaitGroup, t *testing.T) {
	// continuously create iterators to verify the data
	// * when we know the writer is done, we make sure
	// we do 1 last pass verifying all data
	// * until then, we only require data to be correct up until
	// there's no more data returned,and don't require all data to be returned
	lastPass := false
	for {
		select {
		case <-done:
			lastPass = true
		default:
		}
		it := s.Iter()
		iters := 0
		for _, w := range TwoHoursData {
			iters += 1
			if it.Next() {
				tt, vv := it.Values()
				if w.t != tt || w.v != vv {
					errors <- fmt.Sprintf("Values()=(%v,%v), want (%v,%v)\n", tt, vv, w.t, w.v)
				}
			} else {
				if lastPass {
					errors <- fmt.Sprintf("Next()=false, want true")
				}
				// if we're not the last pass, it's ok to not have all data yet, stop looking
				break
			}
		}

		// if not in last pass, Next() may suddenly have a new value now, or not..
		if lastPass && it.Next() {
			errors <- fmt.Sprintf("Next()=true, want false")
		}

		// if not a single value was written yet, we expect an EOF error
		if iters == 0 {
			if err := it.Err(); err != io.EOF {
				errors <- fmt.Sprintf("it.Err()=%v, want EOF", err)
			}
			// but if one or more values exist, we should have a clean end of record
		} else {
			if err := it.Err(); err != nil {
				errors <- fmt.Sprintf("it.Err()=%v, want nil. after %d iters", err, iters)
			}
		}
		if lastPass {
			break
		}
	}
	wg.Done()
}

func TestConcurrentRoundtripImmediateWrites(t *testing.T) {
	testConcurrentRoundtrip(t, time.Duration(0))
}
func TestConcurrentRoundtrip1MsBetweenWrites(t *testing.T) {
	testConcurrentRoundtrip(t, time.Millisecond)
}
func TestConcurrentRoundtrip10MsBetweenWrites(t *testing.T) {
	testConcurrentRoundtrip(t, 10*time.Millisecond)
}

// poorly synchronized. we want to test reading while writing, after finishing writing
// we could introduce explicit synchronization but i don't see a simple way to do so
// in a way that doesn't reduce the possible timing overlaps.
func testConcurrentRoundtrip(t *testing.T, sleep time.Duration) {

	s := New(TwoHoursData[0].t)
	done := make(chan struct{})
	errors := make(chan string)

	wg := &sync.WaitGroup{}

	go func() {
		for err := range errors {
			t.Error(err)
		}
	}()

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go reader(s, done, errors, wg, t)
	}

	go writer(s, done, sleep)

	wg.Wait()
}

func BenchmarkEncode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := New(TwoHoursData[0].t)
		for _, tt := range TwoHoursData {
			s.Push(tt.t, tt.v)
		}
	}
}

func BenchmarkDecode(b *testing.B) {
	s := New(TwoHoursData[0].t)
	for _, tt := range TwoHoursData {
		s.Push(tt.t, tt.v)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		it := s.Iter()
		var j int
		for it.Next() {
			j++
		}
	}
}
