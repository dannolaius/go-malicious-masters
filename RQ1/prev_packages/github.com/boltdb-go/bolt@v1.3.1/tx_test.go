	package bolt_test
	
	import (
		"bytes"
		"errors"
		"fmt"
		"log"
		"os"
		"testing"
	
		"github.com/boltdb-go/bolt"
	)
	
	func TestTx_Commit_ErrTxClosed(t *testing.T) {
		db := MustOpenDB()
		defer db.MustClose()
		tx, err := db.Begin(true)
		if err != nil {
			t.Fatal(err)
		}
	
		if _, err := tx.CreateBucket([]byte("foo")); err != nil {
			t.Fatal(err)
		}
	
		if err := tx.Commit(); err != nil {
			t.Fatal(err)
		}
	
		if err := tx.Commit(); err != bolt.ErrTxClosed {
			t.Fatalf("unexpected error: %s", err)
		}
	}
	
	func TestTx_Rollback_ErrTxClosed(t *testing.T) {
		db := MustOpenDB()
		defer db.MustClose()
	
		tx, err := db.Begin(true)
		if err != nil {
			t.Fatal(err)
		}
	
		if err := tx.Rollback(); err != nil {
			t.Fatal(err)
		}
		if err := tx.Rollback(); err != bolt.ErrTxClosed {
			t.Fatalf("unexpected error: %s", err)
		}
	}
	
	func TestTx_Commit_ErrTxNotWritable(t *testing.T) {
		db := MustOpenDB()
		defer db.MustClose()
		tx, err := db.Begin(false)
		if err != nil {
			t.Fatal(err)
		}
		if err := tx.Commit(); err != bolt.ErrTxNotWritable {
			t.Fatal(err)
		}
	}
	
	func TestTx_Cursor(t *testing.T) {
		db := MustOpenDB()
		defer db.MustClose()
		if err := db.Update(func(tx *bolt.Tx) error {
			if _, err := tx.CreateBucket([]byte("widgets")); err != nil {
				t.Fatal(err)
			}
	
			if _, err := tx.CreateBucket([]byte("woojits")); err != nil {
				t.Fatal(err)
			}
	
			c := tx.Cursor()
			if k, v := c.First(); !bytes.Equal(k, []byte("widgets")) {
				t.Fatalf("unexpected key: %v", k)
			} else if v != nil {
				t.Fatalf("unexpected value: %v", v)
			}
	
			if k, v := c.Next(); !bytes.Equal(k, []byte("woojits")) {
				t.Fatalf("unexpected key: %v", k)
			} else if v != nil {
				t.Fatalf("unexpected value: %v", v)
			}
	
			if k, v := c.Next(); k != nil {
				t.Fatalf("unexpected key: %v", k)
			} else if v != nil {
				t.Fatalf("unexpected value: %v", k)
			}
	
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	
	func TestTx_CreateBucket_ErrTxNotWritable(t *testing.T) {
		db := MustOpenDB()
		defer db.MustClose()
		if err := db.View(func(tx *bolt.Tx) error {
			_, err := tx.CreateBucket([]byte("foo"))
			if err != bolt.ErrTxNotWritable {
				t.Fatalf("unexpected error: %s", err)
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	
	func TestTx_CreateBucket_ErrTxClosed(t *testing.T) {
		db := MustOpenDB()
		defer db.MustClose()
		tx, err := db.Begin(true)
		if err != nil {
			t.Fatal(err)
		}
		if err := tx.Commit(); err != nil {
			t.Fatal(err)
		}
	
		if _, err := tx.CreateBucket([]byte("foo")); err != bolt.ErrTxClosed {
			t.Fatalf("unexpected error: %s", err)
		}
	}
	
	func TestTx_Bucket(t *testing.T) {
		db := MustOpenDB()
		defer db.MustClose()
		if err := db.Update(func(tx *bolt.Tx) error {
			if _, err := tx.CreateBucket([]byte("widgets")); err != nil {
				t.Fatal(err)
			}
			if tx.Bucket([]byte("widgets")) == nil {
				t.Fatal("expected bucket")
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	
	func TestTx_Get_NotFound(t *testing.T) {
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
			if b.Get([]byte("no_such_key")) != nil {
				t.Fatal("expected nil value")
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	
	func TestTx_CreateBucket(t *testing.T) {
		db := MustOpenDB()
		defer db.MustClose()
	
		if err := db.Update(func(tx *bolt.Tx) error {
			b, err := tx.CreateBucket([]byte("widgets"))
			if err != nil {
				t.Fatal(err)
			} else if b == nil {
				t.Fatal("expected bucket")
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	
		if err := db.View(func(tx *bolt.Tx) error {
			if tx.Bucket([]byte("widgets")) == nil {
				t.Fatal("expected bucket")
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	
	func TestTx_CreateBucketIfNotExists(t *testing.T) {
		db := MustOpenDB()
		defer db.MustClose()
		if err := db.Update(func(tx *bolt.Tx) error {
	
			if b, err := tx.CreateBucketIfNotExists([]byte("widgets")); err != nil {
				t.Fatal(err)
			} else if b == nil {
				t.Fatal("expected bucket")
			}
	
			if b, err := tx.CreateBucketIfNotExists([]byte("widgets")); err != nil {
				t.Fatal(err)
			} else if b == nil {
				t.Fatal("expected bucket")
			}
	
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	
		if err := db.View(func(tx *bolt.Tx) error {
			if tx.Bucket([]byte("widgets")) == nil {
				t.Fatal("expected bucket")
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	
	func TestTx_CreateBucketIfNotExists_ErrBucketNameRequired(t *testing.T) {
		db := MustOpenDB()
		defer db.MustClose()
		if err := db.Update(func(tx *bolt.Tx) error {
			if _, err := tx.CreateBucketIfNotExists([]byte{}); err != bolt.ErrBucketNameRequired {
				t.Fatalf("unexpected error: %s", err)
			}
	
			if _, err := tx.CreateBucketIfNotExists(nil); err != bolt.ErrBucketNameRequired {
				t.Fatalf("unexpected error: %s", err)
			}
	
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	
	func TestTx_CreateBucket_ErrBucketExists(t *testing.T) {
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
	
		if err := db.Update(func(tx *bolt.Tx) error {
			if _, err := tx.CreateBucket([]byte("widgets")); err != bolt.ErrBucketExists {
				t.Fatalf("unexpected error: %s", err)
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	
	func TestTx_CreateBucket_ErrBucketNameRequired(t *testing.T) {
		db := MustOpenDB()
		defer db.MustClose()
		if err := db.Update(func(tx *bolt.Tx) error {
			if _, err := tx.CreateBucket(nil); err != bolt.ErrBucketNameRequired {
				t.Fatalf("unexpected error: %s", err)
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	
	func TestTx_DeleteBucket(t *testing.T) {
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
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	
		if err := db.Update(func(tx *bolt.Tx) error {
			if err := tx.DeleteBucket([]byte("widgets")); err != nil {
				t.Fatal(err)
			}
			if tx.Bucket([]byte("widgets")) != nil {
				t.Fatal("unexpected bucket")
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	
		if err := db.Update(func(tx *bolt.Tx) error {
	
			b, err := tx.CreateBucket([]byte("widgets"))
			if err != nil {
				t.Fatal(err)
			}
			if v := b.Get([]byte("foo")); v != nil {
				t.Fatalf("unexpected phantom value: %v", v)
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	
	func TestTx_DeleteBucket_ErrTxClosed(t *testing.T) {
		db := MustOpenDB()
		defer db.MustClose()
		tx, err := db.Begin(true)
		if err != nil {
			t.Fatal(err)
		}
		if err := tx.Commit(); err != nil {
			t.Fatal(err)
		}
		if err := tx.DeleteBucket([]byte("foo")); err != bolt.ErrTxClosed {
			t.Fatalf("unexpected error: %s", err)
		}
	}
	
	func TestTx_DeleteBucket_ReadOnly(t *testing.T) {
		db := MustOpenDB()
		defer db.MustClose()
		if err := db.View(func(tx *bolt.Tx) error {
			if err := tx.DeleteBucket([]byte("foo")); err != bolt.ErrTxNotWritable {
				t.Fatalf("unexpected error: %s", err)
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	
	func TestTx_DeleteBucket_NotFound(t *testing.T) {
		db := MustOpenDB()
		defer db.MustClose()
		if err := db.Update(func(tx *bolt.Tx) error {
			if err := tx.DeleteBucket([]byte("widgets")); err != bolt.ErrBucketNotFound {
				t.Fatalf("unexpected error: %s", err)
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	
	func TestTx_ForEach_NoError(t *testing.T) {
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
	
			if err := tx.ForEach(func(name []byte, b *bolt.Bucket) error {
				return nil
			}); err != nil {
				t.Fatal(err)
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	
	func TestTx_ForEach_WithError(t *testing.T) {
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
	
			marker := errors.New("marker")
			if err := tx.ForEach(func(name []byte, b *bolt.Bucket) error {
				return marker
			}); err != marker {
				t.Fatalf("unexpected error: %s", err)
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	
	func TestTx_OnCommit(t *testing.T) {
		db := MustOpenDB()
		defer db.MustClose()
	
		var x int
		if err := db.Update(func(tx *bolt.Tx) error {
			tx.OnCommit(func() { x += 1 })
			tx.OnCommit(func() { x += 2 })
			if _, err := tx.CreateBucket([]byte("widgets")); err != nil {
				t.Fatal(err)
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		} else if x != 3 {
			t.Fatalf("unexpected x: %d", x)
		}
	}
	
	func TestTx_OnCommit_Rollback(t *testing.T) {
		db := MustOpenDB()
		defer db.MustClose()
	
		var x int
		if err := db.Update(func(tx *bolt.Tx) error {
			tx.OnCommit(func() { x += 1 })
			tx.OnCommit(func() { x += 2 })
			if _, err := tx.CreateBucket([]byte("widgets")); err != nil {
				t.Fatal(err)
			}
			return errors.New("rollback this commit")
		}); err == nil || err.Error() != "rollback this commit" {
			t.Fatalf("unexpected error: %s", err)
		} else if x != 0 {
			t.Fatalf("unexpected x: %d", x)
		}
	}
	
	func TestTx_CopyFile(t *testing.T) {
		db := MustOpenDB()
		defer db.MustClose()
	
		path := tempfile()
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
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	
		if err := db.View(func(tx *bolt.Tx) error {
			return tx.CopyFile(path, 0600)
		}); err != nil {
			t.Fatal(err)
		}
	
		db2, err := bolt.Open(path, 0600, nil)
		if err != nil {
			t.Fatal(err)
		}
	
		if err := db2.View(func(tx *bolt.Tx) error {
			if v := tx.Bucket([]byte("widgets")).Get([]byte("foo")); !bytes.Equal(v, []byte("bar")) {
				t.Fatalf("unexpected value: %v", v)
			}
			if v := tx.Bucket([]byte("widgets")).Get([]byte("baz")); !bytes.Equal(v, []byte("bat")) {
				t.Fatalf("unexpected value: %v", v)
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	
		if err := db2.Close(); err != nil {
			t.Fatal(err)
		}
	}
	
	type failWriterError struct{}
	
	func (failWriterError) Error() string {
		return "error injected for tests"
	}
	
	type failWriter struct {
	
		After int
	}
	
	func (f *failWriter) Write(p []byte) (n int, err error) {
		n = len(p)
		if n > f.After {
			n = f.After
			err = failWriterError{}
		}
		f.After -= n
		return n, err
	}
	
	func TestTx_CopyFile_Error_Meta(t *testing.T) {
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
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	
		if err := db.View(func(tx *bolt.Tx) error {
			return tx.Copy(&failWriter{})
		}); err == nil || err.Error() != "meta 0 copy: error injected for tests" {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	
	func TestTx_CopyFile_Error_Normal(t *testing.T) {
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
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	
		if err := db.View(func(tx *bolt.Tx) error {
			return tx.Copy(&failWriter{3 * db.Info().PageSize})
		}); err == nil || err.Error() != "error injected for tests" {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	
	func ExampleTx_Rollback() {
	
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
	
		if err := db.Update(func(tx *bolt.Tx) error {
			return tx.Bucket([]byte("widgets")).Put([]byte("foo"), []byte("bar"))
		}); err != nil {
			log.Fatal(err)
		}
	
		tx, err := db.Begin(true)
		if err != nil {
			log.Fatal(err)
		}
		b := tx.Bucket([]byte("widgets"))
		if err := b.Put([]byte("foo"), []byte("baz")); err != nil {
			log.Fatal(err)
		}
		if err := tx.Rollback(); err != nil {
			log.Fatal(err)
		}
	
		if err := db.View(func(tx *bolt.Tx) error {
			value := tx.Bucket([]byte("widgets")).Get([]byte("foo"))
			fmt.Printf("The value for 'foo' is still: %s\n", value)
			return nil
		}); err != nil {
			log.Fatal(err)
		}
	
		if err := db.Close(); err != nil {
			log.Fatal(err)
		}
	
	}
	
	func ExampleTx_CopyFile() {
	
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
	
		toFile := tempfile()
		if err := db.View(func(tx *bolt.Tx) error {
			return tx.CopyFile(toFile, 0666)
		}); err != nil {
			log.Fatal(err)
		}
		defer os.Remove(toFile)
	
		db2, err := bolt.Open(toFile, 0666, nil)
		if err != nil {
			log.Fatal(err)
		}
	
		if err := db2.View(func(tx *bolt.Tx) error {
			value := tx.Bucket([]byte("widgets")).Get([]byte("foo"))
			fmt.Printf("The value for 'foo' in the clone is: %s\n", value)
			return nil
		}); err != nil {
			log.Fatal(err)
		}
	
		if err := db.Close(); err != nil {
			log.Fatal(err)
		}
	
		if err := db2.Close(); err != nil {
			log.Fatal(err)
		}
	
	}
	