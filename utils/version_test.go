package utils

import (
	"testing"
	"errors"
)


// 1.0.0-alpha < 1.0.0-alpha.1 < 1.0.0-alpha.beta < 1.0.0-beta < 1.0.0-beta.2 < 1.0.0-beta.11111
// < 1.0.0-beta.k < 1.0.0-rc.1 < 1.0.0
func TestSemanticVersion_Compare(t *testing.T) {
	testList := []string {
		"1.0.0-alpha",
		"1.0.0-alpha.1",
		"1.0.0-alpha.beta",
		"1.0.0-beta",
		"1.0.0-beta.2",
		"1.0.0-beta.11111",
		"1.0.0-beta.k",
		"1.0.0-rc.1",
		"1.0.0",
	}

	for i := 0; i < len(testList); i++ {
		for j := 0; j < len(testList); j++ {
			v1, err := NewSemVerFromString(testList[i])
			if err != nil {
				t.Fatal(err)
			}
			v2, err := NewSemVerFromString(testList[j])
			if err != nil {
				t.Fatal(err)
			}
			if (i == j) {
				if v1.Compare(v2) != Same {
					t.Fatal(errors.New("Version compare not correct"))
				}
			} else if i < j {
				if v1.Compare(v2) != LT {
					t.Fatal(errors.New("Version compare not correct"))
				}
			} else {
				if v1.Compare(v2) != GT {
					t.Fatal(errors.New("Version compare not correct"))
				}
			}
		}
	}


}
