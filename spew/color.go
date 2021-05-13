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

// Colors used for the type VALUE.
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
	TArgs
)

// Colors used for the TYPE string.
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
)

// pre-defined colors
var (
	cBlue = gcolor.HEX("#82aaff", false)
)

var colorPalette map[Type]gcolor.RGBColor = map[Type]gcolor.RGBColor{
	TInteger: cBlue,
	TFloat:   cBlue,
	TMap:     gcolor.LightRed.RGB(),
	TDate:    gcolor.Green.RGB(),
	TString:  gcolor.LightYellow.RGB(),
	TBool:    gcolor.LightBlue.RGB(),
	TNil:     gcolor.LightMagenta.RGB(),
	TArray:   gcolor.LightWhite.RGB(),
	TStruct:  gcolor.Yellow.RGB(),
	TArgs:    gcolor.LightRed.RGB(),

	TTInteger: gcolor.Cyan.RGB(),
	TTFloat:   gcolor.Cyan.RGB(),
	TTMap:     gcolor.LightRed.RGB(),
	TTDate:    gcolor.Green.RGB(),
	TTString:  gcolor.LightYellow.RGB(),
	TTBool:    gcolor.LightBlue.RGB(),
	TTNil:     gcolor.LightMagenta.RGB(),
	TTArray:   gcolor.White.RGB(),
	TTPtr:     gcolor.Magenta.RGB(),
}

type colorWriter struct {
	origWriter     io.Writer
	globalDisabled bool
	disabled       bool
	col            gcolor.RGBColor
}

func stopColor() {
	cWriter.disabled = true
}

func (c *colorWriter) Write(p []byte) (n int, err error) {
	if c.globalDisabled || c.disabled {
		return c.origWriter.Write(p)
	}

	str := c.col.Sprint(byteSlice2String(p))

	return c.origWriter.Write(s2b(str))
}

var cWriter *colorWriter

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
	case reflect.Map:
		cWriter.col = colorPalette[TMap]
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
