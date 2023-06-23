package rss

import (
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestParseAtomTitle(t *testing.T) {
	tests := map[string]string{
		"atom_1.0":           "Titel des Weblogs",
		"atom_1.0_enclosure": "Titel des Weblogs",
		"atom_1.0-1":         "Golem.de",
	}

	for test, want := range tests {
		name := filepath.Join("testdata", test)
		data, err := ioutil.ReadFile(name)
		if err != nil {
			t.Fatalf("Reading %s: %v", name, err)
		}

		feed, err := Parse(data)
		if err != nil {
			t.Fatalf("Parsing %s: %v", name, err)
		}

		if feed.Title != want {
			t.Errorf("%s: got %q, want %q", name, feed.Title, want)
		}
	}
}
