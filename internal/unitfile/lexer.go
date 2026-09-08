package unitfile

import "fmt"

type TokenKind int

const (
	TokEOF TokenKind = iota
	TokComment
	TokSectionOpen
	TokSectionName
	TokSectionClose
	TokKey
	TokSep
	TokValue
	TokNewline
	TokError
)

type Token struct {
	Kind TokenKind
	Val  string
	Pos  Position
}

type lexer struct {
	src  []byte
	pos  int
	line int
	col  int
	toks []Token
	errs []error
}

func (l *lexer) next() byte {
	if l.pos >= len(l.src) {
		return 0
	}
	b := l.src[l.pos]
	l.pos++
	if b == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	return b
}

func (l *lexer) emit(kind TokenKind, val string) {
	tok := Token{
		Kind: kind,
		Val:  val,
		Pos: Position{
			Line: l.line,
			Col:  l.col - len(val),
		},
	}
	l.toks = append(l.toks, tok)
}

func (l *lexer) errorf(format string, args ...interface{}) {
	err := fmt.Errorf(format, args...)
	l.errs = append(l.errs, err)
	l.emit(TokError, err.Error())
}

func (l *lexer) peek() byte {
	if l.pos >= len(l.src) {
		return 0
	}
	return l.src[l.pos]
}

func (l *lexer) lexComment(start byte) {
	var buf []byte
	buf = append(buf, start)
	for {
		b := l.next()
		if b == 0 || b == '\n' {
			break
		}
		buf = append(buf, b)
	}
	l.emit(TokComment, string(buf))
}

func (l *lexer) lexSection() {
	l.emit(TokSectionOpen, "[")
	var buf []byte
	closed := false
	for {
		b := l.next()
		if b == ']' {
			closed = true
			break
		}
		if b == 0 || b == '\n' {
			break
		}
		buf = append(buf, b)
	}
	l.emit(TokSectionName, string(buf))
	if closed {
		l.emit(TokSectionClose, "]")
	} else {
		l.errorf("unterminated section header")
	}
}

func (l *lexer) lexKey(start byte) {
	var buf []byte
	buf = append(buf, start)
	containsEqual := false
	for {
		b := l.next()
		if b == '=' {
			containsEqual = true
		}

		if b == 0 || b == '\n' {
			break
		}
		buf = append(buf, b)
	}
	l.emit(TokKey, string(buf))
	if containsEqual {
		l.emit(TokSep, "=")
		var valBuf []byte
		for {
			b := l.next()
			if b == 0 || b == '\n' {
				break
			}
			valBuf = append(valBuf, b)
		}
		l.emit(TokValue, string(valBuf))
	} else {
		l.errorf("expected '=' after key")
	}
}

func (l *lexer) run() {
	for {
		b := l.next()
		switch b {
		case 0:
			l.emit(TokEOF, "")
			return
		case '\n':
			l.emit(TokNewline, "\n")
		case '#', ';':
			l.lexComment(b)
		case '[':
			l.lexSection()
		default:
			if isSpace(b) {
				continue
			}
			l.lexKey(b)
		}
	}
}

func isSpace(b byte) bool {
	return b == ' ' || b == '\t'
}

// Lex tokenizes a unit file's raw bytes.
func Lex(src []byte) ([]Token, []error) {
	l := &lexer{src: src, line: 1, col: 1}
	l.run()
	return l.toks, l.errs
}
