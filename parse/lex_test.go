// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parse

import (
	"fmt"
	"testing"
)

// Make the types prettyprint.
var itemName = map[itemType]string{
	itemError:        "error",
	itemBool:         "bool",
	itemChar:         "char",
	itemCharConstant: "charconst",
	itemComment:      "comment",
	itemComplex:      "complex",
	itemDeclare:      ":=",
	itemEOF:          "EOF",
	itemField:        "field",
	itemIdentifier:   "identifier",
	// itemLeftDelim:    "left delim",
	itemLeftParen:  "(",
	itemNumber:     "number",
	itemPipe:       "pipe",
	itemRawString:  "raw string",
	itemRightDelim: "right delim",
	itemRightParen: ")",
	itemSpace:      "space",
	itemString:     "string",
	itemVariable:   "variable",

	// keywords
	itemDot:      ".",
	itemBlock:    "block",
	itemBreak:    "break",
	itemContinue: "continue",
	itemDefine:   "define",
	itemElse:     "else",
	itemIf:       "if",
	itemEnd:      "end",
	itemNil:      "nil",
	itemRange:    "range",
	itemTemplate: "template",
	itemWith:     "with",
}

func (i itemType) String() string {
	s := itemName[i]
	if s == "" {
		return fmt.Sprintf("item%d", int(i))
	}
	return s
}

type lexTest struct {
	name  string
	input string
	items []item
}

func mkItem(typ itemType, text string) item {
	return item{
		typ: typ,
		val: text,
	}
}

var (
	tDot        = mkItem(itemDot, ".")
	tBlock      = mkItem(itemBlock, "block")
	tEOF        = mkItem(itemEOF, "")
	tFor        = mkItem(itemIdentifier, "for")
	tLeft       = mkItem(itemLeftDelim, "")
	tLpar       = mkItem(itemLeftParen, "(")
	tPipe       = mkItem(itemPipe, "|")
	tQuote      = mkItem(itemString, `"abc \n\t\" "`)
	tRange      = mkItem(itemRange, "range")
	tRight      = mkItem(itemRightDelim, "")
	tRpar       = mkItem(itemRightParen, ")")
	tSpace      = mkItem(itemSpace, " ")
	raw         = "`" + `abc\n\t\" ` + "`"
	rawNL       = "`now is{{\n}}the time`" // Contains newline inside raw quote.
	tRawQuote   = mkItem(itemRawString, raw)
	tRawQuoteNL = mkItem(itemRawString, rawNL)
)

var lexTests = []lexTest{
	{"empty", "", []item{tEOF}},
	{"spaces", " \t\n", []item{tEOF}},
	{"identifiers", `now is the time`, []item{
		tLeft,
		mkItem(itemIdentifier, "now"),
		tSpace,
		mkItem(itemIdentifier, "is"),
		tSpace,
		mkItem(itemIdentifier, "the"),
		tSpace,
		mkItem(itemIdentifier, "time"),
		tRight,
		tEOF,
	}},
	{"semi-colon separated actions", "`x`;foo;;;", []item{
		tLeft,
		mkItem(itemRawString, "`x`"),
		tRight,
		tLeft,
		mkItem(itemIdentifier, "foo"),
		tRight,
		tLeft,
		tRight,
		tLeft,
		tRight,
		tEOF,
	}},
	{"identifier with comment", "hello # this is a comment", []item{
		tLeft,
		mkItem(itemIdentifier, "hello"),
		tRight,
		mkItem(itemComment, " this is a comment"),
		tEOF,
	}},
	{"punctuation", ",@% ", []item{
		tLeft,
		mkItem(itemChar, ","),
		mkItem(itemChar, "@"),
		mkItem(itemChar, "%"),
		tRight,
		tEOF,
	}},
	{"parens", "((3))", []item{
		tLeft,
		tLpar,
		tLpar,
		mkItem(itemNumber, "3"),
		tRpar,
		tRpar,
		tRight,
		tEOF,
	}},
	{"for", `for`, []item{tLeft, tFor, tRight, tEOF}},
	{"block", `block "foo" .`, []item{
		tLeft, tBlock, tSpace, mkItem(itemString, `"foo"`), tSpace, tDot, tRight, tEOF,
	}},
	{"quote", `"abc \n\t\" "`, []item{tLeft, tQuote, tRight, tEOF}},
	{"raw quote", raw, []item{tLeft, tRawQuote, tRight, tEOF}},
	{"raw quote with newline", rawNL, []item{tLeft, tRawQuoteNL, tRight, tEOF}},
	{"numbers", "1 02 0x14 0X14 -7.2i 1e3 1E3 +1.2e-4 4.2i 1+2i 1_2 0x1.e_fp4 0X1.E_FP4", []item{
		tLeft,
		mkItem(itemNumber, "1"),
		tSpace,
		mkItem(itemNumber, "02"),
		tSpace,
		mkItem(itemNumber, "0x14"),
		tSpace,
		mkItem(itemNumber, "0X14"),
		tSpace,
		mkItem(itemNumber, "-7.2i"),
		tSpace,
		mkItem(itemNumber, "1e3"),
		tSpace,
		mkItem(itemNumber, "1E3"),
		tSpace,
		mkItem(itemNumber, "+1.2e-4"),
		tSpace,
		mkItem(itemNumber, "4.2i"),
		tSpace,
		mkItem(itemComplex, "1+2i"),
		tSpace,
		mkItem(itemNumber, "1_2"),
		tSpace,
		mkItem(itemNumber, "0x1.e_fp4"),
		tSpace,
		mkItem(itemNumber, "0X1.E_FP4"),
		tRight,
		tEOF,
	}},
	{"characters", `'a' '\n' '\'' '\\' '\u00FF' '\xFF' '本'`, []item{
		tLeft,
		mkItem(itemCharConstant, `'a'`),
		tSpace,
		mkItem(itemCharConstant, `'\n'`),
		tSpace,
		mkItem(itemCharConstant, `'\''`),
		tSpace,
		mkItem(itemCharConstant, `'\\'`),
		tSpace,
		mkItem(itemCharConstant, `'\u00FF'`),
		tSpace,
		mkItem(itemCharConstant, `'\xFF'`),
		tSpace,
		mkItem(itemCharConstant, `'本'`),
		tRight,
		tEOF,
	}},
	{"bools", "true false", []item{
		tLeft,
		mkItem(itemBool, "true"),
		tSpace,
		mkItem(itemBool, "false"),
		tRight,
		tEOF,
	}},
	{"dot", ".", []item{
		tLeft,
		tDot,
		tRight,
		tEOF,
	}},
	{"nil", "nil", []item{
		tLeft,
		mkItem(itemNil, "nil"),
		tRight,
		tEOF,
	}},
	{"dots", ".x . .2 .x.y.z", []item{
		tLeft,
		mkItem(itemField, ".x"),
		tSpace,
		tDot,
		tSpace,
		mkItem(itemNumber, ".2"),
		tSpace,
		mkItem(itemField, ".x"),
		mkItem(itemField, ".y"),
		mkItem(itemField, ".z"),
		tRight,
		tEOF,
	}},
	{"keywords", "range if else end with", []item{
		tLeft,
		mkItem(itemRange, "range"),
		tSpace,
		mkItem(itemIf, "if"),
		tSpace,
		mkItem(itemElse, "else"),
		tSpace,
		mkItem(itemEnd, "end"),
		tSpace,
		mkItem(itemWith, "with"),
		tRight,
		tEOF,
	}},
	{"variables", "$c := printf $ $hello $23 $ $var.Field .Method", []item{
		tLeft,
		mkItem(itemVariable, "$c"),
		tSpace,
		mkItem(itemDeclare, ":="),
		tSpace,
		mkItem(itemIdentifier, "printf"),
		tSpace,
		mkItem(itemVariable, "$"),
		tSpace,
		mkItem(itemVariable, "$hello"),
		tSpace,
		mkItem(itemVariable, "$23"),
		tSpace,
		mkItem(itemVariable, "$"),
		tSpace,
		mkItem(itemVariable, "$var"),
		mkItem(itemField, ".Field"),
		tSpace,
		mkItem(itemField, ".Method"),
		tRight,
		tEOF,
	}},
	{"variable invocation", "$x 23", []item{
		tLeft,
		mkItem(itemVariable, "$x"),
		tSpace,
		mkItem(itemNumber, "23"),
		tRight,
		tEOF,
	}},
	{"pipeline", `echo hi 1.2 |noargs|args 1 "hi"`, []item{
		tLeft,
		mkItem(itemIdentifier, "echo"),
		tSpace,
		mkItem(itemIdentifier, "hi"),
		tSpace,
		mkItem(itemNumber, "1.2"),
		tSpace,
		tPipe,
		mkItem(itemIdentifier, "noargs"),
		tPipe,
		mkItem(itemIdentifier, "args"),
		tSpace,
		mkItem(itemNumber, "1"),
		tSpace,
		mkItem(itemString, `"hi"`),
		tRight,
		tEOF,
	}},
	{"declaration", "$v := 3", []item{
		tLeft,
		mkItem(itemVariable, "$v"),
		tSpace,
		mkItem(itemDeclare, ":="),
		tSpace,
		mkItem(itemNumber, "3"),
		tRight,
		tEOF,
	}},
	{"2 declarations", "$v , $w := 3", []item{
		tLeft,
		mkItem(itemVariable, "$v"),
		tSpace,
		mkItem(itemChar, ","),
		tSpace,
		mkItem(itemVariable, "$w"),
		tSpace,
		mkItem(itemDeclare, ":="),
		tSpace,
		mkItem(itemNumber, "3"),
		tRight,
		tEOF,
	}},
	{"field of parenthesized expression", "(.X).Y", []item{
		tLeft,
		tLpar,
		mkItem(itemField, ".X"),
		tRpar,
		mkItem(itemField, ".Y"),
		tRight,
		tEOF,
	}},
	// {"trimming spaces before and after", "hello- {{- 3 -}} -world", []item{
	// 	mkItem(itemText, "hello-"),
	// 	tLeft,
	// 	mkItem(itemNumber, "3"),
	// 	tRight,
	// 	mkItem(itemText, "-world"),
	// 	tEOF,
	// }},
	// {"trimming spaces before and after comment", "hello- {{- /* hello */ -}} -world", []item{
	// 	mkItem(itemText, "hello-"),
	// 	mkItem(itemComment, "/* hello */"),
	// 	mkItem(itemText, "-world"),
	// 	tEOF,
	// }},
	// errors
	{"badchar", "\x01", []item{
		tLeft,
		mkItem(itemError, "unrecognized character in action: U+0001"),
	}},
	// {"unclosed action", "{{", []item{
	// 	tLeft,
	// 	mkItem(itemError, "unclosed action"),
	// }},
	// {"EOF in action", "{{range", []item{
	// 	tLeft,
	// 	tRange,
	// 	mkItem(itemError, "unclosed action"),
	// }},
	{"unclosed quote", "\"\n\"", []item{
		tLeft,
		mkItem(itemError, "unterminated quoted string"),
	}},
	{"unclosed raw quote", "`xx", []item{
		tLeft,
		mkItem(itemError, "unterminated raw quoted string"),
	}},
	{"unclosed char constant", "'\n", []item{
		tLeft,
		mkItem(itemError, "unterminated character constant"),
	}},
	{"bad number", "3k", []item{
		tLeft,
		mkItem(itemError, `bad number syntax: "3k"`),
	}},
	{"unclosed paren", "(3", []item{
		tLeft,
		tLpar,
		mkItem(itemNumber, "3"),
		mkItem(itemError, `unclosed left paren`),
	}},
	{"extra right paren", "3)", []item{
		tLeft,
		mkItem(itemNumber, "3"),
		// tRpar,
		mkItem(itemError, `unexpected right paren U+0029 ')'`),
	}},

	// Fixed bugs
	// Many elements in an action blew the lookahead until
	// we made lexInsideAction not loop.
	{"long pipeline deadlock", "|||||", []item{
		tLeft,
		tPipe,
		tPipe,
		tPipe,
		tPipe,
		tPipe,
		tRight,
		tEOF,
	}},
}

// collect gathers the emitted items into a slice.
func collect(t *lexTest) (items []item) {
	l := lex(t.name, t.input, true)
	for {
		item := l.nextItem()
		items = append(items, item)
		if item.typ == itemEOF || item.typ == itemError {
			break
		}
	}
	return
}

func equal(i1, i2 []item, checkPos bool) bool {
	if len(i1) != len(i2) {
		return false
	}
	for k := range i1 {
		if i1[k].typ != i2[k].typ {
			return false
		}
		if i1[k].val != i2[k].val {
			return false
		}
		if checkPos && i1[k].pos != i2[k].pos {
			return false
		}
		if checkPos && i1[k].line != i2[k].line {
			return false
		}
	}
	return true
}

func TestLex(t *testing.T) {
	for _, test := range lexTests {
		items := collect(&test)
		if !equal(items, test.items, false) {
			t.Errorf("\n%s: got\n\t%+v\nexpected\n\t%v", test.name, items, test.items)
		}
	}
}

var lexPosTests = []lexTest{
	{"empty", "", []item{{itemEOF, 0, "", 1, true}}},
	// {"punctuation", "{{,@%#}}", []item{
	// 	{itemLeftDelim, 0, "{{", 1, true},
	// 	{itemChar, 2, ",", 1, true},
	// 	{itemChar, 3, "@", 1, true},
	// 	{itemChar, 4, "%", 1, true},
	// 	{itemChar, 5, "#", 1, true},
	// 	{itemRightDelim, 6, "}}", 1, true},
	// 	{itemEOF, 8, "", 1, true},
	// }},
	// {"sample", "0123{{hello}}xyz", []item{
	// 	{itemText, 0, "0123", 1, true},
	// 	{itemLeftDelim, 4, "{{", 1, true},
	// 	{itemIdentifier, 6, "hello", 1, true},
	// 	{itemRightDelim, 11, "}}", 1, true},
	// 	{itemText, 13, "xyz", 1, true},
	// 	{itemEOF, 16, "", 1, true},
	// }},
	// {"trimafter", "{{x -}}\n{{y}}", []item{
	// 	{itemLeftDelim, 0, "{{", 1, true},
	// 	{itemIdentifier, 2, "x", 1, true},
	// 	{itemRightDelim, 5, "}}", 1, true},
	// 	{itemLeftDelim, 8, "{{", 2, true},
	// 	{itemIdentifier, 10, "y", 2, true},
	// 	{itemRightDelim, 11, "}}", 2, true},
	// 	{itemEOF, 13, "", 2, true},
	// }},
	// {"trimbefore", "{{x}}\n{{- y}}", []item{
	// 	{itemLeftDelim, 0, "{{", 1, true},
	// 	{itemIdentifier, 2, "x", 1, true},
	// 	{itemRightDelim, 3, "}}", 1, true},
	// 	{itemLeftDelim, 6, "{{", 2, true},
	// 	{itemIdentifier, 10, "y", 2, true},
	// 	{itemRightDelim, 11, "}}", 2, true},
	// 	{itemEOF, 13, "", 2, true},
	// }},
}

// The other tests don't check position, to make the test cases easier to construct.
// This one does.
func TestPos(t *testing.T) {
	for _, test := range lexPosTests {
		items := collect(&test)
		if !equal(items, test.items, true) {
			t.Errorf("\n%s: got\n\t%v\nexpected\n\t%v", test.name, items, test.items)
			if len(items) == len(test.items) {
				// Detailed print; avoid item.String() to expose the position value.
				for i := range items {
					if !equal(items[i:i+1], test.items[i:i+1], true) {
						i1 := items[i]
						i2 := test.items[i]
						t.Errorf("\t#%d: got {%v %d %q %d} expected {%v %d %q %d}",
							i, i1.typ, i1.pos, i1.val, i1.line, i2.typ, i2.pos, i2.val, i2.line)
					}
				}
			}
		}
	}
}

// Test that an error shuts down the lexing goroutine.
// func TestShutdown(t *testing.T) {
// 	// We need to duplicate template.Parse here to hold on to the lexer.
// 	const text = "erroneous{{define}}{{else}}1234"
// 	lexer := lex("foo", text, "{{", "}}", false)
// 	_, err := New("root", nil).parseLexer(lexer)
// 	if err == nil {
// 		t.Fatalf("expected error")
// 	}
// 	// The error should have drained the input. Therefore, the lexer should be shut down.
// 	token, ok := <-lexer.items
// 	if ok {
// 		t.Fatalf("input was not drained; got %v", token)
// 	}
// }

// parseLexer is a local version of parse that lets us pass in the lexer instead of building it.
// We expect an error, so the tree set and funcs list are explicitly nil.
func (t *Tree) parseLexer(lex *lexer) (tree *Tree, err error) {
	defer t.recover(&err)
	t.ParseName = t.Name
	t.startParse(nil, lex, map[string]*Tree{})
	t.parse()
	t.add()
	t.stopParse()
	return t, nil
}