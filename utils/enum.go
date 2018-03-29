package utils

import "strconv"

type IntName struct {
	i uint32
	s string
}

func StringName(i uint32, names []IntName, goSyntaxPre string, goSyntax bool) string {
	for _, n := range names {
		if n.i == i {
			if goSyntax {
				return goSyntaxPre + n.s
			}
			return n.s
		}
	}

	// second pass - look for smaller to add with.
	// assume sorted already
	for j := len(names) - 1; j >= 0; j-- {
		n := names[j]
		if n.i < i {
			s := n.s
			if goSyntax {
				s = goSyntaxPre + s
			}
			return s + "+" + strconv.FormatUint(uint64(i-n.i), 10)
		}
	}

	return strconv.FormatUint(uint64(i), 10)
}
