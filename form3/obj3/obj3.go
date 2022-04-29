package obj3

type CylinderStyle int

const (
	_ CylinderStyle = iota
	CylinderCircular
	CylinderHex
	CylinderKnurl
)

func (c CylinderStyle) String() (str string) {
	switch c {
	case CylinderCircular:
		str = "circular"
	case CylinderHex:
		str = "hex"
	case CylinderKnurl:
		str = "knurl"
	default:
		str = "unknown"
	}
	return str
}
