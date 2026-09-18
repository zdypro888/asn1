package asn1

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"testing"
)

// INTEGERs wider than 8 bytes (certificate serial numbers) used to make
// decoding into any fail; they now decode to *big.Int and re-encode identically.
func TestAnyDecodesWideInteger(t *testing.T) {
	der, _ := hex.DecodeString("3012" + "020d00f1e2d3c4b5a6978877665544" + "020105")
	var v any
	rest, err := Unmarshal(der, &v)
	if err != nil || len(rest) != 0 {
		t.Fatalf("Unmarshal: %v rest=%x", err, rest)
	}
	cv, ok := v.(*CompoundValue)
	if !ok || len(cv.Items) != 2 {
		t.Fatalf("got %#v", v)
	}
	want, _ := new(big.Int).SetString("f1e2d3c4b5a6978877665544", 16)
	if n, ok := cv.Items[0].(*big.Int); !ok || n.Cmp(want) != 0 {
		t.Errorf("wide integer: got %#v", cv.Items[0])
	}
	// narrow integers keep their historical int64 type
	if n, ok := cv.Items[1].(int64); !ok || n != 5 {
		t.Errorf("narrow integer: got %#v", cv.Items[1])
	}
	out, err := Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out, der) {
		t.Errorf("re-encoded %x want %x", out, der)
	}
}

func TestAnyIntegerBoundary(t *testing.T) {
	// exactly 8 bytes stays int64; non-minimal wide encodings are still rejected
	var v any
	if _, err := Unmarshal([]byte{0x02, 0x08, 0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, &v); err != nil {
		t.Fatal(err)
	} else if _, ok := v.(int64); !ok {
		t.Errorf("got %T", v)
	}
	if _, err := Unmarshal([]byte{0x02, 0x09, 0x00, 0x00, 1, 2, 3, 4, 5, 6, 7}, &v); err == nil {
		t.Error("non-minimal integer accepted")
	}
}
