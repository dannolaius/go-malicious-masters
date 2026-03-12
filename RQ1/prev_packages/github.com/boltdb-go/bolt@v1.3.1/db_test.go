	package bolt_test
	
	import (
		"bytes"
		"encoding/binary"
		"errors"
		"flag"
		"fmt"
		"hash/fnv"
		"io/ioutil"
		"log"
		"os"
		"path/filepath"
		"regexp"
		"sort"
		"strings"
		"sync"
		"testing"
		"time"
		"unsafe"
	
		"github.com/boltdb-go/bolt"
	)
	
	var statsFlag = flag.Bool("stats", false, "show performance stats")
	
	const version = 2
	
	const magic uint32 = 0xED0CDAED
	
	const pageSize = 4096
	
	const pageHeaderSize = 16
	
	type meta struct {
		magic    uint32
		version  uint32
		_        uint32
		_        uint32
		_        [16]byte
		_        uint64
		pgid     uint64
		_        uint64
		checksum uint64
	}
	
	func TestOpen(t *testing.T) {
		path := tempfile()
		db, err := bolt.Open(path, 0666, nil)
		if err != nil {
			t.Fatal(err)
		} else if db == nil {
			t.Fatal("expected db")
		}
	
		if s := db.Path(); s != path {
			t.Fatalf("unexpected path: %s", s)
		}
	
		if err := db.Close(); err != nil {
			t.Fatal(err)
		}
	}
	
	func TestOpen_ErrPathRequired(t *testing.T) {
		_, err := bolt.Open("", 0666, nil)
		if err == nil {
			t.Fatalf("expected error")
		}
	}
	
	func TestOpen_ErrNotExists(t *testing.T) {
		_, err := bolt.Open(filepath.Join(tempfile(), "bad-path"), 0666, nil)
		if err == nil {
			t.Fatal("expected error")
		}
	}
	
	func TestOpen_ErrInvalid(t *testing.T) {
		path := tempfile()
	
		f, err := os.Create(path)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := fmt.Fprintln(f, "this is not a bolt database"); err != nil {
			t.Fatal(err)
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
		defer os.Remove(path)
	
		if _, err := bolt.Open(path, 0666, nil); err != bolt.ErrInvalid {
			t.Fatalf("unexpected error: %s", err)
		}
	}
	
	func TestOpen_ErrVersionMismatch(t *testing.T) {
		if pageSize != os.Getpagesize() {
			t.Skip("page size mismatch")
		}
	
		db := MustOpenDB()
		path := db.Path()
		defer db.MustClose()
	
		if err := db.DB.Close(); err != nil {
			t.Fatal(err)
		}
	
		buf, err := ioutil.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
	
		meta0 := (*meta)(unsafe.Pointer(&buf[pageHeaderSize]))
		meta0.version++
		meta1 := (*meta)(unsafe.Pointer(&buf[pageSize+pageHeaderSize]))
		meta1.version++
		if err := ioutil.WriteFile(path, buf, 0666); err != nil {
			t.Fatal(err)
		}
	
		if _, err := bolt.Open(path, 0666, nil); err != bolt.ErrVersionMismatch {
			t.Fatalf("unexpected error: %s", err)
		}
	}
	
	func TestOpen_ErrChecksum(t *testing.T) {
		if pageSize != os.Getpagesize() {
			t.Skip("page size mismatch")
		}
	
		db := MustOpenDB()
		path := db.Path()
		defer db.MustClose()
	
		if err := db.DB.Close(); err != nil {
			t.Fatal(err)
		}
	
		buf, err := ioutil.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
	
		meta0 := (*meta)(unsafe.Pointer(&buf[pageHeaderSize]))
		meta0.pgid++
		meta1 := (*meta)(unsafe.Pointer(&buf[pageSize+pageHeaderSize]))
		meta1.pgid++
		if err := ioutil.WriteFile(path, buf, 0666); err != nil {
			t.Fatal(err)
		}
	
		if _, err := bolt.Open(path, 0666, nil); err != bolt.ErrChecksum {
			t.Fatalf("unexpected error: %s", err)
		}
	}
	
	func TestOpen_Size(t *testing.T) {
	
		db := MustOpenDB()
		path := db.Path()
		defer db.MustClose()
	
		pagesize := db.Info().PageSize
	
		if err := db.Update(func(tx *bolt.Tx) error {
			b, _ := tx.CreateBucketIfNotExists([]byte("data"))
			for i := 0; i < 10000; i++ {
				if err := b.Put([]byte(fmt.Sprintf("%04d", i)), make([]byte, 1000)); err != nil {
					t.Fatal(err)
				}
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	
		if err := db.DB.Close(); err != nil {
			t.Fatal(err)
		}
		sz := fileSize(path)
		if sz == 0 {
			t.Fatalf("unexpected new file size: %d", sz)
		}
	
		db0, err := bolt.Open(path, 0666, nil)
		if err != nil {
			t.Fatal(err)
		}
		if err := db0.Update(func(tx *bolt.Tx) error {
			if err := tx.Bucket([]byte("data")).Put([]byte{0}, []byte{0}); err != nil {
				t.Fatal(err)
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
		if err := db0.Close(); err != nil {
			t.Fatal(err)
		}
		newSz := fileSize(path)
		if newSz == 0 {
			t.Fatalf("unexpected new file size: %d", newSz)
		}
	
		if sz < newSz-5*int64(pagesize) {
			t.Fatalf("unexpected file growth: %d => %d", sz, newSz)
		}
	}
	
	func TestOpen_Size_Large(t *testing.T) {
		if testing.Short() {
			t.Skip("short mode")
		}
	
		db := MustOpenDB()
		path := db.Path()
		defer db.MustClose()
	
		pagesize := db.Info().PageSize
	
		var index uint64
		for i := 0; i < 10000; i++ {
			if err := db.Update(func(tx *bolt.Tx) error {
				b, _ := tx.CreateBucketIfNotExists([]byte("data"))
				for j := 0; j < 1000; j++ {
					if err := b.Put(u64tob(index), make([]byte, 50)); err != nil {
						t.Fatal(err)
					}
					index++
				}
				return nil
			}); err != nil {
				t.Fatal(err)
			}
		}
	
		if err := db.DB.Close(); err != nil {
			t.Fatal(err)
		}
		sz := fileSize(path)
		if sz == 0 {
			t.Fatalf("unexpected new file size: %d", sz)
		} else if sz < (1 << 30) {
			t.Fatalf("expected larger initial size: %d", sz)
		}
	
		db0, err := bolt.Open(path, 0666, nil)
		if err != nil {
			t.Fatal(err)
		}
		if err := db0.Update(func(tx *bolt.Tx) error {
			return tx.Bucket([]byte("data")).Put([]byte{0}, []byte{0})
		}); err != nil {
			t.Fatal(err)
		}
		if err := db0.Close(); err != nil {
			t.Fatal(err)
		}
	
		newSz := fileSize(path)
		if newSz == 0 {
			t.Fatalf("unexpected new file size: %d", newSz)
		}
	
		if sz < newSz-5*int64(pagesize) {
			t.Fatalf("unexpected file growth: %d => %d", sz, newSz)
		}
	}
	
	func TestOpen_Check(t *testing.T) {
		path := tempfile()
	
		db, err := bolt.Open(path, 0666, nil)
		if err != nil {
			t.Fatal(err)
		}
		if err := db.View(func(tx *bolt.Tx) error { return <-tx.Check() }); err != nil {
			t.Fatal(err)
		}
		if err := db.Close(); err != nil {
			t.Fatal(err)
		}
	
		db, err = bolt.Open(path, 0666, nil)
		if err != nil {
			t.Fatal(err)
		}
		if err := db.View(func(tx *bolt.Tx) error { return <-tx.Check() }); err != nil {
			t.Fatal(err)
		}
		if err := db.Close(); err != nil {
			t.Fatal(err)
		}
	}
	
	func TestOpen_MetaInitWriteError(t *testing.T) {
		t.Skip("pending")
	}
	
	func TestOpen_FileTooSmall(t *testing.T) {
		path := tempfile()
	
		db, err := bolt.Open(path, 0666, nil)
		if err != nil {
			t.Fatal(err)
		}
		if err := db.Close(); err != nil {
			t.Fatal(err)
		}
	
		if err := os.Truncate(path, int64(os.Getpagesize())); err != nil {
			t.Fatal(err)
		}
	
		db, err = bolt.Open(path, 0666, nil)
		if err == nil || err.Error() != "file size too small" {
			t.Fatalf("unexpected error: %s", err)
		}
	}
	
	func TestDB_Open_InitialMmapSize(t *testing.T) {
		path := tempfile()
		defer os.Remove(path)
	
		initMmapSize := 1 << 31  
		testWriteSize := 1 << 27 
	
		db, err := bolt.Open(path, 0666, &bolt.Options{InitialMmapSize: initMmapSize})
		if err != nil {
			t.Fatal(err)
		}
	
		rtx, err := db.Begin(false)
		if err != nil {
			t.Fatal(err)
		}
	
		wtx, err := db.Begin(true)
		if err != nil {
			t.Fatal(err)
		}
	
		b, err := wtx.CreateBucket([]byte("test"))
		if err != nil {
			t.Fatal(err)
		}
	
		err = b.Put([]byte("foo"), make([]byte, testWriteSize))
		if err != nil {
			t.Fatal(err)
		}
	
		done := make(chan struct{})
	
		go func() {
			if err := wtx.Commit(); err != nil {
				t.Fatal(err)
			}
			done <- struct{}{}
		}()
	
		select {
		case <-time.After(5 * time.Second):
			t.Errorf("unexpected that the reader blocks writer")
		case <-done:
		}
	
		if err := rtx.Rollback(); err != nil {
			t.Fatal(err)
		}
	}
	
	func TestDB_Begin_ErrDatabaseNotOpen(t *testing.T) {
		var db bolt.DB
		if _, err := db.Begin(false); err != bolt.ErrDatabaseNotOpen {
			t.Fatalf("unexpected error: %s", err)
		}
	}
	
	func TestDB_BeginRW(t *testing.T) {
		db := MustOpenDB()
		defer db.MustClose()
	
		tx, err := db.Begin(true)
		if err != nil {
			t.Fatal(err)
		} else if tx == nil {
			t.Fatal("expected tx")
		}
	
		if tx.DB() != db.DB {
			t.Fatal("unexpected tx database")
		} else if !tx.Writable() {
			t.Fatal("expected writable tx")
		}
	
		if err := tx.Commit(); err != nil {
			t.Fatal(err)
		}
	}
	
	func TestDB_BeginRW_Closed(t *testing.T) {
		var db bolt.DB
		if _, err := db.Begin(true); err != bolt.ErrDatabaseNotOpen {
			t.Fatalf("unexpected error: %s", err)
		}
	}
	
	func TestDB_Close_PendingTx_RW(t *testing.T) { testDB_Close_PendingTx(t, true) }
	func TestDB_Close_PendingTx_RO(t *testing.T) { testDB_Close_PendingTx(t, false) }
	
	func testDB_Close_PendingTx(t *testing.T, writable bool) {
		db := MustOpenDB()
		defer db.MustClose()
	
		tx, err := db.Begin(true)
		if err != nil {
			t.Fatal(err)
		}
	
		done := make(chan struct{})
		go func() {
			if err := db.Close(); err != nil {
				t.Fatal(err)
			}
			close(done)
		}()
	
		time.Sleep(100 * time.Millisecond)
		select {
		case <-done:
			t.Fatal("database closed too early")
		default:
		}
	
		if err := tx.Commit(); err != nil {
			t.Fatal(err)
		}
	
		time.Sleep(100 * time.Millisecond)
		select {
		case <-done:
		default:
			t.Fatal("database did not close")
		}
	}
	
	func TestDB_Update(t *testing.T) {
		db := MustOpenDB()
		defer db.MustClose()
		if err := db.Update(func(tx *bolt.Tx) error {
			b, err := tx.CreateBucket([]byte("widgets"))
			if err != nil {
				t.Fatal(err)
			}
			if err := b.Put([]byte("foo"), []byte("bar")); err != nil {
				t.Fatal(err)
			}
			if err := b.Put([]byte("baz"), []byte("bat")); err != nil {
				t.Fatal(err)
			}
			if err := b.Delete([]byte("foo")); err != nil {
				t.Fatal(err)
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
		if err := db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("widgets"))
			if v := b.Get([]byte("foo")); v != nil {
				t.Fatalf("expected nil value, got: %v", v)
			}
			if v := b.Get([]byte("baz")); !bytes.Equal(v, []byte("bat")) {
				t.Fatalf("unexpected value: %v", v)
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	
	func TestDB_Update_Closed(t *testing.T) {
		var db bolt.DB
		if err := db.Update(func(tx *bolt.Tx) error {
			if _, err := tx.CreateBucket([]byte("widgets")); err != nil {
				t.Fatal(err)
			}
			return nil
		}); err != bolt.ErrDatabaseNotOpen {
			t.Fatalf("unexpected error: %s", err)
		}
	}
	
	func TestDB_Update_ManualCommit(t *testing.T) {
		db := MustOpenDB()
		defer db.MustClose()
	
		var panicked bool
		if err := db.Update(func(tx *bolt.Tx) error {
			func() {
				defer func() {
					if r := recover(); r != nil {
						panicked = true
					}
				}()
	
				if err := tx.Commit(); err != nil {
					t.Fatal(err)
				}
			}()
			return nil
		}); err != nil {
			t.Fatal(err)
		} else if !panicked {
			t.Fatal("expected panic")
		}
	}
	
	func TestDB_Update_ManualRollback(t *testing.T) {
		db := MustOpenDB()
		defer db.MustClose()
	
		var panicked bool
		if err := db.Update(func(tx *bolt.Tx) error {
			func() {
				defer func() {
					if r := recover(); r != nil {
						panicked = true
					}
				}()
	
				if err := tx.Rollback(); err != nil {
					t.Fatal(err)
				}
			}()
			return nil
		}); err != nil {
			t.Fatal(err)
		} else if !panicked {
			t.Fatal("expected panic")
		}
	}
	
	func TestDB_View_ManualCommit(t *testing.T) {
		db := MustOpenDB()
		defer db.MustClose()
	
		var panicked bool
		if err := db.View(func(tx *bolt.Tx) error {
			func() {
				defer func() {
					if r := recover(); r != nil {
						panicked = true
					}
				}()
	
				if err := tx.Commit(); err != nil {
					t.Fatal(err)
				}
			}()
			return nil
		}); err != nil {
			t.Fatal(err)
		} else if !panicked {
			t.Fatal("expected panic")
		}
	}
	
	func TestDB_View_ManualRollback(t *testing.T) {
		db := MustOpenDB()
		defer db.MustClose()
	
		var panicked bool
		if err := db.View(func(tx *bolt.Tx) error {
			func() {
				defer func() {
					if r := recover(); r != nil {
						panicked = true
					}
				}()
	
				if err := tx.Rollback(); err != nil {
					t.Fatal(err)
				}
			}()
			return nil
		}); err != nil {
			t.Fatal(err)
		} else if !panicked {
			t.Fatal("expected panic")
		}
	}
	
	func TestDB_Update_Panic(t *testing.T) {
		db := MustOpenDB()
		defer db.MustClose()
	
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Log("recover: update", r)
				}
			}()
	
			if err := db.Update(func(tx *bolt.Tx) error {
				if _, err := tx.CreateBucket([]byte("widgets")); err != nil {
					t.Fatal(err)
				}
				panic("omg")
			}); err != nil {
				t.Fatal(err)
			}
		}()
	
		if err := db.Update(func(tx *bolt.Tx) error {
			if _, err := tx.CreateBucket([]byte("widgets")); err != nil {
				t.Fatal(err)
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	
		if err := db.Update(func(tx *bolt.Tx) error {
			if tx.Bucket([]byte("widgets")) == nil {
				t.Fatal("expected bucket")
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	
	func TestDB_View_Error(t *testing.T) {
		db := MustOpenDB()
		defer db.MustClose()
	
		if err := db.View(func(tx *bolt.Tx) error {
			return errors.New("xxx")
		}); err == nil || err.Error() != "xxx" {
			t.Fatalf("unexpected error: %s", err)
		}
	}
	
	func TestDB_View_Panic(t *testing.T) {
		db := MustOpenDB()
		defer db.MustClose()
	
		if err := db.Update(func(tx *bolt.Tx) error {
			if _, err := tx.CreateBucket([]byte("widgets")); err != nil {
				t.Fatal(err)
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Log("recover: view", r)
				}
			}()
	
			if err := db.View(func(tx *bolt.Tx) error {
				if tx.Bucket([]byte("widgets")) == nil {
					t.Fatal("expected bucket")
				}
				panic("omg")
			}); err != nil {
				t.Fatal(err)
			}
		}()
	
		if err := db.View(func(tx *bolt.Tx) error {
			if tx.Bucket([]byte("widgets")) == nil {
				t.Fatal("expected bucket")
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	
	func TestDB_Stats(t *testing.T) {
		db := MustOpenDB()
		defer db.MustClose()
		if err := db.Update(func(tx *bolt.Tx) error {
			_, err := tx.CreateBucket([]byte("widgets"))
			return err
		}); err != nil {
			t.Fatal(err)
		}
	
		stats := db.Stats()
		if stats.TxStats.PageCount != 2 {
			t.Fatalf("unexpected TxStats.PageCount: %d", stats.TxStats.PageCount)
		} else if stats.FreePageN != 0 {
			t.Fatalf("unexpected FreePageN != 0: %d", stats.FreePageN)
		} else if stats.PendingPageN != 2 {
			t.Fatalf("unexpected PendingPageN != 2: %d", stats.PendingPageN)
		}
	}
	
	func TestDB_Consistency(t *testing.T) {
		db := MustOpenDB()
		defer db.MustClose()
		if err := db.Update(func(tx *bolt.Tx) error {
			_, err := tx.CreateBucket([]byte("widgets"))
			return err
		}); err != nil {
			t.Fatal(err)
		}
	
		for i := 0; i < 10; i++ {
			if err := db.Update(func(tx *bolt.Tx) error {
				if err := tx.Bucket([]byte("widgets")).Put([]byte("foo"), []byte("bar")); err != nil {
					t.Fatal(err)
				}
				return nil
			}); err != nil {
				t.Fatal(err)
			}
		}
	
		if err := db.Update(func(tx *bolt.Tx) error {
			if p, _ := tx.Page(0); p == nil {
				t.Fatal("expected page")
			} else if p.Type != "meta" {
				t.Fatalf("unexpected page type: %s", p.Type)
			}
	
			if p, _ := tx.Page(1); p == nil {
				t.Fatal("expected page")
			} else if p.Type != "meta" {
				t.Fatalf("unexpected page type: %s", p.Type)
			}
	
			if p, _ := tx.Page(2); p == nil {
				t.Fatal("expected page")
			} else if p.Type != "free" {
				t.Fatalf("unexpected page type: %s", p.Type)
			}
	
			if p, _ := tx.Page(3); p == nil {
				t.Fatal("expected page")
			} else if p.Type != "free" {
				t.Fatalf("unexpected page type: %s", p.Type)
			}
	
			if p, _ := tx.Page(4); p == nil {
				t.Fatal("expected page")
			} else if p.Type != "leaf" {
				t.Fatalf("unexpected page type: %s", p.Type)
			}
	
			if p, _ := tx.Page(5); p == nil {
				t.Fatal("expected page")
			} else if p.Type != "freelist" {
				t.Fatalf("unexpected page type: %s", p.Type)
			}
	
			if p, _ := tx.Page(6); p != nil {
				t.Fatal("unexpected page")
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	
	func TestDBStats_Sub(t *testing.T) {
		var a, b bolt.Stats
		a.TxStats.PageCount = 3
		a.FreePageN = 4
		b.TxStats.PageCount = 10
		b.FreePageN = 14
		diff := b.Sub(&a)
		if diff.TxStats.PageCount != 7 {
			t.Fatalf("unexpected TxStats.PageCount: %d", diff.TxStats.PageCount)
		}
	
		if diff.FreePageN != 14 {
			t.Fatalf("unexpected FreePageN: %d", diff.FreePageN)
		}
	}
	
	func TestDB_Batch(t *testing.T) {
		db := MustOpenDB()
		defer db.MustClose()
	
		if err := db.Update(func(tx *bolt.Tx) error {
			if _, err := tx.CreateBucket([]byte("widgets")); err != nil {
				t.Fatal(err)
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	
		n := 2
		ch := make(chan error)
		for i := 0; i < n; i++ {
			go func(i int) {
				ch <- db.Batch(func(tx *bolt.Tx) error {
					return tx.Bucket([]byte("widgets")).Put(u64tob(uint64(i)), []byte{})
				})
			}(i)
		}
	
		for i := 0; i < n; i++ {
			if err := <-ch; err != nil {
				t.Fatal(err)
			}
		}
	
		if err := db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("widgets"))
			for i := 0; i < n; i++ {
				if v := b.Get(u64tob(uint64(i))); v == nil {
					t.Errorf("key not found: %d", i)
				}
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	
	func TestDB_Batch_Panic(t *testing.T) {
		db := MustOpenDB()
		defer db.MustClose()
	
		var sentinel int
		var bork = &sentinel
		var problem interface{}
		var err error
	
		func() {
			defer func() {
				if p := recover(); p != nil {
					problem = p
				}
			}()
			err = db.Batch(func(tx *bolt.Tx) error {
				panic(bork)
			})
		}()
	
		if g, e := err, error(nil); g != e {
			t.Fatalf("wrong error: %v != %v", g, e)
		}
	
		if g, e := problem, bork; g != e {
			t.Fatalf("wrong error: %v != %v", g, e)
		}
	}
	
	func TestDB_BatchFull(t *testing.T) {
		db := MustOpenDB()
		defer db.MustClose()
		if err := db.Update(func(tx *bolt.Tx) error {
			_, err := tx.CreateBucket([]byte("widgets"))
			return err
		}); err != nil {
			t.Fatal(err)
		}
	
		const size = 3
	
		ch := make(chan error, size)
		put := func(i int) {
			ch <- db.Batch(func(tx *bolt.Tx) error {
				return tx.Bucket([]byte("widgets")).Put(u64tob(uint64(i)), []byte{})
			})
		}
	
		db.MaxBatchSize = size
	
		db.MaxBatchDelay = 1 * time.Hour
	
		go put(1)
		go put(2)
	
		time.Sleep(10 * time.Millisecond)
	
		select {
		case <-ch:
			t.Fatalf("batch triggered too early")
		default:
		}
	
		go put(3)
	
		for i := 0; i < size; i++ {
			if err := <-ch; err != nil {
				t.Fatal(err)
			}
		}
	
		if err := db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("widgets"))
			for i := 1; i <= size; i++ {
				if v := b.Get(u64tob(uint64(i))); v == nil {
					t.Errorf("key not found: %d", i)
				}
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	
	func TestDB_BatchTime(t *testing.T) {
		db := MustOpenDB()
		defer db.MustClose()
		if err := db.Update(func(tx *bolt.Tx) error {
			_, err := tx.CreateBucket([]byte("widgets"))
			return err
		}); err != nil {
			t.Fatal(err)
		}
	
		const size = 1
	
		ch := make(chan error, size)
		put := func(i int) {
			ch <- db.Batch(func(tx *bolt.Tx) error {
				return tx.Bucket([]byte("widgets")).Put(u64tob(uint64(i)), []byte{})
			})
		}
	
		db.MaxBatchSize = 1000
		db.MaxBatchDelay = 0
	
		go put(1)
	
		for i := 0; i < size; i++ {
			if err := <-ch; err != nil {
				t.Fatal(err)
			}
		}
	
		if err := db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("widgets"))
			for i := 1; i <= size; i++ {
				if v := b.Get(u64tob(uint64(i))); v == nil {
					t.Errorf("key not found: %d", i)
				}
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	
	func ExampleDB_Update() {
	
		db, err := bolt.Open(tempfile(), 0666, nil)
		if err != nil {
			log.Fatal(err)
		}
		defer os.Remove(db.Path())
	
		if err := db.Update(func(tx *bolt.Tx) error {
			b, err := tx.CreateBucket([]byte("widgets"))
			if err != nil {
				return err
			}
			if err := b.Put([]byte("foo"), []byte("bar")); err != nil {
				return err
			}
			return nil
		}); err != nil {
			log.Fatal(err)
		}
	
		if err := db.View(func(tx *bolt.Tx) error {
			value := tx.Bucket([]byte("widgets")).Get([]byte("foo"))
			fmt.Printf("The value of 'foo' is: %s\n", value)
			return nil
		}); err != nil {
			log.Fatal(err)
		}
	
		if err := db.Close(); err != nil {
			log.Fatal(err)
		}
	
	}
	
	func ExampleDB_View() {
	
		db, err := bolt.Open(tempfile(), 0666, nil)
		if err != nil {
			log.Fatal(err)
		}
		defer os.Remove(db.Path())
	
		if err := db.Update(func(tx *bolt.Tx) error {
			b, err := tx.CreateBucket([]byte("people"))
			if err != nil {
				return err
			}
			if err := b.Put([]byte("john"), []byte("doe")); err != nil {
				return err
			}
			if err := b.Put([]byte("susy"), []byte("que")); err != nil {
				return err
			}
			return nil
		}); err != nil {
			log.Fatal(err)
		}
	
		if err := db.View(func(tx *bolt.Tx) error {
			v := tx.Bucket([]byte("people")).Get([]byte("john"))
			fmt.Printf("John's last name is %s.\n", v)
			return nil
		}); err != nil {
			log.Fatal(err)
		}
	
		if err := db.Close(); err != nil {
			log.Fatal(err)
		}
	
	}
	
	func ExampleDB_Begin_ReadOnly() {
	
		db, err := bolt.Open(tempfile(), 0666, nil)
		if err != nil {
			log.Fatal(err)
		}
		defer os.Remove(db.Path())
	
		if err := db.Update(func(tx *bolt.Tx) error {
			_, err := tx.CreateBucket([]byte("widgets"))
			return err
		}); err != nil {
			log.Fatal(err)
		}
	
		tx, err := db.Begin(true)
		if err != nil {
			log.Fatal(err)
		}
		b := tx.Bucket([]byte("widgets"))
		if err := b.Put([]byte("john"), []byte("blue")); err != nil {
			log.Fatal(err)
		}
		if err := b.Put([]byte("abby"), []byte("red")); err != nil {
			log.Fatal(err)
		}
		if err := b.Put([]byte("zephyr"), []byte("purple")); err != nil {
			log.Fatal(err)
		}
		if err := tx.Commit(); err != nil {
			log.Fatal(err)
		}
	
		tx, err = db.Begin(false)
		if err != nil {
			log.Fatal(err)
		}
		c := tx.Bucket([]byte("widgets")).Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			fmt.Printf("%s likes %s\n", k, v)
		}
	
		if err := tx.Rollback(); err != nil {
			log.Fatal(err)
		}
	
		if err := db.Close(); err != nil {
			log.Fatal(err)
		}
	
	}
	
	func BenchmarkDBBatchAutomatic(b *testing.B) {
		db := MustOpenDB()
		defer db.MustClose()
		if err := db.Update(func(tx *bolt.Tx) error {
			_, err := tx.CreateBucket([]byte("bench"))
			return err
		}); err != nil {
			b.Fatal(err)
		}
	
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			start := make(chan struct{})
			var wg sync.WaitGroup
	
			for round := 0; round < 1000; round++ {
				wg.Add(1)
	
				go func(id uint32) {
					defer wg.Done()
					<-start
	
					h := fnv.New32a()
					buf := make([]byte, 4)
					binary.LittleEndian.PutUint32(buf, id)
					_, _ = h.Write(buf[:])
					k := h.Sum(nil)
					insert := func(tx *bolt.Tx) error {
						b := tx.Bucket([]byte("bench"))
						return b.Put(k, []byte("filler"))
					}
					if err := db.Batch(insert); err != nil {
						b.Error(err)
						return
					}
				}(uint32(round))
			}
			close(start)
			wg.Wait()
		}
	
		b.StopTimer()
		validateBatchBench(b, db)
	}
	
	func BenchmarkDBBatchSingle(b *testing.B) {
		db := MustOpenDB()
		defer db.MustClose()
		if err := db.Update(func(tx *bolt.Tx) error {
			_, err := tx.CreateBucket([]byte("bench"))
			return err
		}); err != nil {
			b.Fatal(err)
		}
	
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			start := make(chan struct{})
			var wg sync.WaitGroup
	
			for round := 0; round < 1000; round++ {
				wg.Add(1)
				go func(id uint32) {
					defer wg.Done()
					<-start
	
					h := fnv.New32a()
					buf := make([]byte, 4)
					binary.LittleEndian.PutUint32(buf, id)
					_, _ = h.Write(buf[:])
					k := h.Sum(nil)
					insert := func(tx *bolt.Tx) error {
						b := tx.Bucket([]byte("bench"))
						return b.Put(k, []byte("filler"))
					}
					if err := db.Update(insert); err != nil {
						b.Error(err)
						return
					}
				}(uint32(round))
			}
			close(start)
			wg.Wait()
		}
	
		b.StopTimer()
		validateBatchBench(b, db)
	}
	
	func BenchmarkDBBatchManual10x100(b *testing.B) {
		db := MustOpenDB()
		defer db.MustClose()
		if err := db.Update(func(tx *bolt.Tx) error {
			_, err := tx.CreateBucket([]byte("bench"))
			return err
		}); err != nil {
			b.Fatal(err)
		}
	
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			start := make(chan struct{})
			var wg sync.WaitGroup
	
			for major := 0; major < 10; major++ {
				wg.Add(1)
				go func(id uint32) {
					defer wg.Done()
					<-start
	
					insert100 := func(tx *bolt.Tx) error {
						h := fnv.New32a()
						buf := make([]byte, 4)
						for minor := uint32(0); minor < 100; minor++ {
							binary.LittleEndian.PutUint32(buf, uint32(id*100+minor))
							h.Reset()
							_, _ = h.Write(buf[:])
							k := h.Sum(nil)
							b := tx.Bucket([]byte("bench"))
							if err := b.Put(k, []byte("filler")); err != nil {
								return err
							}
						}
						return nil
					}
					if err := db.Update(insert100); err != nil {
						b.Fatal(err)
					}
				}(uint32(major))
			}
			close(start)
			wg.Wait()
		}
	
		b.StopTimer()
		validateBatchBench(b, db)
	}
	
	func validateBatchBench(b *testing.B, db *DB) {
		var rollback = errors.New("sentinel error to cause rollback")
		validate := func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte("bench"))
			h := fnv.New32a()
			buf := make([]byte, 4)
			for id := uint32(0); id < 1000; id++ {
				binary.LittleEndian.PutUint32(buf, id)
				h.Reset()
				_, _ = h.Write(buf[:])
				k := h.Sum(nil)
				v := bucket.Get(k)
				if v == nil {
					b.Errorf("not found id=%d key=%x", id, k)
					continue
				}
				if g, e := v, []byte("filler"); !bytes.Equal(g, e) {
					b.Errorf("bad value for id=%d key=%x: %s != %q", id, k, g, e)
				}
				if err := bucket.Delete(k); err != nil {
					return err
				}
			}
	
			c := bucket.Cursor()
			for k, v := c.First(); k != nil; k, v = c.Next() {
				b.Errorf("unexpected key: %x = %q", k, v)
			}
			return rollback
		}
		if err := db.Update(validate); err != nil && err != rollback {
			b.Error(err)
		}
	}
	
	type DB struct {
		*bolt.DB
	}
	
	func MustOpenDB() *DB {
		db, err := bolt.Open(tempfile(), 0666, nil)
		if err != nil {
			panic(err)
		}
		return &DB{db}
	}
	
	func (db *DB) Close() error {
	
		if *statsFlag {
			db.PrintStats()
		}
	
		db.MustCheck()
	
		defer os.Remove(db.Path())
		return db.DB.Close()
	}
	
	func (db *DB) MustClose() {
		if err := db.Close(); err != nil {
			panic(err)
		}
	}
	
	func (db *DB) PrintStats() {
		var stats = db.Stats()
		fmt.Printf("[db] %-20s %-20s %-20s\n",
			fmt.Sprintf("pg(%d/%d)", stats.TxStats.PageCount, stats.TxStats.PageAlloc),
			fmt.Sprintf("cur(%d)", stats.TxStats.CursorCount),
			fmt.Sprintf("node(%d/%d)", stats.TxStats.NodeCount, stats.TxStats.NodeDeref),
		)
		fmt.Printf("     %-20s %-20s %-20s\n",
			fmt.Sprintf("rebal(%d/%v)", stats.TxStats.Rebalance, truncDuration(stats.TxStats.RebalanceTime)),
			fmt.Sprintf("spill(%d/%v)", stats.TxStats.Spill, truncDuration(stats.TxStats.SpillTime)),
			fmt.Sprintf("w(%d/%v)", stats.TxStats.Write, truncDuration(stats.TxStats.WriteTime)),
		)
	}
	
	func (db *DB) MustCheck() {
		if err := db.Update(func(tx *bolt.Tx) error {
	
			var errors []error
			for err := range tx.Check() {
				errors = append(errors, err)
				if len(errors) > 10 {
					break
				}
			}
	
			if len(errors) > 0 {
				var path = tempfile()
				if err := tx.CopyFile(path, 0600); err != nil {
					panic(err)
				}
	
				fmt.Print("\n\n")
				fmt.Printf("consistency check failed (%d errors)\n", len(errors))
				for _, err := range errors {
					fmt.Println(err)
				}
				fmt.Println("")
				fmt.Println("db saved to:")
				fmt.Println(path)
				fmt.Print("\n\n")
				os.Exit(-1)
			}
	
			return nil
		}); err != nil && err != bolt.ErrDatabaseNotOpen {
			panic(err)
		}
	}
	
	func (db *DB) CopyTempFile() {
		path := tempfile()
		if err := db.View(func(tx *bolt.Tx) error {
			return tx.CopyFile(path, 0600)
		}); err != nil {
			panic(err)
		}
		fmt.Println("db copied to: ", path)
	}
	
	func tempfile() string {
		f, err := ioutil.TempFile("", "bolt-")
		if err != nil {
			panic(err)
		}
		if err := f.Close(); err != nil {
			panic(err)
		}
		if err := os.Remove(f.Name()); err != nil {
			panic(err)
		}
		return f.Name()
	}
	
	func mustContainKeys(b *bolt.Bucket, m map[string]string) {
		found := make(map[string]string)
		if err := b.ForEach(func(k, _ []byte) error {
			found[string(k)] = ""
			return nil
		}); err != nil {
			panic(err)
		}
	
		var keys []string
		for k, _ := range found {
			if _, ok := m[string(k)]; !ok {
				keys = append(keys, k)
			}
		}
		if len(keys) > 0 {
			sort.Strings(keys)
			panic(fmt.Sprintf("keys found(%d): %s", len(keys), strings.Join(keys, ",")))
		}
	
		for k, _ := range m {
			if _, ok := found[string(k)]; !ok {
				keys = append(keys, k)
			}
		}
		if len(keys) > 0 {
			sort.Strings(keys)
			panic(fmt.Sprintf("keys not found(%d): %s", len(keys), strings.Join(keys, ",")))
		}
	}
	
	func trunc(b []byte, length int) []byte {
		if length < len(b) {
			return b[:length]
		}
		return b
	}
	
	func truncDuration(d time.Duration) string {
		return regexp.MustCompile(`^(\d+)(\.\d+)`).ReplaceAllString(d.String(), "$1")
	}
	
	func fileSize(path string) int64 {
		fi, err := os.Stat(path)
		if err != nil {
			return 0
		}
		return fi.Size()
	}
	
	func warn(v ...interface{})              { fmt.Fprintln(os.Stderr, v...) }
	func warnf(msg string, v ...interface{}) { fmt.Fprintf(os.Stderr, msg+"\n", v...) }
	
	func u64tob(v uint64) []byte {
		b := make([]byte, 8)
		binary.BigEndian.PutUint64(b, v)
		return b
	}
	
	func btou64(b []byte) uint64 { return binary.BigEndian.Uint64(b) }
	