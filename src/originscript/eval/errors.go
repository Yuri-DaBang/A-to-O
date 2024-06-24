package eval

import "fmt"
import "strings"

// constants for error types
const (
	_ int = iota
	PREFIXOP
	INFIXOP
	POSTFIXOP
	MOD_ASSIGNOP
	UNKNOWNIDENT
	UNKNOWNIDENTEX
	NOMETHODERROR
	NOMETHODERROREX
	NOINDEXERROR
	KEYERROR
	INDEXERROR
	SLICEERROR
	ARGUMENTERROR
	INPUTERROR
	RTERROR
	PARAMTYPEERROR
	INLENERR
	INVALIDARG
	DIVIDEBYZERO
	THROWERROR
	THROWNOTHANDLED
	GREPMAPNOTITERABLE
	NOTITERABLE
	RANGETYPEERROR
	DEFERERROR
	SPAWNERROR
	ASSERTIONERROR
	//	STDLIBERROR
	NULLABLEERROR
	JSONERROR
	DBSCANERROR
	FUNCCALLBACKERROR
	FILEMODEERROR
	FILEOPENERROR
	NOTCLASSERROR
	PARENTNOTDECL
	CLSNOTDEFINE
	CLSMEMBERPRIVATE
	CLSCALLPRIVATE
	PROPERTYUSEERROR
	MEMBERUSEERROR
	INDEXERUSEERROR
	INDEXERTYPEERROR
	INDEXERSTATICERROR
	INDEXNOTFOUNDERROR
	CALLNONSTATICERROR
	CLASSCATEGORYERROR
	CLASSCREATEERROR
	PARENTNOTANNOTATION
	OVERRIDEERROR
	METAOPERATORERROR
	SERVICENOURLERROR
	CONSTNOTASSIGNERROR
	DIAMONDOPERERROR
	NAMENOTEXPORTED
	IMPORTERROR
	GENERICERROR
)

var errorType = map[int]string{
	PREFIXOP:           " AeroScript: eUDE: unsupported operator for prefix expression:'%s' and type: %s",
	INFIXOP:            " AeroScript: eUDE: unsupported operator for infix expression: %s '%s' %s",
	POSTFIXOP:          " AeroScript: eUDE: unsupported operator for postfix expression:'%s' and type: %s",
	MOD_ASSIGNOP:       " AeroScript: eUDE: unsupported operator for modulor assignment:'%s'",
	UNKNOWNIDENT:       " AeroScript: eUDE: unknown identifier: '%s' is not defined",
	UNKNOWNIDENTEX:     " AeroScript: eUDE: identifier '%s' not found. \n\nDid you mean one of: \n\n  %s\n",
	NOMETHODERROR:      " AeroScript: eUDE: undefined method '%s' for object %s",
	NOMETHODERROREX:    " AeroScript: eUDE: undefined method '%s' for object '%s'. \n\nDid you mean one of: \n\n  %s\n",
	NOINDEXERROR:       " AeroScript: eUDE: index error: type %s is not indexable",
	KEYERROR:           " AeroScript: eUDE: key error: type %s is not hashable",
	INDEXERROR:         " AeroScript: eUDE: index error: '%d' out of range",
	SLICEERROR:         " AeroScript: eUDE: index error: slice '%d:%d' out of range",
	ARGUMENTERROR:      " AeroScript: eUDE: wrong number of arguments. expected=%s, got=%d",
	INPUTERROR:         " AeroScript: eUDE: unsupported input type '%s' for function or method: %s",
	RTERROR:            " AeroScript: eUDE: return type should be %s",
	PARAMTYPEERROR:     " AeroScript: eUDE: %s argument for '%s' should be type %s. got=%s",
	INLENERR:           " AeroScript: eUDE: function %s takes input with max length %s. got=%s",
	INVALIDARG:         " AeroScript: eUDE: invalid argument supplied",
	DIVIDEBYZERO:       " AeroScript: eUDE: divide by zero",
	THROWERROR:         " AeroScript: eUDE: throw object must be a string",
	THROWNOTHANDLED:    " AeroScript: eUDE: throw object '%s' not handled",
	GREPMAPNOTITERABLE: " AeroScript: eUDE: grep/map's operating type must be iterable",
	NOTITERABLE:        " AeroScript: eUDE: foreach's operating type must be iterable",
	RANGETYPEERROR:     " AeroScript: eUDE: range(..) type should be %s type, got='%s'",
	DEFERERROR:         " AeroScript: eUDE: defer outside function or defer statement not a function",
	SPAWNERROR:         " AeroScript: eUDE: spawn must be followed by a function",
	ASSERTIONERROR:     " AeroScript: eUDE: assertion failed",
	//	STDLIBERROR:     " AeroScript: eUDE: calling '%s' failed",
	NULLABLEERROR:       " AeroScript: eUDE: %s is null",
	JSONERROR:           " AeroScript: eUDE: json error: maybe unsupported type or invalid data",
	DBSCANERROR:         " AeroScript: eUDE: scan type not supported",
	FUNCCALLBACKERROR:   " AeroScript: eUDE: callback error: must be '%d' parameter(s), got '%d'",
	FILEMODEERROR:       " AeroScript: eUDE: known file mode supplied",
	FILEOPENERROR:       " AeroScript: eUDE: file open failed, reason: %s",
	NOTCLASSERROR:       " AeroScript: eUDE: Identifier %s is not a class",
	PARENTNOTDECL:       " AeroScript: eUDE: Parent class %s not declared",
	CLSNOTDEFINE:        " AeroScript: eUDE: Class %s not defined",
	CLSMEMBERPRIVATE:    " AeroScript: eUDE: Variable(%s) of class(%s) is private",
	CLSCALLPRIVATE:      " AeroScript: eUDE: Method (%s) of class(%s) is private",
	PROPERTYUSEERROR:    " AeroScript: eUDE: Invalid use of Property(%s) of class(%s)",
	MEMBERUSEERROR:      " AeroScript: eUDE: Invalid use of member(%s) of class(%s)",
	INDEXERUSEERROR:     " AeroScript: eUDE: Invalid use of Indexer of class(%s)",
	INDEXERTYPEERROR:    " AeroScript: eUDE: Invalid use of Indexer of class(%s), Only interger type of Indexer is supported",
	INDEXERSTATICERROR:  " AeroScript: eUDE: Invalid use of Indexer of class(%s), Indexer cannot declared as static",
	INDEXNOTFOUNDERROR:  " AeroScript: eUDE: Indexer not found for class(%s)",
	CALLNONSTATICERROR:  " AeroScript: eUDE: Could not call non-static",
	CLASSCATEGORYERROR:  " AeroScript: eUDE: No class(%s) found for category(%s)",
	CLASSCREATEERROR:    " AeroScript: eUDE: You must use 'new' to create class('%s')",
	PARENTNOTANNOTATION: " AeroScript: eUDE: Annotation(%s)'s Parent(%s) is not annotation",
	OVERRIDEERROR:       " AeroScript: eUDE: Method(%s) of class(%s) must override a superclass method",
	METAOPERATORERROR:   " AeroScript: eUDE: Meta-Operators' item must be Numbers|String",
	SERVICENOURLERROR:   " AeroScript: eUDE: Service(%s)'s function('%s') must have url",
	CONSTNOTASSIGNERROR: " AeroScript: eUDE: Const variable '%s' cannot be modified",
	DIAMONDOPERERROR:    " AeroScript: eUDE: Diamond operator must be followed by a file object, but got '%s'",
	NAMENOTEXPORTED:     " AeroScript: eUDE: Cannot refer to unexported name '%s.%s'",
	IMPORTERROR:         " AeroScript: eUDE: Import error: %s",
	GENERICERROR:        " AeroScript: eUDE: %s",
}

func NewError(line string, t int, args ...interface{}) Object {
	msg := fmt.Sprintf(errorType[t], args...) + " at line " + strings.TrimLeft(line, " \t")
	return &Error{Kind: t, Message: msg}
}

type Error struct {
	Kind    int
	Message string
}

func (e Error) Error() string {
	return e.Message
}

func (e *Error) Inspect() string  { return "Runtime Error:" + e.Message + "\n" }
func (e *Error) Type() ObjectType { return ERROR_OBJ }
func (e *Error) CallMethod(line string, scope *Scope, method string, args ...Object) Object {
	//	return NewError(line, NOMETHODERROR, method, e.Type())
	return NewError(line, GENERICERROR, e.Message)
}
