package parser

import (
	"aspire/are/main/lib"
)

type Lexer struct {
	path           int
	buffer         []byte
	len            uintptr
	line, col, pos uintptr
	tokens         []Token
	// the last token location
	loc lib.Loc
}

func (l *Lexer) Debug(tk Token) {
	lib.Printf("token %s (%s) [%d:%d, %d:%d]%s", tk.tag.Lexeme(), l.src(tk), tk.loc.Line, tk.loc.Col, tk.loc.Start, tk.loc.End, lib.EOL)
}

func NewLexer(buffer []byte, path int) *Lexer {
	return &Lexer{
		buffer: buffer,
		len:    uintptr(len(buffer)),
		line:   1,
		col:    1,
		pos:    0,
		path:   path,
	}
}

func (l *Lexer) char() rune {
	if l.pos >= l.len {
		return 0
	}
	return rune(l.buffer[l.pos])
}

func (l *Lexer) charAt(idx uintptr) rune {
	if l.pos+idx >= l.len {
		return 0
	}
	// if l.pos+idx >= l.len {
	// 	lib.Panic("index is out of bounds")
	// }
	return rune(l.buffer[l.pos+idx])
}

func (l *Lexer) lex() bool {
	return l.pos < l.len
	// return l.pos < l.len && l.char() != 0
}

// The lexer's current location in source.
func (l *Lexer) Loc() lib.Loc {
	return lib.Loc{
		Line:  l.line,
		Col:   l.col,
		Start: l.pos,
		End:   l.pos + 1,
	}
}

func (l *Lexer) Tokenize() []Token {
	l.tokens = []Token{}
	for l.lex() {
		l.handleChar(l.char())
	}
	// // TODO: review this line..
	// tokens = append(make([]Token, cap(tokens)*2), tokens...)
	l.push(Token{
		tag: EOF,
		loc: l.Loc(),
	})
	return l.tokens
}
func (l *Lexer) advance(n uintptr) {
	if l.pos+n-1 >= l.len {
		lib.Panic("strange behaviour")
	}
	l.col += n
	l.pos += n
	l.loc.End = l.pos
}

func (l *Lexer) push(tk Token) {
	l.tokens = append(l.tokens, tk)
}

func (l *Lexer) src(tk Token) string {
	return string(l.buffer[tk.loc.Start:tk.loc.End])
}

func token(tag TokenTag, loc lib.Loc) Token {
	return Token{
		tag: tag,
		loc: loc,
	}
}

// // Make sure to advance before ending the token
// // and end the token before pushing.
// func endToken(l *Lexer) {
// 	l.loc.End = l.pos
// }

func startToken(l *Lexer) {
	l.loc = l.Loc()
}

const (
	ExpectedNumericDigitInFloat           = "Expected numeric digit in float literal."
	UnclosedString                        = "Unterminated string literal."
	_InNumberMustSeparateSuccessiveDigits = "'_' in number literal must separate successive digits."
	UnexpectedEOT                         = "Unexpected end of text."
)

func (l *Lexer) handleChar(b rune) {
	switch b {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		l.handleNumber()
	case '\n':
		l.advance(1)
		l.line++
		l.col = 1
	case ' ', '\t', '\r':
		l.advance(1)
	case '"', '\'':
		l.handleQuote()
	case ':':
		l.push(token(COLON, l.Loc()))
		l.advance(1)
	case ';':
		l.push(token(SEMICOLON, l.Loc()))
		l.advance(1)
	case '.':
		l.handleDot()
	case ',':
		l.push(token(COMMA, l.Loc()))
		l.advance(1)
	case '(':
		l.push(token(O_PAREN, l.Loc()))
		l.advance(1)
	case ')':
		l.push(token(C_PAREN, l.Loc()))
		l.advance(1)
	case '{':
		l.push(token(O_BRACE, l.Loc()))
		l.advance(1)
	case '}':
		l.push(token(C_BRACE, l.Loc()))
		l.advance(1)
	case '[':
		l.push(token(O_BRACKET, l.Loc()))
		l.advance(1)
	case ']':
		l.push(token(C_BRACKET, l.Loc()))
		l.advance(1)
	case '^':
		l.handleXOR()
	case '~':
		l.push(token(BITNOT, l.Loc()))
		l.advance(1)
	case '+', '-', '*', '%':
		l.handleBinaryOp()
	case '/':
		l.handleForwardSlash()
	case '$':
		l.handleComment()
	case '=':
		l.handleComparison()
	case '!':
		l.handleLogicalOp()
	case '?':
		l.handleQuestionMark()
	case '|', '&':
		l.handleLogicalOp()
	case '>', '<':
		l.handleComparison()
	case '@':
		if lib.IsAlpha(l.charAt(1)) {
			l.handleLabel()
			return
		}
		fallthrough
	default:
		if lib.IsAlpha(l.char()) {
			l.handleIdentifier()
		} else {
			l.lexError(UnrecognisedChar)
		}
	}
}

func (l *Lexer) handleXOR() {
	var tag = XOR
	startToken(l)
	l.advance(1)
	if l.char() == '=' {
		l.advance(1)
		tag = XOR_EQUALS
	}
	// endToken(l)
	l.push(token(tag, l.loc))
}

func (l *Lexer) handleQuestionMark() {
	startToken(l)
	l.advance(1)
	var tag TokenTag = QUESTION
	if l.char() == '?' {
		l.advance(1)
		tag = NULLISH
	}
	if l.char() == '=' {
		l.advance(1)
		tag = NULLISH_EQUALS
	}
	// endToken(l)
	l.push(token(tag, l.loc))
}

func (l *Lexer) handleLogicalOp() {
	startToken(l)
	var tag TokenTag
	switch l.char() {
	case '!':
		l.handleComparison()
		return
	case '&':
		switch l.charAt(1) {
		case '&':
			l.advance(2)
			tag = AND
		case '^':
			if l.charAt(2) == '=' {
				tag = BITCLEAR_EQUALS
				l.advance(3)
			} else {
				l.advance(2)
				tag = BITCLEAR
			}
		case '=':
			l.advance(2)
			tag = BITAND_EQUALS
		default:
			l.advance(1)
			tag = BITAND
		}
	case '|':
		switch l.charAt(1) {
		case '|':
			l.advance(2)
			tag = OR
		case '=':
			l.advance(2)
			tag = BITOR_EQUALS
		default:
			l.advance(1)
			tag = BITOR
		}
	}
	// endToken(l)
	l.push(token(tag, l.loc))
}

func (l *Lexer) handleComparison() {
	startToken(l)
	var tag TokenTag
	switch l.char() {
	case '=':
		tag = EQUALS
	case '>':
		tag = GT
		if l.charAt(1) == '>' {
			l.advance(1)
			if l.charAt(1) == '=' {
				l.advance(1)
				tag = SHBITR_EQUALS
			} else {
				tag = SHBITR
			}
		}
	case '<':
		tag = LT
		if l.charAt(1) == '<' {
			l.advance(1)
			if l.charAt(1) == '=' {
				l.advance(1)
				tag = SHBITL_EQUALS
			} else {
				tag = SHBITL
			}
		}
	case '!':
		tag = NOT
	default:
		lib.Panic("unhandled comaparison op")
	}
	l.advance(1)

	ch := l.char()
	if ch == '=' {
		switch tag {
		case EQUALS:
			if l.charAt(1) == '=' {
				l.advance(1)
				tag = EQUALS3
			} else {
				tag = EQUALS2
			}
		case GT:
			tag = GT_EQUALS
		case LT:
			tag = LT_EQUALS
		case NOT:
			if l.charAt(1) == '=' {
				l.advance(1)
				tag = NOT_EQUALS2
			} else {
				tag = NOT_EQUALS
			}
		}
		l.advance(1)
	} else if tag == EQUALS && ch == '>' {
		tag = ARROW
		l.advance(1)
	}
	// endToken(l)
	l.push(token(tag, l.loc))
}

// func (l *Lexer) handleEquals() {
// 	startToken(l)
// 	l.push(token(EQUALS, l.loc))
// 	l.advance(1)
// }

func (l *Lexer) handleComment() {
	switch l.char() {
	case '/':
		l.advance(1) // consume first /
		if l.charAt(0) == '*' {
			l.advance(1) // consume *
			for l.lex() {
				if l.char() == '*' && l.charAt(1) == '/' {
					l.advance(2)
					return
				}
				if l.char() == '\n' {
					l.advance(1)
					l.line++
					l.col = 1
					continue
				}
				l.advance(1)
			}
		} else {
			for l.char() != '\n' && l.lex() {
				l.advance(1)
			}
		}
	case '$':
		l.advance(1)
		for l.char() != '\n' && l.char() != '$' && l.lex() {
			l.advance(1)
		}
		if l.char() == '$' {
			l.advance(1)
		}
	default:
		lib.Panic("unhandled comment")
	}
}

func (l *Lexer) handleForwardSlash() {
	if l.charAt(1) == '/' || l.charAt(1) == '*' {
		l.handleComment()
	} else {
		l.handleBinaryOp()
	}
}

func (l *Lexer) handleBinaryOp() {
	startToken(l)
	char := l.char()
	l.advance(1)
	var tag TokenTag
	switch char {
	case '*':
		if l.char() == '*' {
			l.advance(1)
			tag = STAR2
		} else {
			tag = TIMES
		}
	case '+':
		tag = PLUS
		if l.char() == '+' {
			l.advance(1)
			tag = PLUS2
		}
	case '-':
		tag = MINUS
		if l.char() == '-' {
			l.advance(1)
			tag = MINUS2
		}
	case '/':
		tag = DIVIDE
	case '%':
		tag = MODULO
	default:
		lib.Panic("unhandled operator")
	}
	if l.char() == '=' && tag != MINUS2 && tag != PLUS2 {
		l.handleAssignmentOp(tag)
		return
	}
	// endToken(l)
	l.push(token(tag, l.loc))
}

func (l *Lexer) handleAssignmentOp(t TokenTag) {
	var tag TokenTag
	switch t {
	case STAR2:
		tag = STAR2_EQUALS
	case PLUS:
		tag = PLUS_EQUALS
	case MINUS:
		tag = MINUS_EQUALS
	case DIVIDE:
		tag = DIVIDE_EQUALS
	case MODULO:
		tag = MODULO_EQUALS
	case NULLISH:
		tag = NULLISH_EQUALS
	default:
		lib.Panic("unhandled operator")
	}
	l.advance(1) // equals sign
	// endToken(l)
	l.push(token(tag, l.loc))
}

func (l *Lexer) handleDot() {
	startToken(l)
	if l.charAt(1) == '.' && l.charAt(2) == '.' {
		l.advance(3)
		// endToken(l)
		l.push(token(DOT3, l.loc))
	} else if lib.IsDigit(l.charAt(1)) {
		l.advance(1)
		l.handleFloat()
		l.push(token(NUMBER, l.loc))
	} else {
		l.advance(1)
		// endToken(l)
		l.push(token(DOT, l.loc))
	}
}

func (l *Lexer) handleLabel() {
	startToken(l)
	l.advance(1) // consume @
	for l.lex() && lib.IsAlpha(l.char()) {
		l.advance(1)
	}
	// endToken(l)
	l.push(token(LABEL, l.loc))
}

func (l *Lexer) handleIdentifier() {
	startToken(l)
	for l.lex() && lib.IsAlpha(l.char()) {
		l.advance(1)
	}

	tk := token(IDENTIFIER, l.loc)
	if t, exists := keywords[l.src(tk)]; exists {
		tk.tag = t
	}
	// endToken(l)
	l.push(tk)
}

func (l *Lexer) handleQuote() {
	startToken(l)
	del := l.char()
	l.advance(1) // consume opening delimiter
	for l.lex() {
		if l.char() == del {
			l.advance(1) // consume closing delimiter
			// endToken(l)
			l.push(token(STRING, l.loc))
			return
		}
		if l.char() == '\\' {
			// escape sequence
			l.advance(1)
			if !l.lex() {
				l.lexError(SyntaxError, UnexpectedEOT)
			}
			l.advance(1)
		} else if l.char() == '\n' {
			l.lexError(SyntaxError, UnclosedString)
		} else {
			l.advance(1)
		}
	}
	// EOF reached without closing delimiter
	l.lexError(SyntaxError, UnclosedString)
}

// ====================== NUMBER PARSING ======================

func (l *Lexer) handleNumber() {
	startToken(l)
	if l.char() == '0' {
		l.advance(1)
		switch l.char() {
		case 'x', 'X':
			if lib.IsHex(l.charAt(1)) {
				l.advance(1)
				l.handleHex()
				return
			}
		case 'o', 'O':
			if lib.IsOctal(l.charAt(1)) {
				l.advance(1)
				l.handleOctal()
				return
			}
		case 'b', 'B':
			if lib.IsBinary(l.charAt(1)) {
				l.advance(1)
				l.handleBinary()
				return
			}
			return
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '_':
			l.lexError(SyntaxError, "Octal literals are not allowed. Use the syntax '0o123'.")
			return
		default:
			// Just "0", or "0" followed by invalid char
			// endToken(l)
			l.push(token(NUMBER, l.loc))
			return
		}
	}

	// Normal decimal integer starting with 1-9
	l.handleInt()
}

func (l *Lexer) handleInt() {
loop:
	for l.lex() {
		switch l.char() {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			l.advance(1)
		case '_':
			l.handleNumSep("dec")
			return
		case '.':
			l.advance(1)
			l.handleFloat()
			return
		default:
			break loop
		}
	}
	// endToken(l)
	l.push(token(NUMBER, l.loc))
}

func (l *Lexer) handleHex() {
	for l.lex() {
		ch := l.char()
		switch {
		case lib.IsHex(ch):
			l.advance(1)
		case ch == '_':
			l.handleNumSep("hex")
			return
		default:
			// endToken(l)
			l.push(token(NUMBER, l.loc))
			return
		}
	}
	// endToken(l)
	l.push(token(NUMBER, l.loc))
}

func (l *Lexer) handleOctal() {
	for l.lex() {
		ch := l.char()
		switch {
		case lib.IsOctal(ch):
			l.advance(1)
		case ch == '_':
			l.handleNumSep("oct")
			return
		default:
			// endToken(l)
			l.push(token(NUMBER, l.loc))
			return
		}
	}
	// endToken(l)
	l.push(token(NUMBER, l.loc))
}

func (l *Lexer) handleBinary() {
	for l.lex() {
		ch := l.char()
		switch {
		case lib.IsBinary(ch):
			l.advance(1)
		case ch == '_':
			l.handleNumSep("bin")
			return
		default:
			// endToken(l)
			l.push(token(NUMBER, l.loc))
			return
		}
	}
	// endToken(l)
	l.push(token(NUMBER, l.loc))
}

func (l *Lexer) handleNumSep(num string) {
	l.advance(1) // consume the _

	switch num {
	case "hex":
		if !lib.IsHex(l.char()) {
			l.lexError(SyntaxError, _InNumberMustSeparateSuccessiveDigits)
		}
		l.handleHex()
	case "dec":
		if !lib.IsDigit(l.char()) {
			l.lexError(SyntaxError, _InNumberMustSeparateSuccessiveDigits)
		}
		l.handleInt()
	case "oct":
		if !lib.IsOctal(l.char()) {
			l.lexError(SyntaxError, _InNumberMustSeparateSuccessiveDigits)
		}
		l.handleOctal()
	case "bin":
		if !lib.IsBinary(l.char()) {
			l.lexError(SyntaxError, _InNumberMustSeparateSuccessiveDigits)
		}
		l.handleBinary()
	default:
		lib.Panic("unhandled case: " + num)
	}
}

// pushes token when done.
func (l *Lexer) handleFloat() {
loop:
	for {
		if lib.IsDigit(l.char()) {
			l.advance(1)
		} else {
			break loop
		}
	}
	// endToken(l)
	l.push(token(NUMBER, l.loc))
}

type ErrorName int

const (
	UnrecognisedChar ErrorName = iota
	SyntaxError
	UnexpectedToken
	ExpectedAToken
	PathError
)

func (l *Lexer) lexError(errname ErrorName, additionals ...string) {
	name := "SyntaxError"
	message := ""
	b := lib.NewStringBuilder()
	b.WriteString(lib.Sprintf("%s: ", lib.Red(string(name))))
	switch errname {
	case UnrecognisedChar:
		message = lib.Sprintf("Unrecognised character found in source: %d%s", l.char(), lib.SourceLog(l.path, l.Loc()))
	case SyntaxError:
		message = lib.Sprintf("%s%s", additionals[0], lib.SourceLog(l.path, l.loc))
		additionals = additionals[1:]
	default:
		lib.Panic("unhandled error name")
	}
	b.WriteString(message)
	b.WriteString(lib.EOL)
	for i, m := range additionals {
		b.WriteString(m)
		if i == len(additionals)-1 && additionals[i][len(additionals[i])-1] != '\n' {
			b.WriteString(lib.EOL)
		}
	}
	lib.Print(b.String())
	lib.ExitWith1()
}
