	package bolt
	
	import (
		"math/rand"
		"reflect"
		"sort"
		"testing"
		"unsafe"
	)
	
	func TestFreelist_free(t *testing.T) {
		f := newFreelist()
		f.free(100, &page{id: 12})
		if !reflect.DeepEqual([]pgid{12}, f.pending[100]) {
			t.Fatalf("exp=%v; got=%v", []pgid{12}, f.pending[100])
		}
	}
	
	func TestFreelist_free_overflow(t *testing.T) {
		f := newFreelist()
		f.free(100, &page{id: 12, overflow: 3})
		if exp := []pgid{12, 13, 14, 15}; !reflect.DeepEqual(exp, f.pending[100]) {
			t.Fatalf("exp=%v; got=%v", exp, f.pending[100])
		}
	}
	
	func TestFreelist_release(t *testing.T) {
		f := newFreelist()
		f.free(100, &page{id: 12, overflow: 1})
		f.free(100, &page{id: 9})
		f.free(102, &page{id: 39})
		f.release(100)
		f.release(101)
		if exp := []pgid{9, 12, 13}; !reflect.DeepEqual(exp, f.ids) {
			t.Fatalf("exp=%v; got=%v", exp, f.ids)
		}
	
		f.release(102)
		if exp := []pgid{9, 12, 13, 39}; !reflect.DeepEqual(exp, f.ids) {
			t.Fatalf("exp=%v; got=%v", exp, f.ids)
		}
	}
	
	func TestFreelist_allocate(t *testing.T) {
		f := &freelist{ids: []pgid{3, 4, 5, 6, 7, 9, 12, 13, 18}}
		if id := int(f.allocate(3)); id != 3 {
			t.Fatalf("exp=3; got=%v", id)
		}
		if id := int(f.allocate(1)); id != 6 {
			t.Fatalf("exp=6; got=%v", id)
		}
		if id := int(f.allocate(3)); id != 0 {
			t.Fatalf("exp=0; got=%v", id)
		}
		if id := int(f.allocate(2)); id != 12 {
			t.Fatalf("exp=12; got=%v", id)
		}
		if id := int(f.allocate(1)); id != 7 {
			t.Fatalf("exp=7; got=%v", id)
		}
		if id := int(f.allocate(0)); id != 0 {
			t.Fatalf("exp=0; got=%v", id)
		}
		if id := int(f.allocate(0)); id != 0 {
			t.Fatalf("exp=0; got=%v", id)
		}
		if exp := []pgid{9, 18}; !reflect.DeepEqual(exp, f.ids) {
			t.Fatalf("exp=%v; got=%v", exp, f.ids)
		}
	
		if id := int(f.allocate(1)); id != 9 {
			t.Fatalf("exp=9; got=%v", id)
		}
		if id := int(f.allocate(1)); id != 18 {
			t.Fatalf("exp=18; got=%v", id)
		}
		if id := int(f.allocate(1)); id != 0 {
			t.Fatalf("exp=0; got=%v", id)
		}
		if exp := []pgid{}; !reflect.DeepEqual(exp, f.ids) {
			t.Fatalf("exp=%v; got=%v", exp, f.ids)
		}
	}
	
	func TestFreelist_read(t *testing.T) {
	
		var buf [4096]byte
		page := (*page)(unsafe.Pointer(&buf[0]))
		page.flags = freelistPageFlag
		page.count = 2
	
		ids := (*[3]pgid)(unsafe.Pointer(&page.ptr))
		ids[0] = 23
		ids[1] = 50
	
		f := newFreelist()
		f.read(page)
	
		if exp := []pgid{23, 50}; !reflect.DeepEqual(exp, f.ids) {
			t.Fatalf("exp=%v; got=%v", exp, f.ids)
		}
	}
	
	func TestFreelist_write(t *testing.T) {
	
		var buf [4096]byte
		f := &freelist{ids: []pgid{12, 39}, pending: make(map[txid][]pgid)}
		f.pending[100] = []pgid{28, 11}
		f.pending[101] = []pgid{3}
		p := (*page)(unsafe.Pointer(&buf[0]))
		if err := f.write(p); err != nil {
			t.Fatal(err)
		}
	
		f2 := newFreelist()
		f2.read(p)
	
		if exp := []pgid{3, 11, 12, 28, 39}; !reflect.DeepEqual(exp, f2.ids) {
			t.Fatalf("exp=%v; got=%v", exp, f2.ids)
		}
	}
	
	func Benchmark_FreelistRelease10K(b *testing.B)    { benchmark_FreelistRelease(b, 10000) }
	func Benchmark_FreelistRelease100K(b *testing.B)   { benchmark_FreelistRelease(b, 100000) }
	func Benchmark_FreelistRelease1000K(b *testing.B)  { benchmark_FreelistRelease(b, 1000000) }
	func Benchmark_FreelistRelease10000K(b *testing.B) { benchmark_FreelistRelease(b, 10000000) }
	
	func benchmark_FreelistRelease(b *testing.B, size int) {
		ids := randomPgids(size)
		pending := randomPgids(len(ids) / 400)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			f := &freelist{ids: ids, pending: map[txid][]pgid{1: pending}}
			f.release(1)
		}
	}
	
	func randomPgids(n int) []pgid {
		rand.Seed(42)
		pgids := make(pgids, n)
		for i := range pgids {
			pgids[i] = pgid(rand.Int63())
		}
		sort.Sort(pgids)
		return pgids
	}
	