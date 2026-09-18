package asn1

import (
	"bytes"
	"encoding/hex"
	"testing"
)

type bitStringInner struct {
	A int
	B int
}

type bitStringOuter struct {
	Key bitStringInner `asn1:"bitstring"`
}

// The package's own encoding of a `bitstring` struct field (fields placed
// directly after the padding byte) must keep its exact bytes and round-trip.
func TestBitStringStructNativeFormUnchanged(t *testing.T) {
	want, _ := hex.DecodeString("3009030700020101020102")
	got, err := Marshal(bitStringOuter{bitStringInner{1, 2}})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("encoding changed: %x want %x", got, want)
	}
	var out bitStringOuter
	if _, err := Unmarshal(want, &out); err != nil || out.Key != (bitStringInner{1, 2}) {
		t.Fatalf("got %+v err %v", out, err)
	}
}

// X.509-style input wraps the fields in a SEQUENCE inside the BIT STRING; it
// used to fail with "tags don't match" and is now accepted.
func TestBitStringStructAcceptsNestedSequence(t *testing.T) {
	der, _ := hex.DecodeString("300b0309003006020101020102")
	var out bitStringOuter
	if _, err := Unmarshal(der, &out); err != nil || out.Key != (bitStringInner{1, 2}) {
		t.Fatalf("got %+v err %v", out, err)
	}
	// genuinely malformed content still reports the original error
	bad, _ := hex.DecodeString("30060304000401ff")
	if _, err := Unmarshal(bad, &out); err == nil {
		t.Error("malformed content accepted")
	}
}
