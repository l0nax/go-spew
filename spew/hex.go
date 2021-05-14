package spew

import (
	"errors"
	"io"
	"strings"
)

// Implements the encoding/hex package but with colors.
// NOTE: The code has been COPIED from https://golang.org/src/encoding/hex/hex.go
//		 => Copyright (c) 2009 The Go Authors. All rights reserved.

func HexDump(data []byte) string {
	if len(data) == 0 {
		return ""
	}

	var buf strings.Builder
	// Dumper will write 79 bytes per complete 16 byte chunk, and at least
	// 64 bytes for whatever remains. Round the allocation up, since only a
	// maximum of 15 bytes will be wasted.
	buf.Grow((1 + ((len(data) - 1) / 16)) * 79)

	dumper := Dumper(&buf)
	dumper.Write(data)
	dumper.Close()
	return buf.String()
}

const hextable = "0123456789abcdef"

// HexEncode encodes src into EncodedLen(len(src))
// bytes of dst. As a convenience, it returns the number
// of bytes written to dst, but this value is always EncodedLen(len(src)).
// Encode implements hexadecimal encoding.
func HexEncode(dst, src []byte) int {
	j := 0

	for _, v := range src {
		dst[j] = hextable[v>>4]
		dst[j+1] = hextable[v&0x0f]
		j += 2
	}

	return len(src) * 2
}

// Dumper returns a WriteCloser that writes a hex dump of all written data to
// w. The format of the dump matches the output of `hexdump -C` on the command
// line.
func Dumper(w io.Writer) io.WriteCloser {
	return &dumper{w: w}
}

type dumper struct {
	w          io.Writer
	rightChars [18]byte
	buf        [14]byte
	used       int  // number of bytes in the current line
	n          uint // number of bytes, total
	closed     bool
}

func toChar(b byte) byte {
	if b < 32 || b > 126 {
		return '.'
	}

	return b
}

func (h *dumper) Write(data []byte) (n int, err error) {
	if h.closed {
		return 0, errors.New("encoding/hex: dumper closed")
	}

	// Output lines look like:
	// 00000010  2e 2f 30 31 32 33 34 35  36 37 38 39 3a 3b 3c 3d  |./0123456789:;<=|
	// ^ offset                          ^ extra space              ^ ASCII of line.
	for i := range data {
		if h.used == 0 {
			// At the beginning of a line we print the current
			// offset in hex.
			h.buf[0] = byte(h.n >> 24)
			h.buf[1] = byte(h.n >> 16)
			h.buf[2] = byte(h.n >> 8)
			h.buf[3] = byte(h.n)
			HexEncode(h.buf[4:], h.buf[:4])
			h.buf[12] = ' '
			h.buf[13] = ' '

			_, err = h.w.Write(h.buf[4:])
			if err != nil {
				return
			}
		}

		HexEncode(h.buf[:], data[i:i+1])
		h.buf[2] = ' '
		l := 3

		if h.used == 7 {
			// There's an additional space after the 8th byte.
			h.buf[3] = ' '
			l = 4
		} else if h.used == 15 {
			// At the end of the line there's an extra space and
			// the bar for the right column.
			h.buf[3] = ' '
			h.buf[4] = '|'
			l = 5
		}

		_, err = h.w.Write(h.buf[:l])
		if err != nil {
			return
		}

		n++
		h.rightChars[h.used] = toChar(data[i])
		h.used++
		h.n++
		if h.used == 16 {
			h.rightChars[16] = '|'
			h.rightChars[17] = '\n'
			_, err = h.w.Write(h.rightChars[:])
			if err != nil {
				return
			}
			h.used = 0
		}
	}

	return
}

func (h *dumper) Close() (err error) {
	// See the comments in Write() for the details of this format.
	if h.closed {
		return
	}

	h.closed = true
	if h.used == 0 {
		return
	}

	h.buf[0] = ' '
	h.buf[1] = ' '
	h.buf[2] = ' '
	h.buf[3] = ' '
	h.buf[4] = '|'
	nBytes := h.used

	for h.used < 16 {
		l := 3

		if h.used == 7 {
			l = 4
		} else if h.used == 15 {
			l = 5
		}

		_, err = h.w.Write(h.buf[:l])
		if err != nil {
			return
		}

		h.used++
	}

	h.rightChars[nBytes] = '|'
	h.rightChars[nBytes+1] = '\n'
	_, err = h.w.Write(h.rightChars[:nBytes+2])

	return
}
