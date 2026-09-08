package unitfile

import (
	"os"
	"path/filepath"
	"testing"
)

// tok is a compact expectation: kind, value, and start position.
type tok struct {
	kind TokenKind
	val  string
	line int
	col  int
}

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "rules", name))
	if err != nil {
		t.Fatalf("reading fixture %s: %v", name, err)
	}
	return data
}

func assertTokens(t *testing.T, got []Token, want []tok) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("token count mismatch: got %d, want %d\ngot: %+v", len(got), len(want), got)
	}
	for i, w := range want {
		g := got[i]
		if g.Kind != w.kind || g.Val != w.val || g.Pos.Line != w.line || g.Pos.Col != w.col {
			t.Errorf("token %d: got {kind:%d val:%q pos:%d:%d}, want {kind:%d val:%q pos:%d:%d}",
				i, g.Kind, g.Val, g.Pos.Line, g.Pos.Col, w.kind, w.val, w.line, w.col)
		}
	}
}

func TestLexContinuation(t *testing.T) {
	src := readFixture(t, "continuation.service")
	toks, errs := Lex(src)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	assertTokens(t, toks, []tok{
		{TokSectionOpen, "[", 1, 1},
		{TokSectionName, "Service", 1, 2},
		{TokSectionClose, "]", 1, 9},
		{TokNewline, "\n", 1, 10},
		{TokKey, "ExecStart", 2, 1},
		{TokSep, "=", 2, 10},
		{TokValue, "/opt/myapp/bin/server   --config /etc/myapp/config.yaml", 2, 11},
		{TokNewline, "\n", 3, 34},
		{TokEOF, "", 4, 1},
	})
}

func TestLexContinuationAtEOF(t *testing.T) {
	src := readFixture(t, "continuation-eof.service")
	_, errs := Lex(src)
	if len(errs) == 0 {
		t.Fatal("expected an error for dangling continuation at EOF, got none")
	}
}

func TestLexSpacedKV(t *testing.T) {
	src := readFixture(t, "spaced-kv.service")
	toks, errs := Lex(src)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	assertTokens(t, toks, []tok{
		{TokSectionOpen, "[", 1, 1},
		{TokSectionName, "Service", 1, 2},
		{TokSectionClose, "]", 1, 9},
		{TokNewline, "\n", 1, 10},
		{TokKey, "Key", 2, 1},
		{TokSep, "=", 2, 5},
		{TokValue, "value", 2, 7},
		{TokNewline, "\n", 2, 12},
		{TokKey, "Key2", 3, 1},
		{TokSep, "=", 3, 5},
		{TokValue, "value2", 3, 8},
		{TokNewline, "\n", 3, 14},
		{TokKey, "Key3", 4, 1},
		{TokSep, "=", 4, 8},
		{TokValue, "value3", 4, 9},
		{TokNewline, "\n", 4, 15},
		{TokEOF, "", 5, 1},
	})
}

func TestLexMalformedKeyRecovers(t *testing.T) {
	src := readFixture(t, "malformed-key.service")
	toks, errs := Lex(src)
	if len(errs) != 1 {
		t.Fatalf("expected exactly 1 error, got %d: %v", len(errs), errs)
	}

	var foundKey, foundVal bool
	for _, tk := range toks {
		if tk.Kind == TokKey && tk.Val == "NextKey" {
			foundKey = true
		}
		if tk.Kind == TokValue && tk.Val == "ok" {
			foundVal = true
		}
	}
	if !foundKey || !foundVal {
		t.Errorf("lexer did not recover after malformed line; tokens: %+v", toks)
	}
}

func TestLexUnterminatedSectionRecovers(t *testing.T) {
	src := readFixture(t, "unterminated-section.service")
	toks, errs := Lex(src)
	if len(errs) != 1 {
		t.Fatalf("expected exactly 1 error, got %d: %v", len(errs), errs)
	}

	var foundKey, foundVal bool
	for _, tk := range toks {
		if tk.Kind == TokKey && tk.Val == "Key" {
			foundKey = true
		}
		if tk.Kind == TokValue && tk.Val == "value" {
			foundVal = true
		}
	}
	if !foundKey || !foundVal {
		t.Errorf("lexer did not recover after unterminated section; tokens: %+v", toks)
	}
}

func TestLexBlankAndComments(t *testing.T) {
	src := readFixture(t, "blank-and-comments.service")
	toks, errs := Lex(src)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	assertTokens(t, toks, []tok{
		{TokComment, "# top comment", 1, 1},
		{TokNewline, "\n", 1, 14},
		{TokSectionOpen, "[", 2, 1},
		{TokSectionName, "Unit", 2, 2},
		{TokSectionClose, "]", 2, 6},
		{TokNewline, "\n", 2, 7},
		{TokComment, "; semicolon comment", 3, 1},
		{TokNewline, "\n", 3, 20},
		{TokKey, "Description", 4, 1},
		{TokSep, "=", 4, 12},
		{TokValue, "Test", 4, 13},
		{TokNewline, "\n", 4, 17},
		{TokNewline, "\n", 5, 1},
		{TokSectionOpen, "[", 6, 1},
		{TokSectionName, "Service", 6, 2},
		{TokSectionClose, "]", 6, 9},
		{TokNewline, "\n", 6, 10},
		{TokKey, "Key", 7, 1},
		{TokSep, "=", 7, 4},
		{TokValue, "value", 7, 5},
		{TokNewline, "\n", 7, 10},
		{TokEOF, "", 8, 1},
	})
}

func TestLexNoTrailingNewline(t *testing.T) {
	src := readFixture(t, "no-trailing-newline.service")
	toks, errs := Lex(src)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	assertTokens(t, toks, []tok{
		{TokSectionOpen, "[", 1, 1},
		{TokSectionName, "Service", 1, 2},
		{TokSectionClose, "]", 1, 9},
		{TokNewline, "\n", 1, 10},
		{TokKey, "Key", 2, 1},
		{TokSep, "=", 2, 4},
		{TokValue, "value", 2, 5},
		{TokEOF, "", 2, 10},
	})
}
