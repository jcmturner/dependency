package maven

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/jcmturner/dependency/components"
	"github.com/jcmturner/mavendownload/metadata"
)

const (
	pomFile = "pom.xml"
)

type POM struct {
	Project
}

type Project struct {
	ModelVersion string       `xml:"modelVersion"`
	GroupID      string       `xml:"groupId"`
	ArtifactID   string       `xml:"artifactId"`
	Version      string       `xml:"version"`
	Packaging    string       `xml:"packaging"`
	Description  string       `xml:"description"`
	URL          string       `xml:"url"`
	Name         string       `xml:"name"`
	Licenses     []License    `xml:"licenses>license"`
	Dependencies []Dependency `xml:"dependencies>dependency"`
}

type License struct {
	Name         string `xml:"name"`
	URL          string `xml:"url"`
	Distribution string `xml:"distribution"`
}

type Dependency struct {
	GroupID    string `xml:"groupId"`
	ArtifactID string `xml:"artifactId"`
	Version    string `xml:"version"`
	Type       string `xml:"type"`
	Scope      string `xml:"scope"`
	Optional   bool   `xml:"optional"`
}

func RepoPOM(repo, groupID, artifactID, version string) (p Project, err error) {
	url := fmt.Sprintf("%s/%s/%s/%s/%s-%s.pom", strings.TrimRight(repo, "/"), groupID, artifactID, version, artifactID, version)

	// Get the POM file
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		err = fmt.Errorf("error forming request of %s: %v", url, err)
		return
	}
	cl := http.DefaultClient
	resp, err := cl.Do(req)
	if err != nil {
		err = fmt.Errorf("error getting %s: %v", url, err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("http response %d downloading POM file", resp.StatusCode)
		return
	}
	mb, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("error reading body from %s: %v", url, err)
		return
	}
	defer resp.Body.Close()

	// Get the POM file SHA1
	psha1, err := metadata.SHA1(url)
	if err != nil {
		err = fmt.Errorf("error getting POM SHA1: %v", err)
		return
	}

	// Check the md5 of the metadata
	hash := sha1.New()
	hash.Write(mb)
	h := hex.EncodeToString(hash.Sum(nil))
	if h != psha1 {
		err = fmt.Errorf("checksum of POM does not match. expected: %s got: %s", psha1, h)
		return
	}

	// Marshal bytes into MetaData type
	rdr := bytes.NewReader(mb)
	decoder := xml.NewDecoder(rdr)
	err = decoder.Decode(&p)
	if err != nil {
		err = fmt.Errorf("error decoding POM from %s: %v", url, err)
		return
	}
	return
}

func LoadPOM(path string) (Project, error) {
	var p Project
	fh, err := os.Open(path)
	if err != nil {
		return p, fmt.Errorf("could not open POM file at %s: %v", path, err)
	}
	defer fh.Close()
	//b, err := ioutil.ReadAll(fh)
	//r := bytes.NewReader(b)
	decoder := xml.NewDecoder(fh)
	err = decoder.Decode(&p)
	if err != nil {
		return p, fmt.Errorf("could not decode POM file at %s: %v", path, err)
	}
	return p, nil
}

func (p *POM) Find(srcRoot string) (c []components.Component, err error) {
	var files []string
	err = filepath.Walk(srcRoot,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && info.Name() == pomFile {
				files = append(files, path)
			}
			return nil
		})
	if err != nil {
		err = fmt.Errorf("error looking for POM files: %v", err)
		return
	}
	for _, f := range files {
		p, e := LoadPOM(f)
		if e != nil {
			return c, e
		}
		for _, d := range p.Dependencies {
			if d.Scope == "test" {
				continue
			}
			c = append(c, components.Component{
				Class:   components.ClassLib,
				Type:    components.TypeJava,
				ID:      fmt.Sprintf("%s.%s", d.GroupID, d.ArtifactID),
				Version: d.Version,
			})
		}
	}
	return
}

func (p *POM) Type() components.Type {
	return components.TypeJava
}

func (p *POM) Class() components.Class {
	return components.ClassLib
}

//func (d *Dependency) Satisfied(v string) bool {
/*Dependencies' version element define version requirements, used to compute effective dependency version. Version requirements have the following syntax:
1.0: "Soft" requirement on 1.0 (just a recommendation, if it matches all other ranges for the dependency)
[1.0]: "Hard" requirement on 1.0
(,1.0]: x <= 1.0
[1.2,1.3]: 1.2 <= x <= 1.3
[1.0,2.0): 1.0 <= x < 2.0
[1.5,): x >= 1.5
(,1.0],[1.2,): x <= 1.0 or x >= 1.2; multiple sets are comma-separated
(,1.1),(1.1,): this excludes 1.1 (for example if it is known not to work in combination with this library)*/

// If [n then n <= x
// if n] then x <= n

// if (n then n < x
// if n) then x < n

//vs := version.ParseSemantic(v)

//}
