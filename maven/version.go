package maven

import (
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

// Versions is a sortable slice of maven versions
type Versions []Version

func (v Versions) Len() int {
	return len(v)
}
func (v Versions) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}
func (v Versions) Less(i, j int) bool {
	if v[i].major != v[j].major {
		return v[i].major < v[j].major
	}
	ip, jp := padForComparison(v[i], v[j])
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
