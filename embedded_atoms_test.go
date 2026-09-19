package asn1

import (
	"bytes"
	"testing"
	"time"
)

type withEmbeddedStruct struct {
	EmbeddedBase
	C int
}

type EmbeddedBase struct {
	A int
	B int
}

type withEmbeddedTime struct {
	time.Time
	N int
}

type withEmbeddedBitString struct {
	BitString
	N int
}

// Ordinary embedded structs are still flattened into the parent, byte for byte.
func TestEmbeddedStructStillInlined(t *testing.T) {
	got, err := Marshal(withEmbeddedStruct{EmbeddedBase{1, 2}, 3})
	if err != nil {
		t.Fatal(err)
	}
	want := []byte{0x30, 0x09, 0x02, 0x01, 0x01, 0x02, 0x01, 0x02, 0x02, 0x01, 0x03}
	if !bytes.Equal(got, want) {
		t.Fatalf("got % x want % x", got, want)
	}
	var out withEmbeddedStruct
	if _, err := Unmarshal(got, &out); err != nil || out.A != 1 || out.B != 2 || out.C != 3 {
		t.Fatalf("got %+v err %v", out, err)
	}
}

// Embedded time.Time / BitString are values, not field lists: they used to be
// flattened into raw bytes that could not be decoded again.
func TestEmbeddedAtomsRoundTrip(t *testing.T) {
	when := time.Date(2026, 9, 19, 1, 2, 3, 0, time.UTC)
	der, err := Marshal(withEmbeddedTime{when, 5})
	if err != nil {
		t.Fatal(err)
	}
	if der[2] != TagUTCTime {
		t.Fatalf("expected a UTCTime element, got % x", der)
	}
	var gotTime withEmbeddedTime
	if _, err := Unmarshal(der, &gotTime); err != nil || !gotTime.Time.Equal(when) || gotTime.N != 5 {
		t.Fatalf("time: %+v err %v", gotTime, err)
	}

	bits := BitString{Bytes: []byte{0xA0}, BitLength: 3}
	der, err = Marshal(withEmbeddedBitString{bits, 7})
	if err != nil {
		t.Fatal(err)
	}
	var gotBits withEmbeddedBitString
	if _, err := Unmarshal(der, &gotBits); err != nil || gotBits.BitLength != 3 || gotBits.N != 7 {
		t.Fatalf("bitstring: %+v err %v", gotBits, err)
	}
}
