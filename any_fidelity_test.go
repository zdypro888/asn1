package asn1

import (
	"bytes"
	"encoding/hex"
	"testing"
)

// Decoding into any and encoding again must reproduce the input: string and
// time types used to be normalised to UTF8String / UTCTime.
func TestAnyReencodeKeepsStringAndTimeTypes(t *testing.T) {
	for name, in := range map[string]string{
		"ia5":         "30051603614062",
		"printable":   "30051303616263",
		"numeric":     "30051203313233",
		"utf8":        "30050c03616263",
		"utctime":     "300f170d3236303931393031303230335a",
		"generalized": "3011180f32303236303931393031303230335a",
		"mixed":       "3028" + "16036140620c03616263180f32303236303931393031303230335a020105300813036162630101ff",
	} {
		der, _ := hex.DecodeString(in)
		var v any
		if rest, err := Unmarshal(der, &v); err != nil || len(rest) != 0 {
			t.Fatalf("%s: Unmarshal: %v", name, err)
		}
		out, err := Marshal(v)
		if err != nil {
			t.Fatalf("%s: Marshal: %v", name, err)
		}
		if !bytes.Equal(out, der) {
			t.Errorf("%s: re-encoded %x, want %x", name, out, der)
		}
	}
}

// Values built by hand carry no ItemTags and encode exactly as before.
func TestCompoundValueWithoutItemTagsUnchanged(t *testing.T) {
	out, err := Marshal(CompoundValue{Class: ClassUniversal, Tag: TagSequence, Items: []any{"abc", int64(5)}})
	if err != nil {
		t.Fatal(err)
	}
	want, _ := hex.DecodeString("30081303616263020105") // auto-selected PrintableString, as before
	if !bytes.Equal(out, want) {
		t.Errorf("got %x want %x", out, want)
	}
	// a caller that replaces a decoded string with another type is not forced
	// into the recorded string type
	var v any
	der, _ := hex.DecodeString("30051603614062")
	if _, err := Unmarshal(der, &v); err != nil {
		t.Fatal(err)
	}
	cv := v.(*CompoundValue)
	cv.Items[0] = int64(7)
	if out, err := Marshal(cv); err != nil || !bytes.Equal(out, []byte{0x30, 0x03, 0x02, 0x01, 0x07}) {
		t.Errorf("replaced item: %x %v", out, err)
	}
}
