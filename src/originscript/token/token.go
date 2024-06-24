package token

import "fmt"

type TokenType int

const (
	ILLEGAL TokenType = (iota - 1) // Illegal token
	EOF

	IDENT    //identifier
	INT      //int literal
	UINT     //unsigned int
	FLOAT    //float literal
	DATETIME //datetime

	EQ         // ==
	NEQ        // !=
	MATCH      // =~
	NOTMATCH   // !~
	ASSIGN     // =
	PLUS       // +
	PLUS_A     // += (PLUS ASSIGN)
	MINUS      // -
	MINUS_A    // -= (MINUS ASSIGN)
	BANG       // !
	ASTERISK   // *
	ASTERISK_A // *= (ASTERISK ASSIGN)
	SLASH      // '/' divide
	SLASH_A    // /= (SLASH ASSIGN)
	MOD        // %
	MOD_A      // %= //MOD ASSIGN
	POWER      // ** (POWER)
	QUESTIONM  // ? (QUESTION MARK)

	LT        // <
	LE        // <=
	SHIFT_L   // << (SHIFT LEFT)
	LD        // <$
	GT        // >
	GE        // >=
	SHIFT_R   // >> (SHIFT RIGHT)
	COMMA     // ,
	SEMICOLON // ;

	LPAREN    // (
	RPAREN    // )
	LBRACE    // {
	RBRACE    // }
	LBRACKET  // [
	RBRACKET  // ]
	COLON     // :
	DOT       // .
	DOTDOT    // ..  (PARTIAL IMPLEMENTED, ONLY SUPPORT INTEGER/SingleString RANGE, AND ONLY USED IN 'FOR X IN A..B {}' )
	ELLIPSIS  //... Function Variadic parameters
	PIPE      // |>
	THINARROW // ->
	FATARROW  // =>

	INCREMENT // ++
	DECREMENT // --

	BITAND   // &
	BITOR    // |
	BITXOR   // ^
	BITAND_A // &=
	BITOR_A  // |=
	BITXOR_A // ^=

	CONDAND // &&
	CONDOR  // ||

	AT  //@
	CMD // `

	FUNCTION
	LET
	CONST
	TRUE
	FALSE
	IF
	ELIF
	ELSE
	RETURN
	IMPORT
	STRING
	ISTRING
	BYTES
	AND
	OR
	STRUCT
	DO
	WHILE
	BREAK
	CONTINUE

	COMMENT // '#' or '//' (single-line comment), '/* */' multiline comment
	REGEX   // REGEX
	FOR
	IN
	WHERE
	GREP
	MAP
	CASE
	IS
	TRY
	CATCH
	FINALLY
	THROW
	DEFER
	SPAWN
	NIL
	ENUM
	QW
	UNLESS

	//class related
	INTERFACE //NOT IMPLEMENTED
	CLASS
	NEW
	PROPERTY
	GET
	SET
	PUBLIC    //NOT IMPLEMENTED
	PRIVATE   //NOT IMPLEMENTED
	PROTECTED //NOT IMPLEMENTED
	STATIC
	DEFAULT

	// User Defined Operator
	UDO
	UNDERSCORE // _(PlaceHolder)

	//Meta-Operators(for working with array)
	TILDEPLUS     // ~+
	TILDEMINUS    // ~-
	TILDEASTERISK // ~*
	TILDESLASH    // ~/
	TILDEMOD      // ~%
	TILDECARET    // ~^

	USING
	QUESTIONMM // ?? (Null Coalescing Operator)

	//linq query
	FROM
	//WHERE
	SELECT
	GROUP
	INTO
	ORDERBY
	JOIN
	//LET
	//IN
	ON
	EQUALS
	BY
	ASCENDING
	DESCENDING

	//async & await
	ASYNC
	AWAIT

	//service
	SERVICE

	//macors
	DEFINE
	IFDEF_MACRO
	ELSE_MACRO
)

var keywords = map[string]TokenType{
	"function":  FUNCTION,
	"fn":        FUNCTION,
	"lit":       LET,
	"const":     CONST,
	"true":      TRUE,
	"false":     FALSE,
	"if":        IF,
	"elif":      ELIF,
	"else":      ELSE,
	"return":    RETURN,
	"require":   IMPORT,
	"and":       AND,
	"or":        OR,
	"struct":    STRUCT,
	"do":        DO,
	"while":     WHILE,
	"break":     BREAK,
	"continue":  CONTINUE,
	"for":       FOR,
	"in":        IN,
	"where":     WHERE,
	"gp":        GREP,
	"map":       MAP,
	"case":      CASE,
	"is":        IS,
	"try":       TRY,
	"catch":     CATCH,
	"finally":   FINALLY,
	"throw":     THROW,
	"defer":     DEFER,
	"spawn":     SPAWN,
	"nil":       NIL,
	"enum":      ENUM,
	"quo":       QW, //“quoted words”
	"unless":    UNLESS,
	"interface": INTERFACE,
	"class":     CLASS,
	"new":       NEW,
	"property":  PROPERTY,
	"get":       GET,
	"set":       SET,
	"public":    PUBLIC,
	"private":   PRIVATE,
	"protected": PROTECTED,
	"static":    STATIC,
	"default":   DEFAULT,
	"using":     USING,

	//linq query
	"from": FROM,
	//"where":      WHERE,
	"select":  SELECT,
	"group":   GROUP,
	"into":    INTO,
	"orderby": ORDERBY,
	"join":    JOIN,
	//"let":      LET,
	//"in":       IN,
	"on":         ON,
	"equals":     EQUALS,
	"by":         BY,
	"ascending":  ASCENDING,
	"descending": DESCENDING,

	//async & await
	"async": ASYNC,
	"await": AWAIT,

	//service
	"service": SERVICE,
}

//for debug & testing
func (tt TokenType) String() string {
	switch tt {
	case EOF:
		return "EOF"
	case IDENT:
		return "IDENT"
	case INT:
		return "INT"
	case UINT:
		return "UINT"
	case FLOAT:
		return "FLOAT"
	case DATETIME:
		return "DATETIME"
	case EQ:
		return "=="
	case NEQ:
		return "!="
	case MATCH:
		return "=~"
	case NOTMATCH:
		return "!~"
	case ASSIGN:
		return "="
	case PLUS:
		return "+"
	case PLUS_A:
		return "+="
	case MINUS:
		return "-"
	case MINUS_A:
		return "-="
	case BANG:
		return "!"
	case ASTERISK:
		return "*"
	case ASTERISK_A:
		return "*="
	case SLASH:
		return "/"
	case SLASH_A:
		return "/="
	case MOD:
		return "%"
	case MOD_A:
		return "%="
	case POWER:
		return "**"
	case LT:
		return "<"
	case LE:
		return "<="
	case SHIFT_L:
		return "<<"
	case LD:
		return "<$>"
	case GT:
		return ">"
	case GE:
		return ">="
	case SHIFT_R:
		return ">>"
	case COMMA:
		return ","
	case SEMICOLON:
		return ";"
	case LPAREN:
		return "("
	case RPAREN:
		return ")"
	case LBRACE:
		return "{"
	case RBRACE:
		return "}"
	case LBRACKET:
		return "["
	case RBRACKET:
		return "]"
	case COLON:
		return ":"
	case DOT:
		return "."
	case DOTDOT:
		return ".."
	case ELLIPSIS:
		return "..."
	case PIPE:
		return "|>"
	case THINARROW:
		return "->"
	case FATARROW:
		return "=>"
	case INCREMENT:
		return "++"
	case DECREMENT:
		return "--"
	case BITAND:
		return "&"
	case BITOR:
		return "|"
	case BITXOR:
		return "^"
	case BITAND_A:
		return "&="
	case BITOR_A:
		return "|="
	case BITXOR_A:
		return "^="
	case CONDAND:
		return "&&"
	case CONDOR:
		return "||"
	case AT:
		return "@"
	case CMD:
		return "``"
	case FUNCTION:
		return "FUNCTION"
	case LET:
		return "LET"
	case CONST:
		return "CONST"
	case TRUE:
		return "TRUE"
	case FALSE:
		return "FALSE"
	case IF:
		return "IF"
	case ELIF:
		return "ELIF"
	case ELSE:
		return "ELSE"
	case RETURN:
		return "RETURN"
	case IMPORT:
		return "IMPORT"
	case STRING:
		return "STRING"
	case ISTRING:
		return "ISTRING"
	case BYTES:
		return "BYTES"
	case AND:
		return "AND"
	case OR:
		return "OR"
	case STRUCT:
		return "STRUCT"
	case DO:
		return "DO"
	case WHILE:
		return "WHILE"
	case BREAK:
		return "BREAK"
	case CONTINUE:
		return "CONTINUE"
	case COMMENT:
		return "#"
	case REGEX:
		return "/"
	case FOR:
		return "FOR"
	case IN:
		return "IN"
	case WHERE:
		return "WHERE"
	case GREP:
		return "GREP"
	case MAP:
		return "MAP"
	case CASE:
		return "CASE"
	case IS:
		return "IS"
	case TRY:
		return "TRY"
	case CATCH:
		return "CATCH"
	case FINALLY:
		return "FINALLY"
	case THROW:
		return "THROW"
	case QUESTIONM:
		return "?"
	case QUESTIONMM:
		return "??"
	case DEFER:
		return "DEFER"
	case NIL:
		return "NIL"
	case ENUM:
		return "ENUM"
	case QW:
		return "QW"
	case UNLESS:
		return "UNLESS"
	case INTERFACE:
		return "INTERFACE"
	case CLASS:
		return "CLASS"
	case NEW:
		return "NEW"
	case PROPERTY:
		return "PROPERTY"
	case GET:
		return "GET"
	case SET:
		return "SET"
	case PUBLIC:
		return "PUBLIC"
	case PRIVATE:
		return "PRIVATE"
	case PROTECTED:
		return "PROTECTED"
	case STATIC:
		return "STATIC"
	case DEFAULT:
		return "DEFAULT"
	case UDO:
		return "USER-DEFINED-OPERATOR"
	case UNDERSCORE:
		return "_"
	case TILDEPLUS:
		return "~+"
	case TILDEMINUS:
		return "~-"
	case TILDEASTERISK:
		return "~*"
	case TILDESLASH:
		return "~/"
	case TILDEMOD:
		return "~%"
	case TILDECARET:
		return "~^"
	case USING:
		return "USING"

	//linq query
	case FROM:
		return "FROM"
	//case WHERE:
	//	return "WHERE"
	case SELECT:
		return "SELECT"
	case GROUP:
		return "GROUP"
	case INTO:
		return "INTO"
	case ORDERBY:
		return "ORDERBY"
	case JOIN:
		return "JOIN"
	//case LET:
	//	return "LET"
	//case IN:
	//	return "IN"
	case ON:
		return "ON"
	case EQUALS:
		return "EQUALS"
	case BY:
		return "BY"
	case ASCENDING:
		return "ASCENDING"
	case DESCENDING:
		return "DESCENDING"

	//async & await
	case ASYNC:
		return "ASYNC"
	case AWAIT:
		return "AWAIT"

	case SERVICE:
		return "SERVICE"

	//macors
	case DEFINE:
		return "#define"
	case IFDEF_MACRO:
		return "#ifdef"
	case ELSE_MACRO:
		return "#else"

	default:
		return "UNKNOWN"
	}
}

type Token struct {
	Pos     Position
	Type    TokenType
	Literal string
}

//Stringer method for Token
func (t Token) String() string {
	return fmt.Sprintf("Position: %s, TokType: %s, Literal: %s", t.Pos, t.Type, t.Literal)
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
