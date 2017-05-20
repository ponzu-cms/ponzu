package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewCheckNmkAbs(t *testing.T) {
	savedGOPATH := os.Getenv("GOPATH")
	defer os.Setenv("GOPATH", savedGOPATH)
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Could not determine current working directory: %s", err)
	}

	isNil := func(e error) bool { return e == nil }

	testTable := []struct {
		base, wd, a,
		wantP string
		wantE func(e error) bool
	}{{
		base:  filepath.Join(pwd, "test-fixtures", "new"),
		wd:    filepath.Join("src", "existing"),
		a:     ".",
		wantP: filepath.Join(pwd, "test-fixtures", "new", "src", "existing"),
		wantE: os.IsExist,
	}, {
		base:  filepath.Join(pwd, "test-fixtures", "new"),
		wd:    filepath.Join(""),
		a:     "non-existing",
		wantP: filepath.Join(pwd, "test-fixtures", "new", "src", "non-existing"),
		wantE: isNil,
	}}

	for _, test := range testTable {
		os.Setenv("GOPATH", test.base)
		err = os.Chdir(filepath.Join(test.base, test.wd))
		if err != nil {
			t.Fatalf("could not setup base: %s", err)
		}
		got, gotE := checkNmkAbs(test.a)
		if got != test.wantP {
			t.Errorf("got '%s', want: '%s'", got, test.wantP)
		}
		if !test.wantE(gotE) {
			t.Errorf("got error '%s'", gotE)
		}
	}
}
