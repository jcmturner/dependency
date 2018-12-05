package version

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Semantic struct {
	major         int
	minor         int
	patch         int
	majorExplicit bool // Was the major version explicitly defined?
	minorExplicit bool // Was the minor version explicitly defined?
	patchExplicit bool // Was the patch version explicitly defined?
}

// NewSemantic creates a new Semantic version. If any of the values are to be undefined pass nil.
func NewSemantic(mj, mi, p *int) (Semantic, error) {
	var s Semantic
	if mj != nil {
		s.major = *mj
		s.majorExplicit = true
	}
	if mi != nil {
		if !s.majorExplicit {
			return s, errors.New("invalid to have an explicit minor version with a generic major version")
		}
		s.minor = *mi
		s.minorExplicit = true
	}
	if p != nil {
		if !s.majorExplicit || !s.minorExplicit {
			return s, errors.New("invalid to have an explicit patch version with a generic major or minor version")
		}
		s.patch = *p
		s.patchExplicit = true
	}
	return s, nil
}

func ParseSemantic(v string) (Semantic, error) {
	var s Semantic
	vs := strings.SplitN(v, ".", 3)
	if vs[0] != "" && vs[0] != "x" {
		n, err := strconv.Atoi(vs[0])
		if err != nil {
			return s, fmt.Errorf("failed to parse major version: %v", err)
		}
		s.major = n
		s.majorExplicit = true
	}
	if vs[1] != "" && vs[1] != "x" {
		if !s.majorExplicit {
			return s, errors.New("invalid to have an explicit minor version with a generic major version")
		}
		n, err := strconv.Atoi(vs[1])
		if err != nil {
			return s, fmt.Errorf("failed to parse minor version: %v", err)
		}
		s.minor = n
		s.minorExplicit = true
	}
	if vs[2] != "" && vs[2] != "x" {
		if !s.majorExplicit || !s.minorExplicit {
			return s, errors.New("invalid to have an explicit patch version with a generic major or minor version")
		}
		n, err := strconv.Atoi(vs[2])
		if err != nil {
			return s, fmt.Errorf("failed to parse patch version: %v", err)
		}
		s.patch = n
		s.patchExplicit = true

	}
	return s, nil
}

func (s *Semantic) Major() (int, bool) {
	return s.major, s.majorExplicit
}

func (s *Semantic) Minor() (int, bool) {
	return s.minor, s.minorExplicit
}

func (s *Semantic) Patch() (int, bool) {
	return s.patch, s.patchExplicit
}

//// GreaterThan is equalent to: Is s > v ?
//func (s *Semantic) GreaterThan(v Semantic) bool {
//	if s.major > v.major {
//		return true
//	}
//	if s.major == v.major && s.minor > v.minor {
//		return true
//	}
//	if s.major == v.major && s.minor == v.minor && s.patch > v.patch {
//		return true
//	}
//	return false
//}
//
//// LessThan is equalent to: Is s < v ?
//func (s *Semantic) LessThan(v Semantic) bool {
//	return v.GreaterThan(*s)
//}

func (s *Semantic) Equal(v Semantic) bool {
	// 1.2.3 == 1.2.x
	// 1.2.3 == 1.x.x
	// 1.2.3 == x.x.x
	// x.2.3 ; x.2.x ; x.x.3 ; 1.x.3 - invalid versions
	if s.majorExplicit && v.majorExplicit && s.major != v.major {
		return false
	}
	if s.minorExplicit && v.minorExplicit && (s.major != v.major || s.minor != v.minor) {
		return false
	}
	if s.patchExplicit && v.patchExplicit && (s.major != v.major || s.minor != v.minor || s.patch != v.patch) {
		return false
	}
	return true
}
