package spew

import (
	"errors"
	"io"
	"strings"
)

// Implements the encoding/hex package but with colors.
// NOTE: The code has been COPIED from https://golang.org/src/encoding/hex/hex.go
//		 => Copyright (c) 2009 The Go Authors. All rights reserved.

func HexDump(data []byte, colorize bool) string {
	if len(data) == 0 {
		return ""
	}

	var buf strings.Builder
	// Dumper will write 79 bytes per complete 16 byte chunk, and at least
	// 64 bytes for whatever remains. Round the allocation up, since only a
	// maximum of 15 bytes will be wasted.
	// NOTE: The buffer will grow if over the calculated size, if colorization
	//		 of the hex dump is enabled!
	buf.Grow((1 + ((len(data) - 1) / 16)) * 79)

	dumper := Dumper(&buf, colorize)
	dumper.Write(data)
	dumper.Close()
	return buf.String()
}

const hextable = "0123456789abcdef"

// HexEncode encodes src into EncodedLen(len(src))
// bytes of dst. As a convenience, it returns the number
// of bytes written to dst, but this value is always EncodedLen(len(src)).
// Encode implements hexadecimal encoding.
func HexEncode(dst, src []byte, colorize bool) ([]byte, int) {
	if colorize {
		return hexColorEncode(src)
	}

	j := 0

	for _, v := range src {
		dst[j] = hextable[v>>4]
		dst[j+1] = hextable[v&0x0f] // 0x0f => 0000 1111

		j += 2
	}

	return dst, len(src) * 2
}

// getCharColorType returns the type used to resolve to the correct color for byte v.
// If isHexRepres is true, it wil return a special type (TBase10).
// Set this to true if the hex representation is printed (not the text).
func getCharColorType(v byte, isHexRepres bool) Type {
	// choose color
	switch v {
	case 0x00:
		return TNULByte
	case ' ':
		return TWhitespaceChar
	case '.':
		return TPunctuationChar
	}

	// check if num of base 10 first
	if isHexRepres && v%10 == 0 {
		return TBase10
	} else if v >= 32 && v <= 126 { // printable ASCII character
		return TPrintable
	}

	// non-printable ASCII character
	return TNonPrintable
}

// hexColorEncode is like HexEncode but with colors.
func hexColorEncode(src []byte) ([]byte, int) {
	j := 0
	tmp := []byte{0x00, 0x00}
	dst := make([]byte, 0, len(src)*3)

	for _, v := range src {
		cType := getCharColorType(v, true)

		tmp[j] = hextable[v>>4]
		tmp[j+1] = hextable[v&0x0f] // 0x0f => 0000 1111

		dst = append(dst, hexColor(cType, tmp)...)

		j += 2
	}

	return dst, len(dst)
}

// Dumper returns a WriteCloser that writes a hex dump of all written data to
// w. The format of the dump matches the output of `hexdump -C` on the command
// line.
func Dumper(w io.Writer, colorize bool) io.WriteCloser {
	return &dumper{
		w:        w,
		colorize: colorize,
		buf:      make([]byte, 14),
	}
}

type dumper struct {
	w          io.Writer
	rightChars [18]byte
	buf        []byte
	used       int  // number of bytes in the current line
	n          uint // number of bytes, total
	closed     bool
	colorize   bool
}

func toChar(b byte) byte {
	if b < 32 || b > 126 {
		return '.'
	}

	return b
}

var nOffset int

func (h *dumper) Write(data []byte) (n int, err error) {
	if h.closed {
		return 0, errors.New("encoding/hex: dumper closed")
	}

	var dataPos int

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
			// ignore returned dst since we disabled colorization
			_, _ = HexEncode(h.buf[4:], h.buf[:4], false) // do not colorize offset
			h.buf[12] = ' '
			h.buf[13] = ' '

			_, err = h.w.Write(h.buf[4:14])
			if err != nil {
				return
			}

			nOffset++
		}

		h.buf, _ = HexEncode(h.buf[:], data[i:i+1], h.colorize) // colorize, if enabled, the hex value (mid)
		h.buf = append(h.buf, ' ')

		if h.used == 7 {
			// There's an additional space after the 8th byte.
			h.buf = append(h.buf, ' ')
		} else if h.used == 15 {
			// At the end of the line there's an extra space and
			// the bar for the right column.
			h.buf = append(h.buf, ' ', '|')
		}

		_, err = h.w.Write(h.buf)
		if err != nil {
			return
		}

		h.rightChars[h.used] = toChar(data[i])
		n++
		h.used++
		h.n++

		if h.used == 16 {
			if h.colorize {
				err = colorizeChars(h.w, data[dataPos:dataPos+16])
				dataPos += 16
			} else {
				h.rightChars[16] = '|'
				h.rightChars[17] = '\n'

				_, err = h.w.Write(h.rightChars[:])
			}

			if err != nil {
				return
			}

			h.used = 0
		}
	}

	return
}

func colorizeChars(w io.Writer, data []byte) error {
	var err error
	buf := make([]byte, 1)

	for i := range data {
		cType := getCharColorType(data[i], false)

		buf[0] = toChar(data[i])
		_, err = w.Write(hexColor(cType, buf))
		if err != nil {
			return err
		}
	}

	_, err = w.Write([]byte("|\n"))

	return err
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
