package config

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bmatcuk/doublestar/v3"
	"sigs.k8s.io/yaml"
)

type Parser struct {
	Config Config
	parsed map[string]struct{}
}

func NewParser() *Parser {
	return &Parser{
		parsed: map[string]struct{}{},
	}
}

func (p *Parser) Parse(reader io.Reader, path string) error {
	var c Config
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read bytes: %v", err)
	}

	if err := yaml.UnmarshalStrict(content, &c); err != nil {
		return fmt.Errorf("failed to parse yaml: %v", err)
	}

	// channels, err := mergeChannels(p.Config.Channels, c.Channels)
	// if err != nil {
	// 	return fmt.Errorf("couldn't merge channels: %v", err)
	// }
	p.Config.Channels = c.Channels

	return nil
}

func (p *Parser) ParseFile(path, basedir string) error {
	if _, ok := p.parsed[path]; ok {
		return nil
	}
	p.parsed[path] = struct{}{}
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open %s: %v", path, err)
	}
	defer f.Close()
	if !strings.HasPrefix(path, basedir) {
		return fmt.Errorf("%q is not a prefix of %q", basedir, path)
	}
	if err := p.Parse(f, path[len(basedir):]); err != nil {
		return fmt.Errorf("failed to parse %s: %v", path, err)
	}
	return nil
}

func (p *Parser) ParseDir(path string) error {
	matches, err := doublestar.Glob(filepath.Join(path, "**/*.yaml"))
	if err != nil {
		return fmt.Errorf("failed to find config files: %v", err)
	}

	for _, f := range matches {
		if err := p.ParseFile(f, path); err != nil {
			return err
		}
	}
	return nil
}

func ParseFile(path string) (Config, error) {
	p := NewParser()
	if err := p.ParseFile(path, path); err != nil {
		return Config{}, err
	}
	return p.Config, nil
}

func ParseDir(path string) (Config, error) {
	p := NewParser()
	if err := p.ParseDir(path); err != nil {
		return Config{}, err
	}
	return p.Config, nil
}

func matchesRegexList(s string, tests []*regexp.Regexp) bool {
	for _, r := range tests {
		if r.MatchString(s) {
			return true
		}
	}
	return false
}

// func mergeChannels(a []Channel, b []Channel) ([]Channel, error) {
// 	names := map[string]struct{}{}
// 	ids := map[string]struct{}{}
// 	for _, v := range a {
// 		names[v.Name] = struct{}{}
// 		if v. != "" {
// 			ids[v.ID] = struct{}{}
// 		}
// 	}
// 	for _, v := range b {
// 		if v.Name == "" {
// 			return nil, fmt.Errorf("channels must have names")
// 		}
// 		if _, ok := names[v.Name]; ok {
// 			return nil, fmt.Errorf("cannot overwrite channel definitions (duplicate channel name %s)", v.Name)
// 		}
// 		if _, ok := ids[v.ID]; ok {
// 			return nil, fmt.Errorf("cannot overwrite channel definitions (duplicate channel ID %s)", v.Name)
// 		}
// 	}

// 	return append(a, b...), nil
// }