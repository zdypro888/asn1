// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package asn1

import (
	"log"
	"testing"
)

type AppInline struct {
	Num int32
}
type ASNS struct {
	App *AppInline `asn1:"application,tag:1"`
}

func TestMarshal(t *testing.T) {
	var asn ASNS
	asn.App = &AppInline{Num: 0x12345678}
	data, err := MarshalWithParams(asn, "application,tag:1")
	if err != nil {
		t.Error(err)
	}
	log.Printf("%x", data)
	t.Error("")

}
