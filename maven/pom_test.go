package maven

import (
	"testing"

	"github.com/jcmturner/mavendownload/metadata"
	"github.com/stretchr/testify/assert"
)

func TestPOM(t *testing.T) {
	md, err := metadata.Get("http://central.maven.org/maven2", "log4j", "log4j")
	if err != nil {
		t.Fatalf("error getting repo metadata: %v", err)
	}
	p, err := RepoPOM("http://central.maven.org/maven2", "log4j", "log4j", md.Versioning.Latest)
	if err != nil {
		t.Fatalf("error getting pom: %v", err)
	}
	assert.Equal(t, "log4j", p.GroupID, "GroupID not as expected")
	assert.Equal(t, "log4j", p.ArtifactID, "ArtifactID not as expected")
}

func TestNormaliseVersion(t *testing.T) {
	tests := []struct {
		in  string
		out string
	}{
		{"1-1.foo-bar1baz-.1", "1-1.foo-bar-1-baz-0.1"},
	}
	for _, test := range tests {
		norm := normaliseVersion(test.in)
		assert.Equal(t, test.out, norm, "normalisation of %s incorrect", test.in)
	}
}

func TestHyphenateAlphaNumeric(t *testing.T) {
	tests := []struct {
		in  string
		out string
	}{
		{"foo1bar", "foo-1-bar"},
		{"foo-1bar", "foo-1-bar"},
		{"foo-1-bar", "foo-1-bar"},
		{"foo1", "foo-1"},
		{"1bar", "1-bar"},
		{"foo-1-bar2foo", "foo-1-bar-2-foo"},
		{"foo123bar", "foo-123-bar"},
		{"foo-123bar", "foo-123-bar"},
		{"foo-123-bar", "foo-123-bar"},
		{"foo123", "foo-123"},
		{"123bar", "123-bar"},
		{"foo-bar-1baz-0", "foo-bar-1-baz-0"},
	}
	for _, test := range tests {
		norm := hyphenateAlphaNumeric(test.in)
		assert.Equal(t, test.out, norm, "hypenating of %s incorrect", test.in)
	}
}
