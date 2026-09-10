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
	file string
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

func (l *lexer) emit(kind TokenKind, val string, startPos Position) {
	tok := Token{
		Kind: kind,
		Val:  val,
		Pos:  startPos,
	}
	l.toks = append(l.toks, tok)
}

func (l *lexer) errorf(pos Position, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	l.errs = append(l.errs, fmt.Errorf("%s:%d:%d: %s", l.file, pos.Line, pos.Col, msg))
	l.emit(TokError, msg, pos)
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

func (l *lexer) lexComment(start byte, startPos Position) {
	var buf []byte
	buf = append(buf, start)
	for {
		b := l.peek()
		if b == 0 || b == '\n' {
			break
		}
		buf = append(buf, l.next())
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
		l.errorf(startPos, "unterminated section header")
	}
}

func (l *lexer) lexKey(start byte, startPos Position) {
	name := l.scanKeyName(start)
	l.emit(TokKey, name, startPos)

	l.skipSpaces()
	eqPos := l.mark()
	if l.peek() != '=' {
		l.errorf(eqPos, "expected '=' after key")
		return
	}

	l.next()
	l.emit(TokSep, "=", eqPos)

	l.skipSpaces()
	valStart := l.mark()
	val, err := l.scanValue()
	if err != nil {
		l.errorf(valStart, "%s", err)
	}
	l.emit(TokValue, val, valStart)
}

func (l *lexer) scanKeyName(start byte) string {
	buf := []byte{start}
	for {
		b := l.peek()
		if b == '=' || b == 0 || b == '\n' || isSpace(b) {
			break
		}
		buf = append(buf, l.next())
	}
	return string(buf)
}

func (l *lexer) skipSpaces() {
	for isSpace(l.peek()) {
		l.next()
	}
}

func (l *lexer) scanValue() (string, error) {
	var buf []byte
	for {
		b := l.peek()
		switch {
		case b == 0:
			if len(buf) > 0 && buf[len(buf)-1] == '\\' {
				return string(buf[:len(buf)-1]), fmt.Errorf("dangling line continuation at EOF")
			}
			return string(buf), nil
		case b == '\n' && len(buf) > 0 && buf[len(buf)-1] == '\\':
			buf = buf[:len(buf)-1]
			l.next()
		case b == '\n':
			return string(buf), nil
		default:
			buf = append(buf, l.next())
		}
	}
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
		case '#', ';':
			l.lexComment(b, startPos)
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
func Lex(file string, src []byte) ([]Token, []error) {
	l := &lexer{file: file, src: src, line: 1, col: 1}
	l.run()
	return l.toks, l.errs
}
