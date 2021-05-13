package main

import (
	"bytes"

	"github.com/l0nax/go-spew/spew"
)

type myType struct {
	Str        string
	Number     int
	NNumber    float32
	Map        map[string]interface{}
	NestedType *myType
	Bytes      []byte
	Ref        interface{}
}

func main() {
	master := myType{
		Str:     "Super cool String",
		Number:  432534,
		NNumber: 453.3453,
		Map:     make(map[string]interface{}),
		NestedType: &myType{
			Str:    "Second super cool string",
			Number: 12312,
			Ref:    nil,
			Bytes:  []byte("Hello World"),
		},
		Bytes: []byte("Foo Bar"),
		Ref:   bytes.NewBuffer([]byte("test")),
	}

	master.Map["Hello"] = "World"
	master.Map["Foo"] = 15
	master.Map["MyFloat"] = 345345.345
	master.Map["False?"] = false
	master.Map["True?"] = true

	spew.Dump(master)
}
