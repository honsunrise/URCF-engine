package utils

import "strconv"

type IntName struct {
	I uint32
	S string
}

func StringName(i uint32, names []IntName, goSyntaxPre string, goSyntax bool) string {
	for _, n := range names {
		if n.I == i {
			if goSyntax {
				return goSyntaxPre + n.S
			}
			return n.S
		}
	}

	// second pass - look for smaller to add with.
	// assume sorted already
	for j := len(names) - 1; j >= 0; j-- {
		n := names[j]
		if n.I < i {
			s := n.S
			if goSyntax {
				s = goSyntaxPre + s
			}
			return s + "+" + strconv.FormatUint(uint64(i-n.I), 10)
		}
	}

	return strconv.FormatUint(uint64(i), 10)
}
