	package bolt
	
	import (
		"syscall"
		"unsafe"
	)
	
	const (
		msAsync      = 1 << iota 
		msSync                   
		msInvalidate             
	)
	
	func msync(db *DB) error {
		_, _, errno := syscall.Syscall(syscall.SYS_MSYNC, uintptr(unsafe.Pointer(db.data)), uintptr(db.datasz), msInvalidate)
		if errno != 0 {
			return errno
		}
		return nil
	}
	
	func fdatasync(db *DB) error {
		if db.data != nil {
			return msync(db)
		}
		return db.file.Sync()
	}
	