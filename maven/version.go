package maven

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Version struct {
	major      int
	fields     []vfield
	normalised string
}

type vfield struct {
	dot          bool
	numeric      bool
	value        string
	numericValue int
}

func NewVersion(s string) (v Version, err error) {
	s = normaliseVersion(s)
	v.normalised = s
	i := strings.IndexAny(s, "-.")
	if i == -1 {
		// there are no separators for fields
		v.major, err = strconv.Atoi(s)
		if err != nil {
			err = fmt.Errorf("invalid version string. major value is not a number: %v", err)
		}
		return
	}
	var d bool
	if string(s[i]) == "." {
		d = true
	}
	v.major, err = strconv.Atoi(s[:i])
	if err != nil {
		err = fmt.Errorf("invalid version string. major value is not a number: %v", err)
	}
	s = s[i+1:]
	for {
		i := strings.IndexAny(s, "-.")
		if i < 0 {
			var numeric bool
			n, err := strconv.Atoi(s)
			if err == nil {
				numeric = true
			}
			v.fields = trimAppend(v.fields, vfield{dot: d, numeric: numeric, value: s, numericValue: n})
			break
		}
		var numeric bool
		n, err := strconv.Atoi(s[:i])
		if err == nil {
			numeric = true
		}
		v.fields = trimAppend(v.fields, vfield{dot: d, numeric: numeric, value: s[:i], numericValue: n})
		d = false
		if string(s[i]) == "." {
			d = true
		}
		s = s[i+1:]
	}
	return
}

func normaliseVersion(v string) string {
	dots := strings.Split(v, ".")
	for i := range dots {
		//Add zero to trailing hyphens
		if strings.HasSuffix(dots[i], "-") {
			dots[i] = dots[i] + "0"
		}
		if dots[i] == "" {
			dots[i] = "0"
		}
		//Ensure there are hypens between letters and numbers
		dots[i] = hyphenateAlphaNumeric(dots[i])
	}
	return strings.Join(dots, ".")
}

func hyphenateAlphaNumeric(s string) string {
	re := regexp.MustCompile("[a-zA-Z][0-9]")
	ns := re.FindAllStringIndex(s, -1)
	if len(ns) != 0 {
		var p []string
		var l int
		for _, f := range ns {
			p = append(p, s[l:f[0]+1])
			l = f[0] + 1
		}
		p = append(p, s[l:])
		s = strings.Join(p, "-")
	}
	re = regexp.MustCompile("[0-9][a-zA-Z]")
	ns = re.FindAllStringIndex(s, -1)
	if len(ns) == 0 {
		return s
	}
	var p []string
	var l int
	for _, f := range ns {
		p = append(p, s[l:f[0]+1])
		l = f[0] + 1
	}
	p = append(p, s[l:])
	return strings.Join(p, "-")
}

func trimAppend(f []vfield, v ...vfield) []vfield {
	for _, vf := range v {
		vf.value = trimNullSuffix(vf.value)
		if vf.value != "" {
			f = append(f, vf)
		}
	}
	return f
}

func trimNullSuffix(s string) string {
	//trailing "null" values (0, "", "final", "ga") are trimmed
	s = strings.TrimSuffix(s, "0")
	s = strings.TrimSuffix(s, "final")
	s = strings.TrimSuffix(s, "ga")
	s = strings.TrimSuffix(s, "-")
	return strings.TrimSuffix(s, ".")
}

// String returns a normalised version string
func (v *Version) String() string {
	s := strconv.Itoa(v.major)
	for _, f := range v.fields {
		sep := "-"
		if f.dot {
			sep = "."
		}
		s = s + sep + f.value
	}
	return s
}

// Less indicates if the Version v is less than the Version w
func (v Version) Less(w Version) bool {
	if v.major != w.major {
		return v.major < w.major
	}
	ip, jp := padForComparison(v, w)
	for fi := range ip.fields {
		if ip.fields[fi].dot == jp.fields[fi].dot {
			// older -> newer
			// "alpha" = "a" < "beta" = "b" < "milestone" = "m" < "rc" = "cr" < "snapshot" < "" = "final" = "ga" < "sp" < [A-Z] < [1-9]
			if ip.fields[fi].numeric && jp.fields[fi].numeric {
				if ip.fields[fi].numericValue == jp.fields[fi].numericValue {
					continue
				}
				return ip.fields[fi].numericValue < jp.fields[fi].numericValue
			}
			if ip.fields[fi].numeric && !jp.fields[fi].numeric {
				return false
			}
			if !ip.fields[fi].numeric && jp.fields[fi].numeric {
				return true
			}
			// both are qualifiers not numbers
			t := map[string]string{
				"alpha":     "1",
				"a":         "1",
				"beta":      "2",
				"b":         "2",
				"milestone": "3",
				"m":         "3",
				"rc":        "4",
				"cr":        "4",
				"snapshot":  "5",
				"":          "6",
				"final":     "6",
				"ga":        "6",
				"sp":        "7",
			}
			ip.fields[fi].value = strings.ToLower(ip.fields[fi].value)
			ip.fields[fi].value = strings.ToLower(ip.fields[fi].value)
			if r, ok := t[ip.fields[fi].value]; ok {
				ip.fields[fi].value = r
			}
			if r, ok := t[jp.fields[fi].value]; ok {
				jp.fields[fi].value = r
			}
			x := strings.Compare(ip.fields[fi].value, jp.fields[fi].value)
			if x == 0 {
				continue
			}
			if x == -1 {
				return true
			}
			return false
		} else {
			// ".qualifier" < "-qualifier" < "-number" < ".number"
			if !ip.fields[fi].numeric && jp.fields[fi].numeric {
				// qualifier always less than qualifier
				return true
			}
			if ip.fields[fi].numeric && jp.fields[fi].numeric {
				// both are numeric: - < .
				return !ip.fields[fi].dot
			} else {
				// both are qualifier: . < -
				return ip.fields[fi].dot
			}
		}
	}
	return false
}

func (v *Version) Equal(w Version) bool {
	if v.major != w.major {
		return false
	}
	vp, wp := padForComparison(*v, w)
	for i := range vp.fields {
		if vp.fields[i].dot != wp.fields[i].dot {
			return false
		}
		if vp.fields[i].numeric != wp.fields[i].numeric {
			return false
		}
		if vp.fields[i].numeric {
			if vp.fields[i].numericValue != wp.fields[i].numericValue {
				return false
			}
		} else {
			vstr := strings.ToLower(vp.fields[i].value)
			wstr := strings.ToLower(wp.fields[i].value)
			t := map[string]string{
				"alpha":     "a",
				"beta":      "b",
				"milestone": "m",
				"rc":        "cr",
				"final":     "",
				"ga":        "",
			}
			if s, ok := t[vstr]; ok {
				vstr = s
			}
			if s, ok := t[wstr]; ok {
				wstr = s
			}
			if vstr != wstr {
				return false
			}
		}
	}
	return true
}

// padForComparison stretches the shorter version by padding with enough "null" values with matching prefix to have the
// same length as the longer one. Padded "null" values depend on the prefix of the other version: 0 for '.', "" for '-'.
func padForComparison(v, w Version) (Version, Version) {
	if len(w.fields) == len(v.fields) {
		return v, w
	}
	v, _ = NewVersion(v.normalised)
	w, _ = NewVersion(w.normalised)
	if len(w.fields) > len(v.fields) {
		for i := len(v.fields); i < len(w.fields); i++ {
			val := ""
			var numeric bool
			if w.fields[i].dot {
				val = "0"
				numeric = true
			}
			v.fields = append(v.fields, vfield{dot: w.fields[i].dot, numeric: numeric, value: val})
		}
		return v, w
	}
	for i := len(w.fields); i < len(v.fields); i++ {
		val := ""
		var numeric bool
		if v.fields[i].dot {
			val = "0"
			numeric = true
		}
		w.fields = append(w.fields, vfield{dot: v.fields[i].dot, numeric: numeric, value: val})
	}
	return v, w
}

// Versions is a sortable slice of maven versions.
type Versions []Version

// Len returns the length of the Versions slice. Required to satisfy the sort interface.
func (v Versions) Len() int {
	return len(v)
}

// Swap elements in the Versions slice. Required to satisfy the sort interface.
func (v Versions) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

// Less indicates if the Version at position i is less (older) than the element at position j.
// Required to satisfy the sort interface.
func (v Versions) Less(i, j int) bool {
	return v[i].Less(v[j])
}

//Version requirements have the following syntax:
//
//1.0: "Soft" requirement on 1.0 (just a recommendation, if it matches all other ranges for the dependency)
//[1.0]: "Hard" requirement on 1.0
//(,1.0]: x <= 1.0
//[1.2,1.3]: 1.2 <= x <= 1.3
//[1.0,2.0): 1.0 <= x < 2.0
//[1.5,): x >= 1.5
//(,1.0],[1.2,): x <= 1.0 or x >= 1.2; multiple sets are comma-separated
//(,1.1),(1.1,): this excludes 1.1 (for example if it is known not to work in combination with this library)

// If [n then n <= x
// if n] then x <= n

// if (n then n < x
// if n) then x < n

// Invalid combinations:
// curved brackets with only one number
// curved bracket with same number twice
// first number greater than 2nd
// numbers not in brackets with comma

type condition struct {
	openBracketSquare  bool
	closeBracketSquare bool
	lower              Version
	upper              Version
	undefLower         bool // lower is undefined
	undefUpper         bool // upper is undefined
}

func parseRequirement(s string) (conds []condition, err error) {
	i := strings.IndexAny(s, "[]()")
	if i == -1 {
		if strings.Contains(s, ",") {
			err = errors.New("invalid version requirements")
			return
		}
		// there are no brackets
		var c condition
		c.lower, err = NewVersion(s)
		if err != nil {
			err = fmt.Errorf("could not parse version condition %s: %v", s, err)
			return
		}
		c.upper = c.lower
		conds = append(conds, c)
		return
	}
	idx := regexp.MustCompile(`[\)\]\(\[]`).FindAllStringIndex(s, -1)
	if idx[0][0] != 0 || len(s) > idx[len(idx)-1][1] {
		// there are characters outside of brackets. eg [1.0,1.2),1.3 or 0.9,[1.0,1.2)
		err = errors.New("invalid version requirements")
		return
	}
	for i := range idx {
		if i%2 == 0 {
			continue
		}
		f := idx[i-1][1]
		l := idx[i][0]
		//if l < len(s)-1 {
		//	l++
		//}
		c := condition{
			openBracketSquare:  string(s[f-1]) == "[",
			closeBracketSquare: string(s[l]) == "]",
		}
		v := strings.Split(s[f:l], ",")
		if len(v) == 1 {
			c.lower, err = NewVersion(v[0])
			if err != nil {
				err = fmt.Errorf("could not parse lower condition version %s: %v", v[0], err)
				return
			}
			c.upper = c.lower
		} else {
			if v[0] != "" {
				c.lower, err = NewVersion(v[0])
				if err != nil {
					err = fmt.Errorf("could not parse lower condition version %s: %v", v[0], err)
					return
				}
			} else {
				c.undefLower = true
			}
			if v[1] != "" {
				c.upper, err = NewVersion(v[1])
				if err != nil {
					err = fmt.Errorf("could not parse upper condition version %s: %v", v[1], err)
					return
				}
			} else {
				c.undefUpper = true
			}
		}
		if !c.undefUpper && !c.undefLower && c.upper.Less(c.lower) {
			err = errors.New("invalid version requirements")
			return
		}
		if (!c.closeBracketSquare || !c.openBracketSquare) && c.upper.Equal(c.lower) {
			// curved brackets with only one number
			// curved bracket with same number twice
			err = errors.New("invalid version requirements")
			return
		}
		conds = append(conds, c)
	}
	return
}

func (v Version) Satisfies(r string) bool {
	conds, err := parseRequirement(r)
	if err != nil {
		return false
	}
	for _, c := range conds {
		var ls, us bool
		if c.undefLower {
			ls = true
		} else {
			if c.lower.Less(v) {
				ls = true
			}
			if c.openBracketSquare && v.Equal(c.lower) {
				ls = true
			}
		}
		if c.undefUpper {
			us = true
		} else {
			if v.Less(c.upper) {
				us = true
			}
			if c.closeBracketSquare && v.Equal(c.upper) {
				us = true
			}
		}
		if ls && us {
			return true
		}
	}
	return false
}
