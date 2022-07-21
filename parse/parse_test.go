// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parse

import (
	"flag"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

var debug = flag.Bool("debug", false, "show the errors produced by the main tests")

type numberTest struct {
	text      string
	isInt     bool
	isUint    bool
	isFloat   bool
	isComplex bool
	int64
	uint64
	float64
	complex128
}

var numberTests = []numberTest{
	// basics
	{"0", true, true, true, false, 0, 0, 0, 0},
	{"-0", true, true, true, false, 0, 0, 0, 0}, // check that -0 is a uint.
	{"73", true, true, true, false, 73, 73, 73, 0},
	{"7_3", true, true, true, false, 73, 73, 73, 0},
	{"0b10_010_01", true, true, true, false, 73, 73, 73, 0},
	{"0B10_010_01", true, true, true, false, 73, 73, 73, 0},
	{"073", true, true, true, false, 073, 073, 073, 0},
	{"0o73", true, true, true, false, 073, 073, 073, 0},
	{"0O73", true, true, true, false, 073, 073, 073, 0},
	{"0x73", true, true, true, false, 0x73, 0x73, 0x73, 0},
	{"0X73", true, true, true, false, 0x73, 0x73, 0x73, 0},
	{"0x7_3", true, true, true, false, 0x73, 0x73, 0x73, 0},
	{"-73", true, false, true, false, -73, 0, -73, 0},
	{"+73", true, false, true, false, 73, 0, 73, 0},
	{"100", true, true, true, false, 100, 100, 100, 0},
	{"1e9", true, true, true, false, 1e9, 1e9, 1e9, 0},
	{"-1e9", true, false, true, false, -1e9, 0, -1e9, 0},
	{"-1.2", false, false, true, false, 0, 0, -1.2, 0},
	{"1e19", false, true, true, false, 0, 1e19, 1e19, 0},
	{"1e1_9", false, true, true, false, 0, 1e19, 1e19, 0},
	{"1E19", false, true, true, false, 0, 1e19, 1e19, 0},
	{"-1e19", false, false, true, false, 0, 0, -1e19, 0},
	{"0x_1p4", true, true, true, false, 16, 16, 16, 0},
	{"0X_1P4", true, true, true, false, 16, 16, 16, 0},
	{"0x_1p-4", false, false, true, false, 0, 0, 1 / 16., 0},
	{"4i", false, false, false, true, 0, 0, 0, 4i},
	{"-1.2+4.2i", false, false, false, true, 0, 0, 0, -1.2 + 4.2i},
	{"073i", false, false, false, true, 0, 0, 0, 73i}, // not octal!
	// complex with 0 imaginary are float (and maybe integer)
	{"0i", true, true, true, true, 0, 0, 0, 0},
	{"-1.2+0i", false, false, true, true, 0, 0, -1.2, -1.2},
	{"-12+0i", true, false, true, true, -12, 0, -12, -12},
	{"13+0i", true, true, true, true, 13, 13, 13, 13},
	// funny bases
	{"0123", true, true, true, false, 0123, 0123, 0123, 0},
	{"-0x0", true, true, true, false, 0, 0, 0, 0},
	{"0xdeadbeef", true, true, true, false, 0xdeadbeef, 0xdeadbeef, 0xdeadbeef, 0},
	// character constants
	{`'a'`, true, true, true, false, 'a', 'a', 'a', 0},
	{`'\n'`, true, true, true, false, '\n', '\n', '\n', 0},
	{`'\\'`, true, true, true, false, '\\', '\\', '\\', 0},
	{`'\''`, true, true, true, false, '\'', '\'', '\'', 0},
	{`'\xFF'`, true, true, true, false, 0xFF, 0xFF, 0xFF, 0},
	{`'パ'`, true, true, true, false, 0x30d1, 0x30d1, 0x30d1, 0},
	{`'\u30d1'`, true, true, true, false, 0x30d1, 0x30d1, 0x30d1, 0},
	{`'\U000030d1'`, true, true, true, false, 0x30d1, 0x30d1, 0x30d1, 0},
	// some broken syntax
	{text: "+-2"},
	{text: "0x123."},
	{text: "1e."},
	{text: "0xi."},
	{text: "1+2."},
	{text: "'x"},
	{text: "'xx'"},
	{text: "'433937734937734969526500969526500'"}, // Integer too large - issue 10634.
	// Issue 8622 - 0xe parsed as floating point. Very embarrassing.
	{"0xef", true, true, true, false, 0xef, 0xef, 0xef, 0},
}

func TestNumberParse(t *testing.T) {
	for _, test := range numberTests {
		// If fmt.Sscan thinks it's complex, it's complex. We can't trust the output
		// because imaginary comes out as a number.
		var c complex128
		typ := itemNumber
		var tree *Tree
		if test.text[0] == '\'' {
			typ = itemCharConstant
		} else {
			_, err := fmt.Sscan(test.text, &c)
			if err == nil {
				typ = itemComplex
			}
		}
		n, err := tree.newNumber(0, test.text, typ)
		ok := test.isInt || test.isUint || test.isFloat || test.isComplex
		if ok && err != nil {
			t.Errorf("unexpected error for %q: %s", test.text, err)
			continue
		}
		if !ok && err == nil {
			t.Errorf("expected error for %q", test.text)
			continue
		}
		if !ok {
			if *debug {
				fmt.Printf("%s\n\t%s\n", test.text, err)
			}
			continue
		}
		if n.IsComplex != test.isComplex {
			t.Errorf("complex incorrect for %q; should be %t", test.text, test.isComplex)
		}
		if test.isInt {
			if !n.IsInt {
				t.Errorf("expected integer for %q", test.text)
			}
			if n.Int64 != test.int64 {
				t.Errorf("int64 for %q should be %d Is %d", test.text, test.int64, n.Int64)
			}
		} else if n.IsInt {
			t.Errorf("did not expect integer for %q", test.text)
		}
		if test.isUint {
			if !n.IsUint {
				t.Errorf("expected unsigned integer for %q", test.text)
			}
			if n.Uint64 != test.uint64 {
				t.Errorf("uint64 for %q should be %d Is %d", test.text, test.uint64, n.Uint64)
			}
		} else if n.IsUint {
			t.Errorf("did not expect unsigned integer for %q", test.text)
		}
		if test.isFloat {
			if !n.IsFloat {
				t.Errorf("expected float for %q", test.text)
			}
			if n.Float64 != test.float64 {
				t.Errorf("float64 for %q should be %g Is %g", test.text, test.float64, n.Float64)
			}
		} else if n.IsFloat {
			t.Errorf("did not expect float for %q", test.text)
		}
		if test.isComplex {
			if !n.IsComplex {
				t.Errorf("expected complex for %q", test.text)
			}
			if n.Complex128 != test.complex128 {
				t.Errorf("complex128 for %q should be %g Is %g", test.text, test.complex128, n.Complex128)
			}
		} else if n.IsComplex {
			t.Errorf("did not expect complex for %q", test.text)
		}
	}
}

type parseTest struct {
	name   string
	input  string
	ok     bool
	result string // what the user would see in an error message.
}

const (
	noError  = true
	hasError = false
)

var parseTests = []parseTest{
	{"empty", "", noError,
		``},
	{"comment", "# foo\n\n", noError,
		``},
	{"spaces", " \t\n", noError,
		``},
	{"field", ".X", noError,
		`{{.X}}`},
	{"simple command", "printf", noError,
		`{{printf}}`},
	{"$ invocation", "$", noError,
		"{{$}}"},
	{"variable invocation", "with $x := 3\n$x 23\nend", noError,
		"{{with $x := 3}}{{$x 23}}{{end}}"},
	{"variable with fields", "$.I", noError,
		"{{$.I}}"},
	{"multi-word command", "printf `%d` 23", noError,
		"{{printf `%d` 23}}"},
	{"pipeline", ".X|.Y", noError,
		`{{.X | .Y}}`},
	{"pipeline with decl", "$x := .X|.Y", noError,
		`{{$x := .X | .Y}}`},
	{"nested pipeline", ".X (.Y .Z) (.A | .B .C) (.E)", noError,
		`{{.X (.Y .Z) (.A | .B .C) (.E)}}`},
	{"field applied to parentheses", "(.Y .Z).Field", noError,
		`{{(.Y .Z).Field}}`},
	{"simple if", "if .X\nprintf\nend", noError,
		"{{if .X}}{{printf}}{{end}}"},
	{"if with else", "if .X\ntrue\nelse\nfalse\nend", noError,
		`{{if .X}}{{true}}{{else}}{{false}}{{end}}`},
	{"if with else if", "if .X\ntrue\nelse if .Y\nfalse\nend", noError,
		`{{if .X}}{{true}}{{else}}{{if .Y}}{{false}}{{end}}{{end}}`},
	// 	{"if else chain", "+{{if .X}}X{{else if .Y}}Y{{else if .Z}}Z{{end}}+", noError,
	// 		`"+"{{if .X}}"X"{{else\nif .Y}}"Y"{{else\nif .Z}}"Z"{{end\nend\nend"+"`},
	// 	{"simple range", "range .X}}hello{{end}}", noError,
	// 		`{{range .X}}"hello"{{end}}`},
	// 	{"chained field range", "range .X.Y.Z}}hello{{end}}", noError,
	// 		`{{range .X.Y.Z}}"hello"{{end}}`},
	// 	{"nested range", "range .X}}hello{{range .Y}}goodbye{{end\nend", noError,
	// 		`{{range .X}}"hello"range .Y}}"goodbye"{{end\nend}}`},
	// 	{"range with else", "range .X}}true{{else}}false{{end}}", noError,
	// 		`{{range .X}}"true"{{else}}"false"{{end}}`},
	// 	{"range over pipeline", "range .X|.M}}true{{else}}false{{end}}", noError,
	// 		`{{range .X | .M}}"true"{{else}}"false"{{end}}`},
	// 	{"range []int", "range .SI\n.\nend", noError,
	// 		`{{range .SI\n.\nend}}`},
	// 	{"range 1 var", "range $x := .SI\n.\nend", noError,
	// 		`{{range $x := .SI\n.\nend}}`},
	// 	{"range 2 vars", "range $x, $y := .SI\n.\nend", noError,
	// 		`{{range $x, $y := .SI\n.\nend}}`},
	// 	{"range with break", "range .SI\n.\nbreak\nend", noError,
	// 		`{{range .SI\n.\nbreak\nend}}`},
	// 	{"range with continue", "range .SI\n.\ncontinue\nend", noError,
	// 		`{{range .SI\n.\ncontinue\nend}}`},
	// 	{"constants", "range .SI 1 -3.2i true false 'a' nil\nend", noError,
	// 		`{{range .SI 1 -3.2i true false 'a' nil\nend}}`},
	// 	{"template", "{{template `x`}}", noError,
	// 		`{{template "x"}}`},
	// 	{"template with arg", "{{template `x` .Y}}", noError,
	// 		`{{template "x" .Y}}`},
	// 	{"with", "{{with .X}}hello{{end}}", noError,
	// 		`{{with .X}}"hello"{{end}}`},
	// 	{"with with else", "{{with .X}}hello{{else}}goodbye{{end}}", noError,
	// 		`{{with .X}}"hello"{{else}}"goodbye"{{end}}`},
	// 	// Trimming spaces.
	// 	{"trim left", "x \r\n\t{{- 3}}", noError, `"x"{{3}}`},
	// 	{"trim right", "{{3 -}}\n\n\ty", noError, `{{3}}"y"`},
	// 	{"trim left and right", "x \r\n\t{{- 3 -}}\n\n\ty", noError, `"x"{{3}}"y"`},
	// 	{"trim with extra spaces", "x\n{{-  3   -}}\ny", noError, `"x"{{3}}"y"`},
	// 	{"comment trim left", "x \r\n\t{{- /* hi */}}", noError, `"x"`},
	// 	{"comment trim right", "{{/* hi */ -}}\n\n\ty", noError, `"y"`},
	// 	{"comment trim left and right", "x \r\n\t{{- /* */ -}}\n\n\ty", noError, `"x""y"`},
	// 	{"block definition", `{{block "foo" .}}hello{{end}}`, noError,
	// 		`{{template "foo" .}}`},
	//
	{"newline in assignment", "$x \\\n := \\\n 1 \\\n", noError, "{{$x := 1}}"},
	// {"newline in empty action", "{{\n}}", hasError, "{{\n}}"},
	{"newline in pipeline", `
"x" \
| \
printf`, noError, `{{"x" | printf}}`},

	// Errors.
	{"unclosed action", "range", hasError, ""},
	{"unmatched end", "end", hasError, ""},
	{"unmatched else", "else", hasError, ""},
	{"unmatched else after if", "if .X\nhello\nend\nelse\n", hasError, ""},
	{"multiple else", "if .X\n1\nelse\n2\nelse\n3\nend", hasError, ""},
	{"missing end", "range .x", hasError, ""},
	{"missing end after else", "range .x\nelse", hasError, ""},
	{"undefined function", "undefined", hasError, ""},
	{"undefined variable", "$x", hasError, ""},
	{"variable undefined after end", "with $x := 4\nend\n$x", hasError, ""},
	{"variable undefined in template", "template $v", hasError, ""},
	{"declare with field", "with $x.Y := 4\nend", hasError, ""},
	{"template with field ref", "{{template .X}}", hasError, ""},
	{"template with var", "template $v", hasError, ""},
	{"invalid punctuation", "printf 3, 4", hasError, ""},
	{"multidecl outside range", "with $v, $u := 3\nend", hasError, ""},
	{"too many decls in range", "range $u, $v, $w := 3\nend", hasError, ""},
	{"dot applied to parentheses", "printf (printf .).", hasError, ""},
	{"adjacent args", "printf 3`x`", hasError, ""},
	{"adjacent args with .", "printf `x`.", hasError, ""},
	// {"extra end after if", "{{if .X}}a{{else if .Y}}b{{end\nend", hasError, ""},
	{"break outside range", "range .\nend\n break", hasError, ""},
	{"continue outside range", "range .\nend continue", hasError, ""},
	{"break in range else", "range .\nelse\nbreak\nend", hasError, ""},
	{"continue in range else", "range .\nelse\ncontinue\nend", hasError, ""},
	// Other kinds of assignments and operators aren't available yet.
	{"bug0a", "$x := 0\n$x", noError, "{{$x := 0}}{{$x}}"},
	{"bug0b", "$x += 1\n$x", hasError, ""},
	{"bug0c", "$x ! 2\n$x", hasError, ""},
	{"bug0d", "$x % 3\n$x", hasError, ""},
	// Check the parse fails for := rather than comma.
	{"bug0e", "range $x := $y := 3\nend", hasError, ""},
	// Another bug: variable read must ignore following punctuation.
	{"bug1a", "$x:=.\n$x!2", hasError, ""},                     // ! is just illegal here.
	{"bug1b", "$x:=.\n$x+2", hasError, ""},                     // $x+2 should not parse as ($x) (+2).
	{"bug1c", "$x:=.\n$x +2", noError, "{{$x := .}}{{$x +2}}"}, // It's OK with a space.
	// dot following a literal value
	{"dot after integer", "1.E", hasError, ""},
	{"dot after float", "0.1.E", hasError, ""},
	{"dot after boolean", "true.E", hasError, ""},
	{"dot after char", "'a'.any", hasError, ""},
	{"dot after string", `"hello".guys`, hasError, ""},
	{"dot after dot", "..E", hasError, ""},
	{"dot after nil", "nil.E", hasError, ""},
	// Wrong pipeline
	{"wrong pipeline dot", "12|.", hasError, ""},
	{"wrong pipeline number", ".|12|printf", hasError, ""},
	{"wrong pipeline string", ".|printf|\"error\"", hasError, ""},
	{"wrong pipeline char", "12|printf|'e'", hasError, ""},
	{"wrong pipeline boolean", ".|true", hasError, ""},
	{"wrong pipeline nil", "'c'|nil", hasError, ""},
	{"empty pipeline", `printf "%d" ( )`, hasError, ""},
	// Missing pipeline in block
	{"block definition", "block \"foo\"\nhello\nend", hasError, ""},
}

type TestTemplateFuncs map[string]any

func (tf TestTemplateFuncs) Has(name string) bool {
	return tf != nil && tf[name] != nil
}

func (tf TestTemplateFuncs) GetByName(name string) reflect.Value {
	panic("unexpected GetByName call during testing")
}

var builtins = TestTemplateFuncs{
	"printf":   fmt.Sprintf,
	"contains": strings.Contains,
}

func testParse(doCopy bool, t *testing.T) {
	textFormat = "%q"
	defer func() { textFormat = "%s" }()
	for _, test := range parseTests {
		tmpl, err := New(test.name, nil).Parse(test.input, make(map[string]*Tree), builtins)
		switch {
		case err == nil && !test.ok:
			t.Errorf("\n%q: expected error; got none", test.name)
			continue
		case err != nil && test.ok:
			t.Errorf("\n%q: unexpected error: %v", test.name, err)
			continue
		case err != nil && !test.ok:
			// expected error, got one
			if *debug {
				fmt.Printf("\n%s: %s\n\t%s\n", test.name, test.input, err)
			}
			continue
		}
		var result string
		if doCopy {
			result = tmpl.Root.Copy().String()
		} else {
			result = tmpl.Root.String()
		}
		if result != test.result {
			t.Errorf("\n%s=(%q): got\n\t%v\nexpected\n\t%v", test.name, test.input, result, test.result)
		}
	}
}

func TestParse(t *testing.T) {
	testParse(false, t)
}

// Same as TestParse, but we copy the node first
func TestParseCopy(t *testing.T) {
	testParse(true, t)
}

func TestParseWithComments(t *testing.T) {
	textFormat = "%q"
	defer func() { textFormat = "%s" }()
	tests := [...]parseTest{
		{"comment", "# foo", noError, "{{/* foo*/}}"},
		// {"comment trim left", "x \r\n\t# hi", noError, `x{{/* hi */}}`},
		// {"comment trim right", "{{/* hi */ -}}\n\n\ty", noError, `{{/* hi */}}"y"`},
		// {"comment trim left and right", "x \r\n\t{{- /* */ -}}\n\n\ty", noError, `"x"{{/* */}}"y"`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tr := New(test.name, nil)
			tr.Mode = ParseComments
			tmpl, err := tr.Parse(test.input, make(map[string]*Tree), nil)
			if err != nil {
				t.Errorf("%q: expected error; got none", test.name)
			}
			if result := tmpl.Root.String(); result != test.result {
				t.Errorf("%s=(%q): got\n\t%v\nexpected\n\t%v", test.name, test.input, result, test.result)
			}
		})
	}
}

func TestSkipFuncCheck(t *testing.T) {
	oldTextFormat := textFormat
	textFormat = "%q"
	defer func() { textFormat = oldTextFormat }()
	tr := New("skip func check", nil)
	tr.Mode = SkipFuncCheck
	tmpl, err := tr.Parse("fn 1 2", make(map[string]*Tree), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "{{fn 1 2}}"
	if result := tmpl.Root.String(); result != expected {
		t.Errorf("got\n\t%v\nexpected\n\t%v", result, expected)
	}
}

type isEmptyTest struct {
	name  string
	input string
	empty bool
}

var isEmptyTests = []isEmptyTest{
	{"empty", ``, true},
	{"nonempty", `"hello"`, false},
	{"spaces only", " \t\n \t\n", true},
	{"comment only", "# comment", true},
	{"definition", "define \"x\"\nsomething\nend", true},
	{"definitions and space", "define `x`\nsomething\nend\n\ndefine `y`\nsomething\nend\n\n", true},
	{"definitions and text", "define `x`\nsomething\nend\n\n'x'\n\ndefine `y`\nsomething\nend\n\n'y'\n", false},
	{"definition and action", "define `x`\nsomething\nend\nif 3\nfoo\nend\n", false},
}

func TestIsEmpty(t *testing.T) {
	if !IsEmptyTree(nil) {
		t.Errorf("nil tree is not empty")
	}
	for _, test := range isEmptyTests {
		tree, err := New("root", nil).Parse(test.input, make(map[string]*Tree), TestTemplateFuncs{
			"something": func() string { return "something" },
			"foo":       func() string { return "foo" },
		})
		if err != nil {
			t.Errorf("%q: unexpected error: %v", test.name, err)
			continue
		}
		if empty := IsEmptyTree(tree.Root); empty != test.empty {
			t.Errorf("%q: expected %t got %t", test.name, test.empty, empty)
		}
	}
}

func TestErrorContextWithTreeCopy(t *testing.T) {
	tree, err := New("root", nil).Parse("if true\nend", make(map[string]*Tree), nil)
	if err != nil {
		t.Fatalf("unexpected tree parse failure: %v", err)
	}
	treeCopy := tree.Copy()
	wantLocation, wantContext := tree.ErrorContext(tree.Root.Nodes[0])
	gotLocation, gotContext := treeCopy.ErrorContext(treeCopy.Root.Nodes[0])
	if wantLocation != gotLocation {
		t.Errorf("wrong error location want %q got %q", wantLocation, gotLocation)
	}
	if wantContext != gotContext {
		t.Errorf("wrong error location want %q got %q", wantContext, gotContext)
	}
}

// All failures, and the result is a string that must appear in the error message.
var errorTests = []parseTest{
	// Check line numbers are accurate.
	// {"unclosed1",
	// 	"line1\n{{",
	// 	hasError, `unclosed1:2: unclosed action`},
	// {"unclosed2",
	// 	"line1\n{{define `x`}}line2\n{{",
	// 	hasError, `unclosed2:3: unclosed action`},
	// {"unclosed3",
	// 	"line1\n{{\"x\"\n\"y\"\n",
	// 	hasError, `unclosed3:4: unclosed action started at unclosed3:2`},
	// {"unclosed4",
	// 	"{{\n\n\n\n\n",
	// 	hasError, `unclosed4:6: unclosed action started at unclosed4:1`},
	{"var1",
		"`line1`\n\nx\n",
		hasError, `var1:3: function "x" not defined`},
	// Specific errors.
	{"function",
		"foo",
		hasError, `function "foo" not defined`},
	// {"comment1",
	// 	"{{/*}}",
	// 	hasError, `comment1:1: unclosed comment`},
	// {"comment2",
	// 	"{{/*\nhello\n}}",
	// 	hasError, `comment2:1: unclosed comment`},
	{"lparen",
		".X (1 2 3",
		hasError, `unclosed left paren`},
	// {"rparen",
	// 	"{{.X 1 2 3 ) }}",
	// 	hasError, `unexpected ")" in command`},
	{"rparen",
		".X 1 2 3 )",
		hasError, `unexpected right paren U+0029 ')'`},
	// {"rparen2",
	// 	"{{(.X 1 2 3",
	// 	hasError, `unclosed action`},
	{"space",
		"`x`3",
		hasError, `in operand`},
	{"idchar",
		"a#",
		hasError, `'#'`},
	{"charconst",
		"'a",
		hasError, `unterminated character constant`},
	{"stringconst",
		`"a`,
		hasError, `unterminated quoted string`},
	{"rawstringconst",
		"`a",
		hasError, `unterminated raw quoted string`},
	{"number",
		"0xi",
		hasError, `number syntax`},
	{"multidefine",
		"define `a`\n'a'\nend\n\ndefine `a`\n'b'\nend\n",
		hasError, `multiple definition of template`},
	{"eof",
		"range .X",
		hasError, `unexpected EOF`},
	{"variable",
		// Declare $x so it's defined, to avoid that error, and then check we don't parse a declaration.
		"$x := 23\nwith $x.y := 3\n$x 23\nend",
		hasError, `unexpected ":="`},
	{"multidecl",
		"$a,$b,$c := 23",
		hasError, `too many declarations`},
	{"undefvar",
		"$a",
		hasError, `undefined variable`},
	{"wrongdot",
		"true.any",
		hasError, `unexpected . after term`},
	{"wrongpipeline",
		"12|false",
		hasError, `non executable command in pipeline`},
	{"emptypipeline",
		`( )`,
		hasError, `missing value for parenthesized pipeline`},
	// {"multilinerawstring",
	// 	"{{ $v := `\n` }} {{",
	// 	hasError, `multilinerawstring:2: unclosed action`},
	{"rangeundefvar",
		"range $k\nend",
		hasError, `undefined variable`},
	{"rangeundefvars",
		"range $k, $v\nend",
		hasError, `undefined variable`},
	{"rangemissingvalue1",
		"range $k,\nend",
		hasError, `missing value for range`},
	{"rangemissingvalue2",
		"range $k, $v := \nend",
		hasError, `missing value for range`},
	{"rangenotvariable1",
		"range $k, .\nend",
		hasError, `range can only initialize variables`},
	{"rangenotvariable2",
		"range $k, 123 := .\nend",
		hasError, `range can only initialize variables`},
}

func TestErrors(t *testing.T) {
	for _, test := range errorTests {
		t.Run(test.name, func(t *testing.T) {
			_, err := New(test.name, nil).Parse(test.input, make(map[string]*Tree), nil)
			if err == nil {
				t.Fatalf("expected error %q, got nil", test.result)
			}
			if !strings.Contains(err.Error(), test.result) {
				t.Fatalf("error %q does not contain %q", err, test.result)
			}
		})
	}
}

func TestBlock(t *testing.T) {
	const (
		input = `"a"
block "inner" .
"bar"
.
"baz"
end
"b"`
		outer = `"a"
template "inner" .
"b"`
		outerExpected = `{{"a"}}{{template "inner" .}}{{"b"}}`

		inner = `"bar"
.
"baz"`
		innerExpected = `{{"bar"}}{{.}}{{"baz"}}`
	)
	treeSet := make(map[string]*Tree)
	tmpl, err := New("outer", nil).Parse(input, treeSet, nil)
	if err != nil {
		t.Fatal(err)
	}
	if g, expected := tmpl.Root.String(), outerExpected; g != expected {
		t.Errorf("outer template = %q, want %q", g, expected)
	}
	inTmpl := treeSet["inner"]
	if inTmpl == nil {
		t.Fatal("block did not define template")
	}
	if g, expected := inTmpl.Root.String(), innerExpected; g != expected {
		t.Errorf("inner template = %q, want %q", g, expected)
	}
}

func TestLineNum(t *testing.T) {
	const count = 100
	text := strings.Repeat("printf 1234\n", count)
	tree, err := New("bench", nil).Parse(text, make(map[string]*Tree), builtins)
	if err != nil {
		t.Fatal(err)
	}
	// Check the line numbers. Each line is an action containing a template, followed by text.
	// That's two nodes per line.
	nodes := tree.Root.Nodes
	for i := 0; i < len(nodes); i++ {
		line := 1 + i
		// Action first.
		action := nodes[i].(*ActionNode)
		if action.Line != line {
			t.Fatalf("line %d: action is line %d", line, action.Line)
		}
		pipe := action.Pipe
		if pipe.Line != line {
			t.Fatalf("line %d: pipe is line %d", line, pipe.Line)
		}
	}
}

func BenchmarkParseLarge(b *testing.B) {
	text := strings.Repeat("1234\n", 10000)
	for i := 0; i < b.N; i++ {
		_, err := New("bench", nil).Parse(text, make(map[string]*Tree), builtins)
		if err != nil {
			b.Fatal(err)
		}
	}
}

var sinkv, sinkl string

func BenchmarkVariableString(b *testing.B) {
	v := &VariableNode{
		Ident: []string{"$", "A", "BB", "CCC", "THIS_IS_THE_VARIABLE_BEING_PROCESSED"},
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sinkv = v.String()
	}
	if sinkv == "" {
		b.Fatal("Benchmark was not run")
	}
}

func BenchmarkListString(b *testing.B) {
	text := `
(printf .Field1.Field2.Field3).Value
$x := (printf .Field1.Field2.Field3).Value
$y := (printf $x.Field1.Field2.Field3).Value
$z := $y.Field1.Field2.Field3
if contains $y $z
	printf "%q" $y
else
	printf "%q" $x
end
with $z.Field1 | contains "boring"
	printf "%q" . | printf "%s"
else
	printf "%d %d %d" 11 11 11
	printf "%d %d %s" 22 22 $x.Field1.Field2.Field3 | printf "%s"
	printf "%v" (contains $z.Field1.Field2 $y)
end
`
	tree, err := New("bench", nil).Parse(text, make(map[string]*Tree), builtins)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sinkl = tree.Root.String()
	}
	if sinkl == "" {
		b.Fatal("Benchmark was not run")
	}
}
