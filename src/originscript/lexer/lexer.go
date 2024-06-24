package lexer

import (
	"bytes"
	"errors"
	"fmt"
	"originscript/token"
	"strings"
	"unicode"
	"unicode/utf8"
)

const bom = 0xFEFF // byte order mark, only permitted as very first character
var prevToken token.Token

var macroMap = map[string]token.TokenType{
	"#define": token.DEFINE,
	"#ifdef":  token.IFDEF_MACRO,
	"#else":   token.ELSE_MACRO,
}

// A mode value is a set of flags (or 0).
// They control scanner behavior.
//
type Mode uint

const (
	ScanComments Mode = 1 << iota // return comments as COMMENT tokens

	//Here I removed the '<' from the 'op_chars', why? See below example
	// 1   f = open("./a.txt", "r")
	// 2   line=<$f>
	// 3   f.close()
	//In line 2, the lexer found '=<', so the lexer treats it as an 'UDO'(User Defined Operator) token,
	//but what we want is to read one line from the file 'f'.
	op_chars = ".=+*/%&,|^~,>},?@#$"
)

type Lexer struct {
	filename     string
	input        []rune
	ch           rune //current character
	position     int  //character offset
	readPosition int  //reading offset

	line int
	col  int

	Mode Mode // scanning mode
}

func New(filename, input string) *Lexer {
	l := &Lexer{filename: filename, input: []rune(input)}
	l.ch = ' '
	l.position = 0
	l.readPosition = 0
	l.line = 1
	l.col = 1

	l.readNext()
	if l.ch == bom {
		l.readNext() //ignore BOM at file beginning
	}

	return l
}

func (l *Lexer) SetMode(mode Mode) {
	l.Mode = mode
}

func (l *Lexer) readNext() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
		if l.ch == '\n' {
			l.col = 1
			l.line++
		} else {
			l.col += 1
		}
	}

	l.position = l.readPosition
	l.readPosition++
}

func (l *Lexer) peek() rune {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func (l *Lexer) peekn(n int) rune {
	if l.readPosition+n >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition+n]
}

var tokenMap = map[rune]token.TokenType{
	'=': token.ASSIGN,
	'.': token.DOT,
	';': token.SEMICOLON,
	'(': token.LPAREN,
	')': token.RPAREN,
	'{': token.LBRACE,
	'}': token.RBRACE,
	'[': token.LBRACKET,
	']': token.RBRACKET,
	'+': token.PLUS,
	',': token.COMMA,
	'-': token.MINUS,
	'!': token.BANG,
	'*': token.ASTERISK,
	'/': token.SLASH,
	'<': token.LT,
	'>': token.GT,
	':': token.COLON,
	'%': token.MOD,
	'#': token.COMMENT,
	'?': token.QUESTIONM,
	'&': token.BITAND,
	'|': token.BITOR,
	'^': token.BITXOR,
	'@': token.AT,
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token
	l.skipWhitespace()
	pos := l.getPos()
	pos.Col -= 1

	if t, ok := tokenMap[l.ch]; ok {
		switch t {
		case token.ASSIGN:
			if l.peek() == '=' {
				tok = token.Token{Type: token.EQ, Literal: string(l.ch) + string(l.peek())}
				l.readNext()
			} else if l.peek() == '~' {
				tok = token.Token{Type: token.MATCH, Literal: string(l.ch) + string(l.peek())}
				l.readNext()
			} else if l.peek() == '>' {
				tok = token.Token{Type: token.FATARROW, Literal: string(l.ch) + string(l.peek())}
				l.readNext()
			} else if strings.Contains(op_chars, string(l.peek())) {
				//User Defined Operator
				tok = token.Token{Type: token.UDO, Literal: string(l.ch) + string(l.peek())}
				l.readNext()
			} else {
				tok = newToken(token.ASSIGN, l.ch)
			}
		case token.PLUS:
			if l.peek() == '+' {
				tok = token.Token{Type: token.INCREMENT, Literal: string(l.ch) + string(l.peek())}
				l.readNext()
			} else if l.peek() == '=' {
				tok = token.Token{Type: token.PLUS_A, Literal: string(l.ch) + string(l.peek())}
				l.readNext()
			} else if strings.Contains(op_chars, string(l.peek())) {
				//User Defined Operator
				tok = token.Token{Type: token.UDO, Literal: string(l.ch) + string(l.peek())}
				l.readNext()
			} else {
				tok = newToken(token.PLUS, l.ch)
			}
		case token.MINUS:
			if l.peek() == '-' {
				tok = token.Token{Type: token.DECREMENT, Literal: string(l.ch) + string(l.peek())}
				l.readNext()
			} else if l.peek() == '=' {
				tok = token.Token{Type: token.MINUS_A, Literal: string(l.ch) + string(l.peek())}
				l.readNext()
			} else if l.peek() == '>' {
				tok = token.Token{Type: token.THINARROW, Literal: string(l.ch) + string(l.peek())}
				l.readNext()
			} else if strings.Contains(op_chars, string(l.peek())) {
				//User Defined Operator
				tok = token.Token{Type: token.UDO, Literal: string(l.ch) + string(l.peek())}
				l.readNext()
			} else {
				tok = newToken(token.MINUS, l.ch)
			}
		case token.GT:
			if l.peek() == '=' {
				tok = token.Token{Type: token.GE, Literal: string(l.ch) + string(l.peek())}
				l.readNext()
			} else if l.peek() == '>' {
				tok = token.Token{Type: token.SHIFT_R, Literal: string(l.ch) + string(l.peek())}
				l.readNext()
			} else if strings.Contains(op_chars, string(l.peek())) {
				//User Defined Operator
				tok = token.Token{Type: token.UDO, Literal: string(l.ch) + string(l.peek())}
				l.readNext()
			} else {
				tok = newToken(token.GT, l.ch)
			}
		case token.LT:
			if l.peek() == '=' {
				tok = token.Token{Type: token.LE, Literal: string(l.ch) + string(l.peek())}
				l.readNext()
			} else if l.peek() == '<' {
				tok = token.Token{Type: token.SHIFT_L, Literal: string(l.ch) + string(l.peek())}
				l.readNext()
			} else if l.peek() == '$' {
				if s, err := l.readDiamond(); err == nil {
					tok = token.Token{Pos: pos, Type: token.LD, Literal: s}
					return tok
				}
			} else if strings.Contains(op_chars, string(l.peek())) {
				//User Defined Operator
				tok = token.Token{Type: token.UDO, Literal: string(l.ch) + string(l.peek())}
				l.readNext()
			} else {
				tok = newToken(token.LT, l.ch)
			}
		case token.BANG:
			if l.peek() == '=' {
				tok = token.Token{Type: token.NEQ, Literal: string(l.ch) + string(l.peek())}
				l.readNext()
			} else if l.peek() == '~' {
				tok = token.Token{Type: token.NOTMATCH, Literal: string(l.ch) + string(l.peek())}
				l.readNext()
			} else if strings.Contains(op_chars, string(l.peek())) {
				//User Defined Operator
				tok = token.Token{Type: token.UDO, Literal: string(l.ch) + string(l.peek())}
				l.readNext()
			} else {
				tok = newToken(token.BANG, l.ch)
			}
		case token.SLASH:
			if l.peek() == '/' || l.peek() == '*' {
				var comment string
				if l.peek() == '/' {
					comment, _ = l.readComment()
				} else {
					comment, _ = l.readMultilineComment()
				}
				if l.Mode&ScanComments == 0 { //skip comment
					return l.NextToken()
				}

				tok.Pos = pos
				tok.Type = token.COMMENT
				tok.Literal = comment
				prevToken = tok
				return tok
			} else {
				if prevToken.Type == token.RBRACE || // impossible?
					prevToken.Type == token.RPAREN || // (a+c) / b
					prevToken.Type == token.RBRACKET || // a[3] / b
					prevToken.Type == token.IDENT || // a / b
					prevToken.Type == token.INT || // 3 / b
					prevToken.Type == token.FLOAT || // 3.5 / b
					prevToken.Type == token.FUNCTION { // e.g. fn /() - operator overloading
					if l.peek() == '=' {
						tok = token.Token{Type: token.SLASH_A, Literal: string(l.ch) + string(l.peek())}
						l.readNext()
					} else {
						tok = newToken(token.SLASH, l.ch)
					}
				} else { //regexp
					tok.Literal = l.readRegExLiteral()
					tok.Type = token.REGEX
					tok.Pos = pos
					return tok
				}
			}
		case token.MOD:
			if l.peek() == '=' {
				tok = token.Token{Type: token.MOD_A, Literal: string(l.ch) + string(l.peek())}
				l.readNext()
			} else {
				tok = newToken(token.MOD, l.ch)
			}
		case token.ASTERISK:
			if l.peek() == '=' {
				tok = token.Token{Type: token.ASTERISK_A, Literal: string(l.ch) + string(l.peek())}
				l.readNext()
			} else if l.peek() == '*' {
				tok = token.Token{Type: token.POWER, Literal: string(l.ch) + string(l.peek())}
				l.readNext()
			} else if strings.Contains(op_chars, string(l.peek())) {
				//User Defined Operator
				tok = token.Token{Type: token.UDO, Literal: string(l.ch) + string(l.peek())}
				l.readNext()
			} else {
				tok = newToken(token.ASTERISK, l.ch)
			}
		case token.DOT:
			if l.peek() == '.' {
				l.readNext()
				if l.peek() == '.' {
					tok = token.Token{Type: token.ELLIPSIS, Literal: "..."}
					l.readNext()
				} else {
					tok = token.Token{Type: token.DOTDOT, Literal: ".."}
				}
			} else if strings.Contains(op_chars, string(l.peek())) {
				//User Defined Operator
				tok = token.Token{Type: token.UDO, Literal: string(l.ch) + string(l.peek())}
				l.readNext()
			} else {
				tok = newToken(token.DOT, l.ch)
			}
		case token.COMMENT:
			comment, isMacro := l.readComment()
			if isMacro == 0 {
				if l.Mode&ScanComments == 0 { //skip comment
					return l.NextToken()
				}
				tok.Pos = pos
				tok.Type = token.COMMENT
				tok.Literal = comment
				prevToken = tok
				return tok
			} else {
				tok.Pos = pos
				tok.Type = macroMap[comment]
				tok.Literal = comment
				prevToken = tok
				return tok
			}

		case token.BITAND:
			if l.peek() == '=' {
				tok = token.Token{Type: token.BITAND_A, Literal: string(l.ch) + string(l.peek())}
				l.readNext()
			} else if l.peek() == '&' {
				tok = token.Token{Type: token.CONDAND, Literal: string(l.ch) + string(l.peek())}
				l.readNext()
			} else if strings.Contains(op_chars, string(l.peek())) {
				//User Defined Operator
				tok = token.Token{Type: token.UDO, Literal: string(l.ch) + string(l.peek())}
				l.readNext()
			} else {
				tok = newToken(token.BITAND, l.ch)
			}
		case token.BITOR:
			if l.peek() == '=' {
				tok = token.Token{Type: token.BITOR_A, Literal: string(l.ch) + string(l.peek())}
				l.readNext()
			} else if l.peek() == '>' {
				tok = token.Token{Type: token.PIPE, Literal: string(l.ch) + string(l.peek())}
				l.readNext()
			} else if l.peek() == '|' {
				tok = token.Token{Type: token.CONDOR, Literal: string(l.ch) + string(l.peek())}
				l.readNext()
			} else if strings.Contains(op_chars, string(l.peek())) {
				//User Defined Operator
				tok = token.Token{Type: token.UDO, Literal: string(l.ch) + string(l.peek())}
				l.readNext()
			} else {
				tok = newToken(token.BITOR, l.ch)
			}
		case token.BITXOR:
			if l.peek() == '=' {
				tok = token.Token{Type: token.BITXOR_A, Literal: string(l.ch) + string(l.peek())}
				l.readNext()
			} else if strings.Contains(op_chars, string(l.peek())) {
				//User Defined Operator
				tok = token.Token{Type: token.UDO, Literal: string(l.ch) + string(l.peek())}
				l.readNext()
			} else {
				tok = newToken(token.BITXOR, l.ch)
			}
		case token.QUESTIONM:
			if l.peek() == '?' {
				tok = token.Token{Type: token.QUESTIONMM, Literal: string(l.ch) + string(l.peek())}
				l.readNext()
			} else {
				tok = newToken(token.QUESTIONM, l.ch)
			}
		default:
			tok = newToken(t, l.ch)
		}

		l.readNext()

		tok.Pos = pos
		prevToken = tok
		return tok
	}

	newTok := l.readRunesToken()
	newTok.Pos = pos
	prevToken = newTok
	return newTok
}

func (l *Lexer) readRunesToken() token.Token {
	var tok token.Token
	switch {
	case l.ch == 0:
		tok.Literal = "<EOF>"
		tok.Type = token.EOF
		return tok
	case isLetter(l.ch):
		if l.ch == '_' && !isLetter(l.peek()) {
			l.readNext()
			return newToken(token.UNDERSCORE, '_')
		} else {
			return l.readIdentifier()
		}
	case isDigit(l.ch):
		literal, isUnsigned, _ := l.readNumber()
		if strings.Contains(literal, ".") {
			tok.Type = token.FLOAT
		} else {
			if isUnsigned {
				tok.Type = token.UINT
			} else {
				tok.Type = token.INT
			}
		}
		tok.Literal = literal
		return tok
	case isQuote(l.ch):
		if l.ch == 34 { //double quotes
			if s, err := l.readString(l.ch); err == nil {
				tok.Type = token.STRING
				tok.Literal = s
				return tok
			}
		} else if l.ch == 96 {
			if l.peek() == 96 { //raw string
				if s, err := l.readRawString(); err == nil {
					tok.Type = token.STRING
					tok.Literal = s
					return tok
				}
			} else { // it's a command
				if s, err := l.readCommand(l.ch); err == nil {
					tok.Type = token.CMD
					tok.Literal = s
					return tok
				}
			}
		}
	case isSingleQuote(l.ch):
		if s, err := l.readInterpString(l.ch); err == nil {
			tok.Type = token.ISTRING
			tok.Literal = s
			return tok
		}
	case l.ch == '~':
		return l.getMetaOperatorToken()
	}
	l.readNext()
	return newToken(token.ILLEGAL, l.ch)
}

func newToken(tokenType token.TokenType, ch rune) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}

func (l *Lexer) readRawString() (string, error) {
	var ret []rune
	l.readNext()
	for {
		l.readNext()
		if l.ch == 0 {
			return "", errors.New("unexpected EOF")
		}

		if l.ch == 96 && l.peek() == 96 {
			l.readNext()
			l.readNext()
			break
		}
		ret = append(ret, l.ch)
	}
	return string(ret), nil
}

func (l *Lexer) readString(r rune) (string, error) {
	var ret []rune
eos:
	for {
		l.readNext()
		switch l.ch {
		case '\n':
			return "", errors.New("unexpected EOL")
		case 0:
			return "", errors.New("unexpected EOF")
		case r:
			l.readNext()
			break eos //eos:end of string
		case '\\':
			l.readNext()
			switch l.ch {
			case 'b':
				ret = append(ret, '\b')
				continue
			case 'f':
				ret = append(ret, '\f')
				continue
			case 'r':
				ret = append(ret, '\r')
				continue
			case 'n':
				ret = append(ret, '\n')
				continue
			case 't':
				ret = append(ret, '\t')
				continue
			}
			ret = append(ret, l.ch)
			continue
		default:
			ret = append(ret, l.ch)
		}
	}

	return string(ret), nil
}

func (l *Lexer) readInterpString(r rune) (string, error) {
	start := l.position + 1
	newStart := start
	var out bytes.Buffer
	pos := "0"[0]
	for {
		l.readNext()
		if l.ch == r {
			l.readNext()
			break
		}
		if l.ch == 0 {
			err := errors.New("")
			return "", err
		}
		if l.ch == 123 {
			newStart = l.position
			if l.peek() != 125 {
				out.WriteRune(l.ch)
				for l.ch != 125 || l.ch == 0 {
					l.readNext()
				}
				if l.ch != 0 {
					out.WriteRune(rune(pos))
					pos++
				}
			}
		}
		out.WriteRune(l.ch)
	}
	l.position = start - 1
	l.readPosition = start
	l.ch = l.input[newStart]
	return out.String(), nil
}

func (l *Lexer) NextInterpToken() token.Token {
	var tok token.Token
	for {
		if l.ch == '{' {
			if l.peek() == '}' {
				continue
			}
			tok = newToken(token.LBRACE, l.ch)
			break
		}
		if l.ch == 0 {
			tok.Type = token.EOF
			tok.Literal = ""
			break
		}
		if isSingleQuote(l.ch) {
			tok = newToken(token.ISTRING, l.ch)
			break
		}
		l.readNext()
	}
	l.readNext()
	tok.Pos = l.getPos()
	return tok
}

// for date-time literal handling
func (l *Lexer) NextInterpToken2() token.Token {
	var tok token.Token
	for {
		if l.ch == '{' {
			if l.peek() == '}' {
				continue
			}
			tok = newToken(token.LBRACE, l.ch)
			break
		}
		if l.ch == 0 {
			tok.Type = token.EOF
			tok.Literal = ""
			break
		}
		if l.ch == '/' {
			tok = newToken(token.DATETIME, l.ch)
			break
		}
		l.readNext()
	}
	l.readNext()
	tok.Pos = l.getPos()
	return tok
}

func (l *Lexer) readCommand(r rune) (string, error) {
	var ret []rune
eoc:
	for {
		l.readNext()
		switch l.ch {
		case '\r':
		case '\n':
			//swallow it
			continue
		case 0:
			return "", errors.New("unexpected EOF")
		case r:
			l.readNext()
			break eoc //eoc:end of command
		case '\\':
			l.readNext()
			switch l.ch {
			case '$':
				ret = append(ret, '\\')
				ret = append(ret, '$')
				continue
			case '`':
				ret = append(ret, '`')
				continue
			}
			ret = append(ret, l.ch)
			continue
		default:
			ret = append(ret, l.ch)
		}
	}

	return string(ret), nil
}

func (l *Lexer) readIdentifier() token.Token {
	// 	defer func() {
	//		if err := recover(); err != nil {
	// 			fmt.Fprintf(os.Stderr, "\x1b[31m%s\x1b[0m\n", err)
	// 		}
	//    }()

	position := l.position
	if l.ch == 'd' {
		if l.peek() == 't' { //It's a datetime token
			l.readNext()
			if l.peek() == '/' {
				l.readNext()
				if s, err := l.readInterpString(l.ch); err == nil {
					return token.Token{Type: token.DATETIME, Literal: s}
				}
			}
		}
	}

	// Why '?' : Because Magpie support Optional, so it should be good for
	// a Optional type to denote it meaning with a '?' like 'isEmpty?'

	// Why '$' : Because Magpie support extend built-in types with 'integer', 'float', etc.
	// For example, you could extend 'integer' type with 'integer$funcname(xxx)'
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '?' || l.ch == '$' {
		l.readNext()
	}

	//if identifier contains '?', then it must be the last one
	ret := string(l.input[position:l.position])

	cnt := strings.Count(ret, "?")
	if cnt > 1 { //multiple '?'
		errStr := fmt.Sprintf("Line[%d]: Identifier() could only contain one '?' character", l.line, ret)
		panic(errStr)
	} else if cnt == 1 { //only one '?'
		if ret[len(ret)-1:] != "?" {
			errStr := fmt.Sprintf("Line[%d]: '?' character must be the last character in '%s' identifier", l.line, ret)
			panic(errStr)
		}
	}

	var tok token.Token
	tok.Literal = ret
	tok.Type = token.LookupIdent(tok.Literal)
	return tok
}

func (l *Lexer) readRegExLiteral() (literal string) {
	position := l.position
	/* read until closing slash */
	for {
		l.readNext()
		if l.ch == '\\' {
			// Skip escape sequence
			l.readNext()
		} else if l.ch == '/' {
			// This is the closing
			literal = string(l.input[position+1 : l.position])
			l.readNext() //skip the '/'

			return
		}
	}
}

func isLetter(ch rune) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_' || ch == '$' || ch == '@' || ch >= utf8.RuneSelf && unicode.IsLetter(ch)
}

// scanNumber returns number begining at current position.
func (l *Lexer) readNumber() (string, bool, error) {
	var isUnsigned bool
	var ret []rune
	ch := l.ch
	ret = append(ret, ch)
	l.readNext()

	if ch == '0' && (l.ch == 'x' || l.ch == 'b' || l.ch == 'o') { //support '0x'(hex) and '0b'(bin) and '0o'(octal)
		savedCh := l.ch
		ret = append(ret, l.ch)
		l.readNext()
		if savedCh == 'x' {
			for isHex(l.ch) || l.ch == '_' {
				if l.ch == '_' {
					l.readNext()
					continue
				}
				ret = append(ret, l.ch)
				l.readNext()
			}
		} else if savedCh == 'b' {
			for isBin(l.ch) || l.ch == '_' {
				if l.ch == '_' {
					l.readNext()
					continue
				}
				ret = append(ret, l.ch)
				l.readNext()
			}
		} else if savedCh == 'o' {
			for isOct(l.ch) || l.ch == '_' {
				if l.ch == '_' {
					l.readNext()
					continue
				}
				ret = append(ret, l.ch)
				l.readNext()
			}
		}

		if l.ch == 'u' {
			isUnsigned = true
			l.readNext()
		}
	} else {
		for isDigit(l.ch) || l.ch == '.' || l.ch == '_' {
			if l.ch == '_' {
				l.readNext()
				continue
			}

			if l.ch == '.' {
				if l.peek() == '.' { //range operator
					return string(ret), false, nil
				} else if !isDigit(l.peek()) && l.peek() != 'e' && l.peek() != 'E' { //should be a method calling, e.g. 10.next()
					return string(ret), false, nil
				}
			} //end if

			ret = append(ret, l.ch)
			l.readNext()
		}

		if l.ch == 'e' || l.ch == 'E' {
			ret = append(ret, l.ch)
			l.readNext()
			if isDigit(l.ch) || l.ch == '+' || l.ch == '-' {
				ret = append(ret, l.ch)
				l.readNext()
				for isDigit(l.ch) || l.ch == '.' || l.ch == '_' {
					if l.ch == '_' {
						l.readNext()
						continue
					}
					ret = append(ret, l.ch)
					l.readNext()
				}
			}
			for isDigit(l.ch) || l.ch == '.' || l.ch == '_' {
				if l.ch == '_' {
					l.readNext()
					continue
				}
				ret = append(ret, l.ch)
				l.readNext()
			}
		} else if l.ch == 'u' {
			isUnsigned = true
			l.readNext()
		}
		//		if isLetter(l.ch) {
		//			return "", errors.New("identifier starts immediately after numeric literal")
		//		}
	}

	if isUnsigned {
		return string(ret), true, nil
	}
	return string(ret), false, nil
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9' || ch >= utf8.RuneSelf && unicode.IsDigit(ch)
}

// isHex returns true if the rune is a hex digits.
func isHex(ch rune) bool {
	return ('0' <= ch && ch <= '9') || ('a' <= ch && ch <= 'f') || ('A' <= ch && ch <= 'F')
}

// isBin returns true if the rune is a binary digits.
func isBin(ch rune) bool {
	return ('0' == ch || '1' == ch)
}

// isOct returns true if the rune is a octal digits.
func isOct(ch rune) bool {
	return ('0' <= ch && ch <= '7')
}

func isQuote(ch rune) bool {
	return ch == 34 || ch == 96
}

func isSingleQuote(ch rune) bool {
	return ch == 39
}

func (l *Lexer) skipWhitespace() {
	for unicode.IsSpace(l.ch) {
		l.readNext()
	}
}

func (l *Lexer) readDiamond() (string, error) {
	var ret []rune
	l.readNext()
	for {
		l.readNext()
		if l.ch == 0 {
			return "", errors.New("unexpected EOF")
		}

		if l.ch == '>' {
			l.readNext()
			break
		}
		ret = append(ret, l.ch)
	}

	return string(ret), nil
}

func (l *Lexer) readComment() (string, int) {
	position := l.position
	for l.ch != '\n' && l.ch != 0 {
		l.readNext()

		tmp := string(l.input[position:l.position])
		if len(tmp) >= 5 {
			if _, ok := macroMap[tmp]; ok {
				if unicode.IsSpace(l.ch) || l.ch == '\n' {
					return tmp, 1
				}
			}
		}
	}
	return string(l.input[position:l.position]), 0
}

func (l *Lexer) readMultilineComment() (string, error) {
	var err error
	position := l.position
loop:

	for {
		l.readNext()
		switch l.ch {
		case '*':
			switch l.peek() {
			case '/': // got the block ending symbol: '*/'
				l.readNext() //skip the '*'
				l.readNext() //skip the '/'
				break loop
			}
		case 0: // Got EOF, but not comment terminator.
			err = errors.New("Unterminated multiline comment, GOT EOF!")
			break loop
		}
	}

	return string(l.input[position:l.position]), err
}

func (l *Lexer) getMetaOperatorToken() token.Token {
	ch := l.ch // ~
	var tok token.Token
	switch l.peek() {
	case '+':
		tok = token.Token{Type: token.TILDEPLUS, Literal: string(ch) + string(l.peek())}
		l.readNext()
	case '-':
		tok = token.Token{Type: token.TILDEMINUS, Literal: string(ch) + string(l.peek())}
		l.readNext()
	case '*':
		tok = token.Token{Type: token.TILDEASTERISK, Literal: string(ch) + string(l.peek())}
		l.readNext()
	case '/':
		tok = token.Token{Type: token.TILDESLASH, Literal: string(ch) + string(l.peek())}
		l.readNext()
	case '%':
		tok = token.Token{Type: token.TILDEMOD, Literal: string(ch) + string(l.peek())}
		l.readNext()
	case '^':
		tok = token.Token{Type: token.TILDECARET, Literal: string(ch) + string(l.peek())}
		l.readNext()
	} //end switch

	l.readNext()
	return tok
}

func (l *Lexer) getPos() token.Position {
	return token.Position{
		Filename: l.filename,
		Offset:   l.position,
		Line:     l.line,
		Col:      l.col,
	}
}
