package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewName2Path(t *testing.T) {
	savedGOPATH := os.Getenv("GOPATH")
	defer os.Setenv("GOPATH", savedGOPATH)
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Could not determine current working directory: %s", err)
	}

	isNil := func(e error) bool { return e == nil }
	isNonNil := func(e error) bool { return e != nil }

	baseDir := filepath.Join(pwd, "test-fixtures", "new")

	testTable := []struct {
		gopath, wd, a,
		wantP string
		wantE func(e error) bool
	}{{
		gopath: baseDir,
		wd:     filepath.Join("src", "existing"),
		a:      ".",
		wantP:  filepath.Join(pwd, "test-fixtures", "new", "src", "existing"),
		wantE:  os.IsExist,
	}, {
		gopath: baseDir,
		wd:     filepath.Join(""),
		a:      "non-existing",
		wantP:  filepath.Join(pwd, "test-fixtures", "new", "src", "non-existing"),
		wantE:  isNil,
	}, {
		gopath: baseDir,
		wd:     filepath.Join(""),
		a:      ".",
		wantP:  "",
		wantE:  isNonNil,
	}, {
		gopath: baseDir,
		wd:     "..",
		a:      ".",
		wantP:  "",
		wantE:  isNonNil,
	}}

	for _, test := range testTable {
		os.Setenv("GOPATH", test.gopath)
		err = os.Chdir(filepath.Join(test.gopath, test.wd))
		if err != nil {
			t.Fatalf("could not setup base: %s", err)
		}
		got, gotE := name2path(test.a)
		if got != test.wantP {
			t.Errorf("got '%s', want: '%s'", got, test.wantP)
		}
		if !test.wantE(gotE) {
			t.Errorf("got error '%s'", gotE)
		}
	}
}
