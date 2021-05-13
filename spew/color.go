package spew

import (
	"io"
	"reflect"
	"unsafe"

	gcolor "github.com/gookit/color"
	"github.com/modern-go/reflect2"
)

// Type represents a GoLang (basic) type and semantics.
type Type int

// type definition used to represent type VALUES.
const (
	TInteger Type = iota + 1
	TFloat
	TMap
	TDate
	TString
	TBool
	TNil
	TArray
	TStruct
)

// type definition used to represent TYPE string.
const (
	TTInteger Type = iota + 100
	TTFloat
	TTMap
	TTDate
	TTString
	TTBool
	TTNil
	TTArray
	TTPtr
	TTAddress
	TTInterface
)

// representation for special cases
const (
	TLen Type = iota + 200
	TCap
	TArgs
)

// pre-defined colors
var (
	cBlue          = gcolor.HEX("#82aaff", false)
	cPurple        = gcolor.HEX("#c792ea", false)
	cOrange        = gcolor.HEX("#f78c6c", false)
	cSpecial       = gcolor.HEXStyle("#ffffff", "#c17e70")
	cGreen         = gcolor.HEX("#c3e88d", false)
	cDarkGreen     = gcolor.HEX("#138040", false)
	cRadiantYellow = gcolor.HEX("#ffea00", false)
)

type ColorPrinter interface {
	Sprint(a ...interface{}) string
}

var colorPalette = map[Type]ColorPrinter{
	TInteger: cBlue,
	TFloat:   cBlue,
	TDate:    gcolor.Green.RGB(),
	TString:  gcolor.LightYellow.RGB(),
	TBool:    gcolor.LightBlue.RGB(),
	TNil:     gcolor.LightMagenta.RGB(),
	TArray:   gcolor.LightWhite.RGB(),
	TStruct:  gcolor.Yellow.RGB(),

	TTInteger:   gcolor.Cyan.RGB(),
	TTFloat:     gcolor.Cyan.RGB(),
	TTMap:       gcolor.LightRed.RGB(),
	TTDate:      gcolor.Green.RGB(),
	TTString:    cOrange,
	TTBool:      cPurple,
	TTArray:     gcolor.White.RGB(),
	TTPtr:       gcolor.Magenta.RGB(),
	TTAddress:   cDarkGreen,
	TTInterface: cRadiantYellow,

	TLen:  cSpecial,
	TCap:  cSpecial,
	TArgs: cGreen,
}

type colorWriter struct {
	origWriter     io.Writer
	globalDisabled bool
	disabled       bool
	col            ColorPrinter
}

func stopColor() {
	cWriter.disabled = true
}

func (c *colorWriter) Write(p []byte) (n int, err error) {
	if c.globalDisabled || c.disabled || c.col == nil {
		return c.origWriter.Write(p)
	}

	str := c.col.Sprint(byteSlice2String(p))

	return c.origWriter.Write(s2b(str))
}

var cWriter *colorWriter

// rawColor allows to use our internal types DIRECTLY.
// Please ONLY use this function if the type is known – skipping the overhead
// of the reflect package (calling the methods and "searching" the correct color).
func rawColor(t Type) {
	col, ok := colorPalette[t]

	cWriter.disabled = !ok
	cWriter.col = col
}

func specialColor(t Type) {
	switch t {
	case TLen, TCap, TArgs:
		cWriter.disabled = false
		cWriter.col = colorPalette[t]
	}
}

// colorPtr handles the special case where we're searching the color for pointers.
// This function exists since there are some special cases, e.g. time.Time or bytes.Buffer,
// which have a different color than a "normal" pointer.
func colorPtr(t string) {
	switch t {
	case "*bytes.Buffer":
		// require special color
		fallthrough
	default:
		cWriter.disabled = false
		cWriter.col = colorPalette[TTPtr]
	}
}

func color(t reflect.Value) {
	cWriter.disabled = false
	typ := reflect2.Type2(t.Type())

	// special case: value is nil
	switch typ.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr,
		reflect.UnsafePointer, reflect.Interface, reflect.Slice:
		if t.IsNil() {
			cWriter.col = colorPalette[TNil]
			return
		}
	}

	switch typ.Kind() {
	case reflect.String:
		cWriter.col = colorPalette[TString]
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint,
		reflect.Complex64, reflect.Complex128:
		cWriter.col = colorPalette[TInteger]
	case reflect.Float32, reflect.Float64:
		cWriter.col = colorPalette[TFloat]
	case reflect.Bool:
		cWriter.col = colorPalette[TBool]
	default:
		cWriter.disabled = true
	}
}

func typeColor(t reflect.Type) {
	cWriter.disabled = false

	switch t.Kind() {
	case reflect.Ptr:
		cWriter.col = colorPalette[TTPtr]
	case reflect.String:
		cWriter.col = colorPalette[TTString]
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint,
		reflect.Complex64, reflect.Complex128:
		cWriter.col = colorPalette[TTInteger]
	case reflect.Float32, reflect.Float64:
		cWriter.col = colorPalette[TTFloat]
	case reflect.Map:
		cWriter.col = colorPalette[TTMap]
	case reflect.Bool:
		cWriter.col = colorPalette[TTBool]
	case reflect.Interface:
		cWriter.col = colorPalette[TTInterface]
	default:
		cWriter.disabled = true
	}
}

// byteSlice2String converts a byte slice to a string in a performant way.
func byteSlice2String(bs []byte) string {
	return *(*string)(unsafe.Pointer(&bs))
}

// s2b converts string to a byte slice without memory allocation.
//
// Note it may break if string and/or slice header will change
// in the future go versions.
func s2b(s string) (b []byte) {
	/* #nosec G103 */
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	/* #nosec G103 */
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh.Data = sh.Data
	bh.Len = sh.Len
	bh.Cap = sh.Len
	return b
}
