package maven

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewVersion(t *testing.T) {
	tests := []struct {
		vstr     string
		expected string
	}{
		{"1.0.0", "1"},
		{"1.ga", "1"},
		{"1.final", "1"},
		{"1.0", "1"},
		{"1.", "1"},
		{"1-", "1"},
		{"1.0.0-foo.0.0", "1-foo"},
		{"1.0.0-0.0.0", "1"},
	}
	for _, test := range tests {
		v, err := NewVersion(test.vstr)
		if err != nil {
			t.Errorf("could not create new maven version: %v", err)
		}
		assert.Equal(t, test.expected, v.String())
	}
}

func TestPadForComparison(t *testing.T) {
	tests := []struct {
		vstr string
		wstr string
	}{
		{"1.1.1", "1"},
		{"1.0", "1.1.2"},
		{"1.0", "1.1-2"},
		{"1.1.1", "1.1.2"},
		{"1-1.1", "1.1.2"},
	}
	for _, test := range tests {
		v, _ := NewVersion(test.vstr)
		w, _ := NewVersion(test.wstr)
		vp, wp := padForComparison(v, w)
		smaller := &vp
		bigger := &wp
		if len(w.fields) < len(v.fields) {
			smaller = &wp
			bigger = &vp
		}
		assert.Equal(t, len(vp.fields), len(wp.fields))
		for c := len(smaller.fields); c < len(bigger.fields); c++ {
			assert.Equal(t, bigger.fields[c].dot, smaller.fields[c].dot)
			if smaller.fields[c].dot {
				assert.Equal(t, "0", smaller.fields[c].value)
			} else {
				assert.Equal(t, "", smaller.fields[c].value)
			}
		}
	}
}

func TestVersions_Less(t *testing.T) {
	tests := []struct {
		lesser  string
		greater string
	}{
		{"1", "2"},
		{"1.5", "2"},
		{"1", "2.5"},
		{"1.0", "1.1"},
		{"1.1", "1.2"},
		{"1.0.0", "1.1"},
		{"1.1", "1.2.0"},

		{"1.1.2.alpha1", "1.1.2"},
		{"1.1.2.alpha1", "1.1.2.beta1"},
		{"1.1.2.beta1", "1.2"},

		{"1.0-alpha-1", "1.0"},
		{"1.0-alpha-1", "1.0-alpha-2"},
		{"1.0-alpha-2", "1.0-alpha-15"},
		{"1.0-alpha-1", "1.0-beta-1"},

		{"1.0-beta-1", "1.0-SNAPSHOT"},
		{"1.0-SNAPSHOT", "1.0"},
		{"1.0-alpha-1-SNAPSHOT", "1.0-alpha-1"},

		{"1.0", "1.0-1"},
		{"1.0-1", "1.0-2"},
		{"2.0", "2.0-1"},
		{"2.0.0", "2.0-1"},
		{"2.0-1", "2.0.1"},

		{"2.0.1-klm", "2.0.1-lmn"},
		{"2.0.1", "2.0.1-xyz"},
		{"2.0.1-xyz-1", "2.0.1-1-xyz"},

		{"2.0.1", "2.0.1-123"},
		{"2.0.1-xyz", "2.0.1-123"},

		{"1.2.3-10000000000", "1.2.3-10000000001"},
		{"1.2.3-1", "1.2.3-10000000001"},
		{"2.3.0-v200706262000", "2.3.0-v200706262130"}, // org.eclipse:emf:2.3.0-v200706262000
		// org.eclipse.wst.common_core.feature_2.0.0.v200706041905-7C78EK9E_EkMNfNOd2d8qq
		{"2.0.0.v200706041905-7C78EK9E_EkMNfNOd2d8qq", "2.0.0.v200706041906-7C78EK9E_EkMNfNOd2d8qq"},
	}
	for _, test := range tests {
		i, err := NewVersion(test.lesser)
		if err != nil {
			t.Errorf("error creating version %s: %v", test.lesser, err)
		}
		j, err := NewVersion(test.greater)
		if err != nil {
			t.Errorf("error creating version %s: %v", test.greater, err)
		}
		v := Versions([]Version{i, j})

		assert.True(t, v.Less(0, 1), "incorrect comparison %s - %s\n%+v\n%+v",
			test.lesser, test.greater, i, j)
	}
}
