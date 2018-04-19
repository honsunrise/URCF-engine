package utils

import (
	"database/sql/driver"
	"errors"
	"strconv"
	"strings"
)

type DetailCompareResult int32

const (
	NoDifferent                     = 0
	MajorLT     DetailCompareResult = 1 << iota
	MajorGT
	MinorLT
	MinorGT
	PatchLT
	PatchGT
	PreReleaseLT
	PreReleaseGT
	BuildLT
	BuildGT
)

type CompareResult int32

const (
	Same CompareResult = iota
	GT
	LT
)

type SemanticVersion struct {
	Major      uint32
	Minor      uint32
	Patch      uint32
	PreRelease []string
	Build      []string
	valid      bool
}

func (ver SemanticVersion) Value() (driver.Value, error) {
	return ver.String(), nil
}

func (ver *SemanticVersion) Scan(value interface{}) (err error) {
	var tmp *SemanticVersion
	switch value.(type) {
	case string:
		tmp, err = NewSemVerFromString(value.(string))
		*ver = *tmp
	case []byte:
		tmp, err = NewSemVerFromString(string(value.([]byte)))
		*ver = *tmp
	default:
		return errors.New("failed to scan SemanticVersion")
	}
	return nil
}

func SemanticVersionMust(semVer *SemanticVersion, err error) *SemanticVersion {
	if err != nil {
		panic(err)
	}
	return semVer
}

func NewSemVerFromString(ver string) (semVer *SemanticVersion, err error) {
	semVer = &SemanticVersion{
		Major: 0, Minor: 0, Patch: 0, valid: false,
	}
	tmp := strings.SplitN(ver, "+", 2)
	if len(tmp) == 2 {
		semVer.Build = strings.Split(tmp[1], ".")
		ver = tmp[0]
	} else if len(tmp) != 1 {
		err = errors.New("semantic version format error")
		return
	}
	tmp = strings.SplitN(ver, "-", 2)
	if len(tmp) == 2 {
		semVer.PreRelease = strings.Split(tmp[1], ".")
		ver = tmp[0]
	} else if len(tmp) != 1 {
		err = errors.New("semantic version format error")
		return
	}
	tmp = strings.SplitN(ver, ".", 3)
	if len(tmp) == 3 {
		ui64 := uint64(0)
		ui64, err = strconv.ParseUint(tmp[2], 10, 0)
		if err != nil {
			return
		}
		semVer.Patch = uint32(ui64)

		ui64, err = strconv.ParseUint(tmp[1], 10, 0)
		if err != nil {
			return
		}
		semVer.Minor = uint32(ui64)

		ui64, err = strconv.ParseUint(tmp[0], 10, 0)
		if err != nil {
			return
		}
		semVer.Major = uint32(ui64)
		semVer.valid = true
	} else if len(tmp) != 1 {
		err = errors.New("semantic version format error")
		return
	}
	return
}

func arrayCompare(this []string, other []string) int {
	thisLeft := len(this)
	OtherLeft := len(other)
	if thisLeft == 0 && OtherLeft > 0 {
		return 1
	} else if thisLeft > 0 && OtherLeft == 0 {
		return -1
	}
	i := 0
	for {
		if thisLeft == 0 && OtherLeft > 0 {
			return -1
		} else if thisLeft > 0 && OtherLeft == 0 {
			return 1
		} else if thisLeft == 0 && OtherLeft == 0 {
			return 0
		} else {
			ui64this, err := strconv.ParseUint(this[i], 10, 0)
			if err == nil {
				ui64other, err := strconv.ParseUint(other[i], 10, 0)
				if err == nil {
					if ui64this < ui64other {
						return -1
					} else if ui64this > ui64other {
						return 1
					}
					return 0
				}
			}

			result := strings.Compare(this[i], other[i])
			if result != 0 {
				return result
			}
			thisLeft -= 1
			OtherLeft -= 1
			i++
		}
	}
}

func (semVer *SemanticVersion) String() string {
	var ret string
	ret += strconv.FormatInt(int64(semVer.Major), 10)
	ret += "."
	ret += strconv.FormatInt(int64(semVer.Minor), 10)
	ret += "."
	ret += strconv.FormatInt(int64(semVer.Patch), 10)
	if len(semVer.PreRelease) > 0 {
		ret += "-"
		ret += strings.Join(semVer.PreRelease, ".")
	}
	if len(semVer.Build) > 0 {
		ret += "+"
		ret += strings.Join(semVer.Build, ".")
	}
	return ret
}

func (semVer *SemanticVersion) DetailCompare(other *SemanticVersion) DetailCompareResult {
	if !other.valid || !semVer.valid {
		panic(errors.New("invalid SemanticVersion can't compare"))
	}
	var result DetailCompareResult = NoDifferent

	if semVer.Major < other.Major {
		result |= MajorLT
	} else if semVer.Major > other.Major {
		result |= MajorGT
	}

	if semVer.Minor < other.Minor {
		result |= MinorLT
	} else if semVer.Minor > other.Minor {
		result |= MinorGT
	}

	if semVer.Patch < other.Patch {
		result |= PatchLT
	} else if semVer.Patch > other.Patch {
		result |= PatchGT
	}

	cr := arrayCompare(semVer.PreRelease, other.PreRelease)
	if cr < 0 {
		result |= PreReleaseLT
	} else if cr > 0 {
		result |= PreReleaseGT
	}

	cr = arrayCompare(semVer.Build, other.Build)
	if cr < 0 {
		result |= BuildLT
	} else if cr > 0 {
		result |= BuildGT
	}

	return result
}

func (semVer *SemanticVersion) Compare(other *SemanticVersion) CompareResult {
	result := semVer.DetailCompare(other)
	if (result & MajorLT) != 0 {
		return LT
	} else if (result & MajorGT) != 0 {
		return GT
	} else {
		if (result & MinorLT) != 0 {
			return LT
		} else if (result & MinorGT) != 0 {
			return GT
		} else {
			if (result & PatchLT) != 0 {
				return LT
			} else if (result & PatchGT) != 0 {
				return GT
			} else {
				if (result & PreReleaseLT) != 0 {
					return LT
				} else if (result & PreReleaseGT) != 0 {
					return GT
				} else {
					return Same
				}
			}
		}
	}
}

func (semVer *SemanticVersion) Compatible(other *SemanticVersion) bool {
	result := semVer.DetailCompare(other)
	if (result&MajorLT) != 0 || (result&MajorGT) != 0 {
		return false
	} else {
		if (result & MinorGT) != 0 {
			return false
		} else {
			return true
		}
	}
}
