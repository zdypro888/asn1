package asn1

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

// legacyParseUTCTime is the previous implementation, kept verbatim.
func legacyParseUTCTime(bytes []byte) (ret time.Time, err error) {
	s := string(bytes)

	formatStr := "0601021504Z0700"
	ret, err = time.Parse(formatStr, s)
	if err != nil {
		formatStr = "060102150405Z0700"
		ret, err = time.Parse(formatStr, s)
	}
	if err != nil {
		return
	}

	if serialized := ret.Format(formatStr); serialized != s {
		err = fmt.Errorf("asn1: time did not serialize back to the original value and may be invalid: given %q, but serialized as %q", s, serialized)
		return
	}

	if ret.Year() >= 2050 {
		ret = ret.AddDate(-100, 0, 0)
	}

	return
}

func TestParseUTCTimeMatchesLegacy(t *testing.T) {
	r := rand.New(rand.NewSource(3))
	inputs := []string{"", "Z", "260919010203Z", "2609190102Z", "260919010203+0800", "2609190102-0700",
		"500101000000Z", "491231235959Z", "260230010203Z", "26091901020Z", "2609190102033Z", "260919010203z",
		"260919250203Z", "2609190102036Z", "260919010203+08000", "aaaaaaaaaaaaZ", "260919010260Z"}
	const alphabet = "0123456789Z+-:. a"
	for i := 0; i < 20000; i++ {
		n := []int{11, 13, 15, 17, r.Intn(20)}[r.Intn(5)]
		b := make([]byte, n)
		for j := range b {
			if r.Intn(8) == 0 {
				b[j] = alphabet[r.Intn(len(alphabet))]
			} else {
				b[j] = byte('0' + r.Intn(10))
			}
		}
		if n > 0 && r.Intn(2) == 0 {
			b[n-1] = 'Z'
		}
		inputs = append(inputs, string(b))
	}
	for _, in := range inputs {
		got, gotErr := parseUTCTime([]byte(in))
		want, wantErr := legacyParseUTCTime([]byte(in))
		if (gotErr == nil) != (wantErr == nil) || (gotErr != nil && gotErr.Error() != wantErr.Error()) || !got.Equal(want) || got.Location().String() != want.Location().String() {
			t.Fatalf("%q: got (%v, %v), want (%v, %v)", in, got, gotErr, want, wantErr)
		}
	}
}
