package spew

import (
	"reflect"

	"github.com/mgutz/ansi"
	"github.com/modern-go/reflect2"
)

// Type represents a GoLang (basic) type and semantics.
type Type int

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

var colorPalette map[Type]string = map[Type]string{
	TInteger: ansi.Blue,
	TFloat:   ansi.Blue,
	TMap:     ansi.LightRed,
	TDate:    ansi.Green,
	TString:  ansi.LightYellow,
	TBool:    ansi.LightBlue,
	TNil:     ansi.LightMagenta,

	TArray:  ansi.LightWhite,
	TStruct: ansi.Yellow,

	TArgs: ansi.LightRed,
}

func color(t reflect.Value) string {
	typ := reflect2.Type2(t.Type())

	// special case: value is nil
	switch typ.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr,
		reflect.UnsafePointer, reflect.Interface, reflect.Slice:
		if t.IsNil() {
			return colorPalette[TNil]
		}
	}

	switch typ.Kind() {
	case reflect.String:
		return colorPalette[TString]
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint,
		reflect.Complex64, reflect.Complex128:
		return colorPalette[TInteger]
	case reflect.Float32, reflect.Float64:
		return colorPalette[TFloat]
	case reflect.Map:
		return colorPalette[TMap]
	case reflect.Bool:
		return colorPalette[TBool]
	}

	return ""
}
