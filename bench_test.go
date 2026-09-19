package asn1

import (
	"encoding/pem"
	"os"
	"testing"
	"time"
)

type benchTBS struct {
	Version  int `asn1:"optional,explicit,default:0,tag:0"`
	Serial   RawValue
	SigAlg   RawValue
	Issuer   RawValue
	Validity struct{ NotBefore, NotAfter time.Time }
	Subject  RawValue
	SPKI     RawValue
	Rest     []RawValue `asn1:"optional"`
}
type benchCertificate struct {
	TBS    benchTBS
	SigAlg RawValue
	Sig    BitString
}

func loadBenchCerts(b *testing.B) [][]byte {
	data, err := os.ReadFile("/etc/ssl/cert.pem")
	if err != nil {
		b.Skip(err)
	}
	var ders [][]byte
	for len(ders) < 100 {
		var block *pem.Block
		if block, data = pem.Decode(data); block == nil {
			break
		}
		ders = append(ders, block.Bytes)
	}
	return ders
}

func BenchmarkUnmarshalCertStruct(b *testing.B) {
	ders := loadBenchCerts(b)
	b.ReportAllocs()
	for b.Loop() {
		for _, der := range ders {
			var c benchCertificate
			if _, err := Unmarshal(der, &c); err != nil {
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkUnmarshalCertAny(b *testing.B) {
	ders := loadBenchCerts(b)
	b.ReportAllocs()
	for b.Loop() {
		for _, der := range ders {
			var v any
			if _, err := Unmarshal(der, &v); err != nil {
				b.Fatal(err)
			}
		}
	}
}
