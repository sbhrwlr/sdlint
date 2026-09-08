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
	l.col++
	return b
}

func (l *lexer) nextline() {
	l.line++
	l.col = 1
}

func (l *lexer) emit(kind TokenKind, val string, startPos Position) {
	tok := Token{
		Kind: kind,
		Val:  val,
		Pos:  startPos,
	}
	l.toks = append(l.toks, tok)
}

func (l *lexer) errorf(format string, args ...interface{}) {
	startPos := l.mark()
	err := fmt.Errorf(format, args...)
	l.errs = append(l.errs, err)
	l.emit(TokError, err.Error(), startPos)
}

func (l *lexer) mark() Position {
	return Position{Line: l.line, Col: l.col}
}

func (l *lexer) peek() byte {
	if l.pos >= len(l.src) {
		return 0
	}
	return l.src[l.pos]
}

func (l *lexer) lexComment(start byte) {
	startPos := l.mark()
	var buf []byte
	buf = append(buf, start)
	for {
		b := l.next()
		if b == 0 || b == '\n' {
			break
		}
		buf = append(buf, b)
	}
	l.emit(TokComment, string(buf), startPos)
}

func (l *lexer) lexSection(startPos Position) {
	l.emit(TokSectionOpen, "[", startPos)

	nameStart := l.mark()
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
	l.emit(TokSectionName, string(buf), nameStart)

	closeStart := Position{Line: l.line, Col: l.col - 1}
	if closed {
		l.emit(TokSectionClose, "]", closeStart)
	} else {
		l.errorf("unterminated section header")
	}
}

func (l *lexer) lexKey(start byte, startPos Position) {

	var buf []byte
	buf = append(buf, start)

	for {
		b := l.peek()
		if b == '=' || b == 0 || b == '\n' {
			break
		}
		if isSpace(b) {
			break
		}
		buf = append(buf, l.next())
	}
	l.emit(TokKey, string(buf), startPos)

	for isSpace(l.peek()) {
		l.next()
	}

	if l.peek() != '=' {
		l.errorf("expected '=' after key")
		return
	}

	sepStart := l.mark()
	l.next()
	l.emit(TokSep, "=", sepStart)

	valStart := l.mark()
	var valBuf []byte
	for {
		b := l.peek()
		if b == 0 {
			break // EOF ends value regardless
		}
		if b == '\n' {
			if len(valBuf) > 0 && valBuf[len(valBuf)-1] == '\\' {
				valBuf = valBuf[:len(valBuf)-1] // drop trailing backslash
				l.next()                        // consume the newline via the single real path
				continue                        // keep scanning — line continues
			}
			break // real end of value; don't consume '\n' — let run() emit TokNewline
		}
		valBuf = append(valBuf, l.next())
	}
	l.emit(TokValue, string(valBuf), valStart)
}

func (l *lexer) run() {
	for {

		startPos := l.mark()

		b := l.next()
		switch b {
		case 0:
			l.emit(TokEOF, "", startPos)
			return
		case '\n':
			l.emit(TokNewline, "\n", startPos)
			l.nextline()
		case '#', ';':
			l.lexComment(b)
		case '[':
			l.lexSection(startPos)
		default:
			if isSpace(b) {
				continue
			}
			l.lexKey(b, startPos)
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
