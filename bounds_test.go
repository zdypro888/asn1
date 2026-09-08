package asn1

import (
	"bytes"
	"reflect"
	"testing"
)

func TestIntegerDecodeRejectsOverflowWithoutChangingDestination(t *testing.T) {
	for _, target := range []any{new(uint8), new(uint16), new(uint32), new(int8), new(int16)} {
		t.Run(reflect.TypeOf(target).String(), func(t *testing.T) {
			v := reflect.ValueOf(target).Elem()
			bits := v.Type().Bits()
			unsigned := v.Kind() >= reflect.Uint && v.Kind() <= reflect.Uint64
			var input any
			if unsigned {
				v.SetUint(7)
				input = uint64(1) << bits
			} else {
				v.SetInt(7)
				input = int64(1) << (bits - 1)
			}
			der, err := Marshal(input)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := Unmarshal(der, target); err == nil {
				t.Fatal("accepted overflowing integer")
			}
			if unsigned && v.Uint() != 7 || !unsigned && v.Int() != 7 {
				t.Fatal("modified destination on error")
			}
		})
	}
}

func TestIntegerDecodeRoundTripsEverySupportedWidth(t *testing.T) {
	for _, input := range []any{uint8(255), uint16(65535), uint32(1<<32 - 1), uint64(1<<64 - 1), int8(-128), int16(-32768)} {
		der, err := Marshal(input)
		if err != nil {
			t.Fatal(err)
		}
		target := reflect.New(reflect.TypeOf(input))
		if _, err := Unmarshal(der, target.Interface()); err != nil {
			t.Fatalf("%T: %v", input, err)
		}
		if !reflect.DeepEqual(input, target.Elem().Interface()) {
			t.Fatalf("%T did not round trip", input)
		}
	}
}

func TestUnsignedIntegerRejectsNegativeDER(t *testing.T) {
	for _, value := range []int64{-1, -128, -129, -1 << 63} {
		der, err := Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		var target uint64 = 7
		if _, err := Unmarshal(der, &target); err == nil {
			t.Fatalf("accepted negative %d as %d", value, target)
		}
		if target != 7 {
			t.Fatal("modified target on error")
		}
	}
}

func TestAnyExplicitWrapperCannotConsumeSibling(t *testing.T) {
	// The INTEGER declares two bytes, but its explicit [0] wrapper contains
	// only one. The next sibling must never supply the missing byte.
	var target struct {
		Value any `asn1:"explicit,tag:0"`
	}
	if _, err := Unmarshal([]byte{0x30, 0x06, 0xa0, 0x03, 0x02, 0x02, 0x01, 0x02}, &target); err == nil {
		t.Fatal("accepted an explicit value extending into its sibling")
	}
}

func TestTypedExplicitWrapperIsBounded(t *testing.T) {
	for _, der := range [][]byte{
		{0x30, 0x06, 0xa0, 0x03, 0x02, 0x02, 0x01, 0x02},
		{0x30, 0x05, 0xa0, 0x7f, 0x30, 0x01, 0x00},
	} {
		var target struct {
			Value []byte `asn1:"explicit,tag:0"`
		}
		if _, err := Unmarshal(der, &target); err == nil {
			t.Fatalf("accepted invalid explicit bytes: %x", der)
		}
	}
	var number struct {
		Value int `asn1:"explicit,tag:0"`
	}
	if _, err := Unmarshal([]byte{0x30, 0x06, 0xa0, 0x03, 0x02, 0x02, 0x01, 0x02}, &number); err == nil {
		t.Fatal("typed INTEGER consumed sibling bytes")
	}
}

func TestOptionalPointerUsesFullTagMatching(t *testing.T) {
	text := "你好"
	input := struct {
		Text   *string `asn1:"optional"`
		Number int
	}{&text, 42}
	der, err := Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	var output struct {
		Text   *string `asn1:"optional"`
		Number int
	}
	if _, err := Unmarshal(der, &output); err != nil {
		t.Fatal(err)
	}
	if output.Text == nil || *output.Text != text || output.Number != 42 {
		t.Fatalf("bad output: %+v", output)
	}
}

func FuzzUnmarshalBounds(f *testing.F) {
	f.Add([]byte{0x30, 0x03, 0x02, 0x01, 0x01})
	f.Add([]byte{0x30, 0x05, 0xa0, 0x7f, 0x30, 0x01, 0x00})
	f.Fuzz(func(t *testing.T, data []byte) {
		var anyValue any
		_, _ = Unmarshal(data, &anyValue)
		var explicit struct {
			Value []byte `asn1:"explicit,tag:0"`
		}
		_, _ = Unmarshal(data, &explicit)
		var integers []uint32
		_, _ = Unmarshal(data, &integers)
	})
}

func TestAnyPrivateExplicitRoundTrip(t *testing.T) {
	input := struct {
		Value any `asn1:"private,explicit,tag:5"`
	}{int64(42)}
	der, err := Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	var target struct {
		Value any `asn1:"private,explicit,tag:5"`
	}
	if _, err := Unmarshal(der, &target); err != nil {
		t.Fatal(err)
	}
	if target.Value != input.Value {
		t.Fatalf("got %v", target.Value)
	}
}

func TestAnyDecodeLimitsNesting(t *testing.T) {
	der := []byte{0x05, 0x00}
	for range 10001 {
		header := appendTagAndLength(nil, tagAndLength{class: ClassUniversal, tag: TagSequence, length: len(der), isCompound: true})
		der = append(header, der...)
	}
	var target any
	if _, err := Unmarshal(der, &target); err == nil {
		t.Fatal("accepted excessive nesting")
	}
}

func TestObjectIdentifierRejectsInvalidArcs(t *testing.T) {
	for _, oid := range []ObjectIdentifier{{-1, 1}, {1, -1}, {1, 2, -1}, {2, int(^uint(0) >> 1)}} {
		if der, err := Marshal(oid); err == nil {
			t.Fatalf("accepted %v: %x", oid, der)
		}
	}
	der, err := Marshal(ObjectIdentifier{2, 999, 3})
	if err != nil {
		t.Fatal(err)
	}
	var target ObjectIdentifier
	if _, err := Unmarshal(der, &target); err != nil {
		t.Fatal(err)
	}
	again, err := Marshal(target)
	if err != nil || !bytes.Equal(der, again) {
		t.Fatal("valid OID round trip failed")
	}
}
