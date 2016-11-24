package main

import "testing"

func TestParseType(t *testing.T) {
	// blog title:string Author:string PostCategory:string content:string some_thing:int
	args := []string{
		"blog", "title:string", "Author:string",
		"PostCategory:string", "content:string",
		"some_thing:int", "Some_otherThing:float64",
	}

	gt, err := parseType(args)
	if err != nil {
		t.Errorf("Failed: %s", err.Error())
	}

	if gt.Name != "Blog" {
		t.Errorf("Expected %s, got: %s", "Blog", gt.Name)
	}
}

func TestFieldJSONName(t *testing.T) {
	cases := map[string]string{
		"_T":                   "t",
		"T":                    "t",
		"_tT_":                 "t_t_",
		"TestCapsNoSym":        "test_caps_no_sym",
		"test_Some_caps_Sym":   "test_some_caps_sym",
		"testnocaps":           "testnocaps",
		"_Test_Caps_Sym_odd":   "test_caps_sym_odd",
		"test-hyphen":          "test-hyphen",
		"Test-hyphen-Caps":     "test-hyphen-caps",
		"Test-Hyphen_Sym-Caps": "test-hyphen_sym-caps",
	}

	for input, expected := range cases {
		output := fieldJSONName(input)
		if output != expected {
			t.Errorf("Expected: %s, got: %s", expected, output)
		}
	}
}

func TestFieldName(t *testing.T) {
	cases := map[string]string{
		"_T":                   "T",
		"T":                    "T",
		"_tT_":                 "TT",
		"TestCapsNoSym":        "TestCapsNoSym",
		"test_Some_caps_Sym":   "TestSomeCapsSym",
		"testnocaps":           "Testnocaps",
		"_Test_Caps_Sym_odd":   "TestCapsSymOdd",
		"test-hyphen":          "TestHyphen",
		"Test-hyphen-Caps":     "TestHyphenCaps",
		"Test-Hyphen_Sym-Caps": "TestHyphenSymCaps",
	}

	for input, expected := range cases {
		output := fieldName(input)
		if output != expected {
			t.Errorf("Expected: %s, got: %s", expected, output)
		}
	}
}
