package asn1

import (
	"fmt"
	"strings"
)

// PrintAny 递归打印 ASN.1 解析后的 any 值，带缩进。
func PrintAny(b *strings.Builder, v any, indent string) {
	switch val := v.(type) {
	case *CompoundValue:
		className := "UNIVERSAL"
		switch val.Class {
		case ClassContextSpecific:
			className = fmt.Sprintf("CTX[%d]", val.Tag)
		case ClassApplication:
			className = fmt.Sprintf("APP[%d]", val.Tag)
		case ClassPrivate:
			className = fmt.Sprintf("PRI[%d]", val.Tag)
		default:
			switch val.Tag {
			case TagSequence:
				className = "SEQUENCE"
			case TagSet:
				className = "SET"
			default:
				className = fmt.Sprintf("TAG(%d)", val.Tag)
			}
		}
		fmt.Fprintf(b, "%s%s {\n", indent, className)
		for _, item := range val.Items {
			PrintAny(b, item, indent+"  ")
		}
		fmt.Fprintf(b, "%s}\n", indent)
	case ObjectIdentifier:
		fmt.Fprintf(b, "%sOID: %s\n", indent, val.String())
	case bool:
		fmt.Fprintf(b, "%sBOOLEAN: %v\n", indent, val)
	case int64:
		fmt.Fprintf(b, "%sINTEGER: %d\n", indent, val)
	case string:
		fmt.Fprintf(b, "%sSTRING: %q\n", indent, val)
	case []byte:
		if len(val) <= 32 {
			fmt.Fprintf(b, "%sOCTET STRING: %x\n", indent, val)
		} else {
			fmt.Fprintf(b, "%sOCTET STRING: %d bytes\n", indent, len(val))
		}
	case BitString:
		fmt.Fprintf(b, "%sBIT STRING: %d bits\n", indent, val.BitLength)
	case RawValue:
		fmt.Fprintf(b, "%sRAW[class=%d,tag=%d]: %x\n", indent, val.Class, val.Tag, val.Bytes)
	default:
		fmt.Fprintf(b, "%s%v\n", indent, val)
	}
}
