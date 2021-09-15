package util

// very perlish, i know
func Chop(s *string) string {
	in := *s
	l := len(in)
	switch l {
	case 0:
		return ""
	default:
		*s = in[0 : l-1]
		return in[l-1:]
	}
}

func Shift(s *string) string {
	in := *s
	l := len(in)
	switch l {
	case 0:
		return ""
	default:
		*s = in[1:]
		return in[0:1]
	}
}
