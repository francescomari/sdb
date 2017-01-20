package main

type format int

const (
	formatText format = iota
	formatHex
)

const (
	formatHexString  = "hex"
	formatTextString = "text"
)

func (f *format) String() string {
	switch *f {
	case formatHex:
		return formatHexString
	case formatText:
		return formatTextString
	default:
		panic("Invalid format")
	}
}

func (f *format) Set(s string) error {
	switch s {
	case formatHexString:
		*f = formatHex
	case formatTextString:
		*f = formatText
	default:
		return fmt.Errorf("Invalid format '%s'", s)
	}
	return nil
}
