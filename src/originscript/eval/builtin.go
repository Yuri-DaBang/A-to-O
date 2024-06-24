package eval

import (
	"container/list"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"originscript/ast"
	"net"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

var fileModeTable = map[string]int{
	"r":   os.O_RDONLY,
	"<":   os.O_RDONLY,
	"w":   os.O_WRONLY | os.O_CREATE | os.O_TRUNC,
	">":   os.O_WRONLY | os.O_CREATE | os.O_TRUNC,
	"a":   os.O_APPEND | os.O_CREATE,
	">>":  os.O_APPEND | os.O_CREATE,
	"r+":  os.O_RDWR,
	"+<":  os.O_RDWR,
	"w+":  os.O_RDWR | os.O_CREATE | os.O_TRUNC,
	"+>":  os.O_RDWR | os.O_CREATE | os.O_TRUNC,
	"a+":  os.O_RDWR | os.O_APPEND | os.O_CREATE,
	"+>>": os.O_RDWR | os.O_APPEND | os.O_CREATE,
}

type BuiltinFunc func(line string, scope *Scope, args ...Object) Object

type Builtin struct {
	Fn BuiltinFunc
}

var builtins map[string]*Builtin

func absBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			if len(args) != 1 {
				return NewError(line, ARGUMENTERROR, "1", len(args))
			}

			switch o := args[0].(type) {
			case *Integer:
				if o.Int64 > -1 {
					return o
				}
				return NewInteger(o.Int64 * -1)
			case *UInteger:
				return o
			default:
				return NewError(line, PARAMTYPEERROR, "first", "abs", "*Integer|*UInteger", args[0].Type())
			}
		}, //Here the ',' is a must, it confused me a lot
	}
}

func rangeBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			if len(args) != 1 && len(args) != 2 {
				return NewError(line, ARGUMENTERROR, "1|2", len(args))
			}

			var iValue int64
			switch o := args[0].(type) {
			case *Integer:
				iValue = o.Int64
			case *UInteger:
				iValue = int64(o.UInt64)
			default:
				return NewError(line, PARAMTYPEERROR, "first", "range", "*Integer|*UInteger", args[0].Type())
			}

			if iValue <= 0 {
				return &Array{}
			}

			var jValue int64
			if len(args) == 2 {
				switch o := args[1].(type) {
				case *Integer:
					jValue = o.Int64
				case *UInteger:
					jValue = int64(o.UInt64)
				default:
					return NewError(line, PARAMTYPEERROR, "second", "range", "*Integer|*UInteger", args[0].Type())
				}

				if jValue <= 0 {
					return NewError(line, GENERICERROR, "second parameter of 'range' should be >=0")
				}
			}

			var k int64
			methods := &Array{}
			for k = 0; k < iValue; {
				methods.Members = append(methods.Members, NewInteger(k))
				if len(args) == 2 {
					k += jValue
				} else {
					k += 1
				}
			}
			return methods
		},
	}
}

func addmBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			if len(args) != 3 {
				return NewError(line, ARGUMENTERROR, "3", len(args))
			}
			st, ok := args[0].(*Struct)
			if !ok {
				return NewError(line, PARAMTYPEERROR, "first", "addm", "*Struct", args[0].Type())
			}
			name, ok := args[1].(*String)
			if !ok {
				return NewError(line, PARAMTYPEERROR, "second", "addm", "*String", args[1].Type())
			}
			fn, ok := args[2].(*Function)
			if !ok {
				return NewError(line, PARAMTYPEERROR, "third", "addm", "*Function", args[2].Type())
			}
			st.methods[name.String] = fn
			return st
		},
	}
}

func chrBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			if len(args) != 1 {
				return NewError(line, ARGUMENTERROR, "1", len(args))
			}

			var i int64
			switch o := args[0].(type) {
			case *Integer:
				i = o.Int64
			case *UInteger:
				i = int64(o.UInt64)
			default:
				return NewError(line, PARAMTYPEERROR, "first", "chr", "*Integer|*UInteger", args[0].Type())
			}

			if i < 0 || i > 255 {
				return NewError(line, INPUTERROR, strconv.FormatInt(i, 10), "chr")
			}
			return NewString(strconv.FormatInt(i, 10))
		},
	}
}

func newFileBuiltin(funcName string) *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			var fname *String
			var flag int = os.O_RDONLY
			var ok bool
			var perm os.FileMode = os.FileMode(0666)

			argLen := len(args)
			if argLen < 1 {
				return NewError(line, ARGUMENTERROR, "at least one", argLen)
			}

			fname, ok = args[0].(*String)
			if !ok {
				return NewError(line, PARAMTYPEERROR, "first", funcName, "*String", args[0].Type())
			}

			if argLen == 2 {
				m, ok := args[1].(*String)
				if !ok {
					return NewError(line, PARAMTYPEERROR, "second", funcName, "*String", args[1].Type())
				}

				flag, ok = fileModeTable[m.String]
				if !ok {
					return NewError(line, FILEMODEERROR)
				}
			}

			if len(args) == 3 {
				p, ok := args[2].(*Integer)
				if !ok {
					return NewError(line, PARAMTYPEERROR, "third", funcName, "*Integer", args[2].Type())
				}

				perm = os.FileMode(int(p.Int64))
			}

			f, err := os.OpenFile(fname.String, flag, perm)
			if err != nil {
				return NewNil(err.Error())
			}
			return &FileObject{File: f, Name: "<file object: " + fname.String + ">"}
		},
	}
}

func intBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			if len(args) == 0 {
				//returns an empty int(defaults to 0)
				return NewInteger(0)
			}
			if len(args) != 1 {
				return NewError(line, ARGUMENTERROR, "1", len(args))
			}
			switch input := args[0].(type) {
			case *Integer:
				return input
			case *UInteger:
				return NewInteger(int64(input.UInt64))
			case *Float:
				return NewInteger(int64(input.Float64))
			case *DecimalObj:
				return NewInteger(input.Number.IntPart())
			case *Boolean:
				if input.Bool {
					return NewInteger(1)
				}
				return NewInteger(0)
			case *String:
				var n int64
				var err error

				var content = input.String
				if len(content) == 0 {
					return NewInteger(0)
				}

				if strings.HasPrefix(content, "0b") {
					n, err = strconv.ParseInt(content[2:], 2, 64)
				} else if strings.HasPrefix(content, "0x") {
					n, err = strconv.ParseInt(content[2:], 16, 64)
				} else if strings.HasPrefix(content, "0o") {
					n, err = strconv.ParseInt(content[2:], 8, 64)
				} else {
					n, err = strconv.ParseInt(content, 10, 64)
				}
				if err != nil {
					return NewError(line, INPUTERROR, "STRING: "+input.String, "int")
				}
				return NewInteger(n)
			}
			return NewError(line, PARAMTYPEERROR, "first", "int", "*String|*Integer|*UInteger|*Boolean|*Float", args[0].Type())
		},
	}
}

func uintBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			if len(args) == 0 {
				//returns an empty int(defaults to 0)
				return NewInteger(0)
			}
			if len(args) != 1 {
				return NewError(line, ARGUMENTERROR, "1", len(args))
			}
			switch input := args[0].(type) {
			case *Integer:
				return NewUInteger(uint64(input.Int64))
			case *UInteger:
				return input
			case *Float:
				return NewUInteger(uint64(input.Float64))
			case *DecimalObj:
				return NewUInteger(uint64(input.Number.IntPart()))
			case *Boolean:
				if input.Bool {
					return NewUInteger(1)
				}
				return NewUInteger(0)
			case *String:
				var n uint64
				var err error

				var content = input.String
				if len(content) == 0 {
					return NewUInteger(0)
				}
				if strings.HasPrefix(content, "0b") {
					n, err = strconv.ParseUint(content[2:], 2, 64)
				} else if strings.HasPrefix(content, "0x") {
					n, err = strconv.ParseUint(content[2:], 16, 64)
				} else if strings.HasPrefix(content, "0o") {
					n, err = strconv.ParseUint(content[2:], 8, 64)
				} else {
					n, err = strconv.ParseUint(content, 10, 64)
				}
				if err != nil {
					return NewError(line, INPUTERROR, "STRING: "+input.String, "uint")
				}
				return NewUInteger(n)
			}
			return NewError(line, PARAMTYPEERROR, "first", "int", "*String|*Integer|*UInteger|*Boolean|*Float", args[0].Type())
		},
	}
}

func floatBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			if len(args) == 0 {
				//returns an empty float(defaults to 0.0)
				return NewFloat(0.0)
			}
			if len(args) != 1 {
				return NewError(line, ARGUMENTERROR, "1", len(args))
			}
			switch input := args[0].(type) {
			case *Integer:
				return NewFloat(float64(input.Int64))
			case *UInteger:
				return NewFloat(float64(input.UInt64))
			case *Float:
				return input
			case *DecimalObj:
				f, _ := input.Number.Float64()
				return NewFloat(f)
			case *Boolean:
				if input.Bool {
					return NewFloat(1)
				}
				return NewFloat(0)
			case *String:
				var n float64
				var err error
				var k int64

				if len(input.String) == 0 {
					return NewFloat(0)
				}
				if strings.HasPrefix(input.String, "0b") {
					k, err = strconv.ParseInt(input.String[2:], 2, 64)
					if err == nil {
						n = float64(k)
					}
				} else if strings.HasPrefix(input.String, "0x") {
					k, err = strconv.ParseInt(input.String[2:], 16, 64)
					if err == nil {
						n = float64(k)
					}
				} else if strings.HasPrefix(input.String, "0o") {
					k, err = strconv.ParseInt(input.String[2:], 8, 64)
					if err == nil {
						n = float64(k)
					}
				} else {
					n, err = strconv.ParseFloat(input.String, 64)
				}
				if err != nil {
					return NewError(line, INPUTERROR, "STRING: "+input.String, "float")
				}
				return NewFloat(float64(n))
			}
			return NewError(line, PARAMTYPEERROR, "first", "float", "*String|*Integer|*UInteger|*Boolean|*Float", args[0].Type())
		},
	}
}

func strBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			if len(args) == 0 {
				//returns an empty string
				return NewString("")
			}
			if len(args) != 1 {
				return NewError(line, ARGUMENTERROR, "1", len(args))
			}
			switch input := args[0].(type) {
			case *String:
				return input
			default:
				return NewString(input.Inspect())
			}
			//return NewError(line, INPUTERROR, args[0].Type(), "str")
		},
	}
}

func arrayBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			if len(args) == 0 {
				//returns an empty array
				return &Array{Members: []Object{}}
			}

			if len(args) != 1 {
				return NewError(line, ARGUMENTERROR, "1", len(args))
			}
			switch input := args[0].(type) {
			case *Array:
				return input
			case *Tuple:
				length := len(input.Members)
				newMembers := make([]Object, length)
				copy(newMembers, input.Members)
				return &Array{Members: newMembers}
			default:
				return &Array{Members: []Object{input}}
			}
		},
	}
}

func tupleBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			if len(args) == 0 {
				//returns an empty tuple
				return &Tuple{Members: []Object{}}
			}

			if len(args) != 1 {
				return NewError(line, ARGUMENTERROR, "1", len(args))
			}
			switch input := args[0].(type) {
			case *Tuple:
				return input
			case *Array:
				length := len(input.Members)
				newMembers := make([]Object, length)
				copy(newMembers, input.Members)
				return &Tuple{Members: newMembers}
			default:
				return &Tuple{Members: []Object{input}}
			}
		},
	}
}

func hashBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			if len(args) == 0 {
				//returns an empty hash
				return NewHash()
			}

			if len(args) != 1 {
				return NewError(line, ARGUMENTERROR, "1", len(args))
			}
			switch input := args[0].(type) {
			case *Hash:
				return input
			case *Tuple:
				length := len(input.Members)
				if length == 0 { //empty tuple
					//return empty hash
					return NewHash()
				}
				newMembers := make([]Object, length)
				copy(newMembers, input.Members)

				if length%2 != 0 {
					length = length + 1
					newMembers = append(newMembers, NIL)
				}

				hash := NewHash()
				for i := 0; i <= length/2; {
					if _, ok := newMembers[i].(Hashable); ok {
						hash.Push(line, newMembers[i], newMembers[i+1])
						//hash.Pairs[hashable.HashKey()] = HashPair{Key: newMembers[i], Value: newMembers[i+1]}
						i = i + 2
					} else {
						return NewError(line, GENERICERROR, fmt.Sprintf("%d index is not hashable", i))
					}
				}

				return hash
			case *Array:
				length := len(input.Members)
				if length == 0 { //empty tuple
					//return empty hash
					return NewHash()
				}
				newMembers := make([]Object, length)
				copy(newMembers, input.Members)

				if length%2 != 0 {
					length = length + 1
					newMembers = append(newMembers, NIL)
				}

				hash := NewHash()
				for i := 0; i <= length/2; {
					if _, ok := newMembers[i].(Hashable); ok {
						hash.Push(line, newMembers[i], newMembers[i+1])
						//hash.Pairs[hashable.HashKey()] = HashPair{Key: newMembers[i], Value: newMembers[i+1]}
						i = i + 2
					} else {
						return NewError(line, GENERICERROR, fmt.Sprintf("%d index is not hashable", i))
					}
				}

				return hash
			}
			return NewError(line, PARAMTYPEERROR, "first", "hash", "*Tuple|*Array|*Hash", args[0].Type())
		},
	}
}

func decimalBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			if len(args) == 0 {
				//returns an empty decimal(defaults to 0)
				return &DecimalObj{Number: NewDec(0, 0), Valid: true}
			}
			if len(args) != 1 {
				return NewError(line, ARGUMENTERROR, "1", len(args))
			}
			switch input := args[0].(type) {
			case *DecimalObj:
				return input
			case *Integer:
				return &DecimalObj{Number: NewFromFloat(float64(input.Int64)), Valid: true}
			case *UInteger:
				return &DecimalObj{Number: NewFromFloat(float64(input.UInt64)), Valid: true}
			case *Float:
				return &DecimalObj{Number: NewFromFloat(input.Float64), Valid: true}
			case *Boolean:
				if input.Bool {
					return &DecimalObj{Number: NewFromFloat(1), Valid: true}
				}
				return &DecimalObj{Number: NewFromFloat(0), Valid: true}
			case *String:
				var n float64
				var err error
				var k int64

				if strings.HasPrefix(input.String, "0b") {
					k, err = strconv.ParseInt(input.String[2:], 2, 64)
					if err == nil {
						n = float64(k)
					}
				} else if strings.HasPrefix(input.String, "0x") {
					k, err = strconv.ParseInt(input.String[2:], 16, 64)
					if err == nil {
						n = float64(k)
					}
				} else if strings.HasPrefix(input.String, "0o") {
					k, err = strconv.ParseInt(input.String[2:], 8, 64)
					if err == nil {
						n = float64(k)
					}
				} else {
					n, err = strconv.ParseFloat(input.String, 64)
				}
				if err != nil {
					return NewError(line, INPUTERROR, "STRING: "+input.String, "decimal")
				}
				return &DecimalObj{Number: NewFromFloat(n), Valid: true}
			}
			return NewError(line, PARAMTYPEERROR, "first", "decimal", "*String|*Integer|*UInteger|*Boolean|*Float|*Decimal", args[0].Type())
		},
	}
}

func lenBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			if len(args) != 1 {
				return NewError(line, ARGUMENTERROR, "1", len(args))
			}
			switch arg := args[0].(type) {
			case *String:
				n := utf8.RuneCountInString(arg.String)
				return NewInteger(int64(n))
			case *Array:
				return NewInteger(int64(len(arg.Members)))
			case *Tuple:
				return NewInteger(int64(len(arg.Members)))
			case *Hash:
				return NewInteger(int64(len(arg.Pairs)))
			case *Nil:
				return NewInteger(0)
			}
			return NewError(line, PARAMTYPEERROR, "first", "len", "*String|*Array|*Hash|*Nil", args[0].Type())
		},
	}
}

func methodsBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			if len(args) != 1 {
				return NewError(line, ARGUMENTERROR, "1", len(args))
			}
			arr := &Array{}
			t := reflect.TypeOf(args[0])
			for i := 0; i < t.NumMethod(); i++ {
				m := t.Method(i).Name
				if !(m == "Type" || m == "CallMethod" || m == "HashKey" || m == "Inspect") {
					arr.Members = append(arr.Members, NewString(strings.ToLower(m)))
				}
			}
			return arr
		},
	}
}

func ordBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			if len(args) != 1 {
				return NewError(line, ARGUMENTERROR, "1", len(args))
			}
			s, ok := args[0].(*String)
			if !ok {
				return NewError(line, PARAMTYPEERROR, "first", "ord", "*String", args[0].Type())
			}
			if len(s.String) > 1 {
				return NewError(line, INLENERR, "1", len(s.String))
			}
			return NewInteger(int64(s.String[0]))
		},
	}
}

func printBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			if len(args) == 0 {
				n, err := fmt.Print(scope.Writer)
				if err != nil {
					return NewNil(err.Error())
				}
				return NewInteger(int64(n))
			}

			format, wrapped := correctPrintResult(false, args...)
			n, err := fmt.Fprintf(scope.Writer, format, wrapped...)

			//Note, here we do not use 'fmt.Print', why? please see correctPrintResult() comments.
			//n, err := fmt.Print(s, wrapped...)
			if err != nil {
				return NewNil(err.Error())
			}

			return NewInteger(int64(n))
		},
	}
}

func printlnBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			if len(args) == 0 {
				n, err := fmt.Fprintln(scope.Writer)
				if err != nil {
					return NewNil(err.Error())
				}
				return NewInteger(int64(n))
			}

			//Note, here we do not use 'fmt.Println', why? please see correctPrintResult() comments.
			//n, err := fmt.Println(s, wrapped...)

			format, wrapped := correctPrintResult(true, args...)
			n, err := fmt.Fprintf(scope.Writer, format, wrapped...)
			if err != nil {
				return NewNil(err.Error())
			}

			return NewInteger(int64(n))
		},
	}
}

func printfBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			if len(args) < 1 {
				return NewError(line, ARGUMENTERROR, ">0", len(args))
			}

			formatObj, ok := args[0].(*String)
			if !ok {
				return NewError(line, PARAMTYPEERROR, "first", "printf", "*String", args[0].Type())
			}

			subArgs := args[1:]
			wrapped := make([]interface{}, len(subArgs))
			for i, v := range subArgs {
				wrapped[i] = &Formatter{Obj: v}
			}

			formatStr := formatObj.String
			if len(subArgs) == 0 {
				if REPLColor {
					formatStr = "\033[1;" + colorMap["STRING"] + "m" + formatStr + "\033[0m"
				}
			}
			n, err := fmt.Fprintf(scope.Writer, formatStr, wrapped...)

			if err != nil {
				return NewNil(err.Error())
			}

			return NewInteger(int64(n))
		},
	}
}

func sprintfBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			if len(args) < 1 {
				return NewError(line, ARGUMENTERROR, ">0", len(args))
			}

			formatObj, ok := args[0].(*String)
			if !ok {
				return NewError(line, PARAMTYPEERROR, "first", "sprintf", "*String", args[0].Type())
			}

			subArgs := args[1:]
			wrapped := make([]interface{}, len(subArgs))
			for i, v := range subArgs {
				wrapped[i] = &Formatter{Obj: v}
			}

			formatStr := formatObj.String
			if len(subArgs) == 0 {
				if REPLColor {
					formatStr = "\033[1;" + colorMap["STRING"] + "m" + formatStr + "\033[0m"
				}
			}
			out := fmt.Sprintf(formatStr, wrapped...)

			return NewString(out)
		},
	}
}

func sscanfBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			if len(args) < 2 {
				return NewError(line, ARGUMENTERROR, ">=2", len(args))
			}

			strObj, ok := args[0].(*String)
			if !ok {
				return NewError(line, PARAMTYPEERROR, "first", "sscanf", "*String", args[0].Type())
			}

			formatObj, ok := args[1].(*String)
			if !ok {
				return NewError(line, PARAMTYPEERROR, "second", "sscanf", "*String", args[1].Type())
			}

			subArgs := args[2:]
			values := make([]interface{}, len(subArgs))

			for i, v := range subArgs {
				switch input := v.(type) {
				case *Integer:
					values[i] = &input.Int64
				case *Float:
					values[i] = &input.Float64
				case *Boolean:
					values[i] = &input.Bool
				case *String:
					values[i] = &input.String
				}
			}

			formatStr := formatObj.String
			if len(subArgs) == 0 {
				if REPLColor {
					formatStr = "\033[1;" + colorMap["STRING"] + "m" + formatStr + "\033[0m"
				}
			}

			_, err := fmt.Sscanf(strObj.String, formatStr, values...)
			if err != nil { //error
				return NewNil(err.Error())
			}

			//convert go's interface{} back to magpie's Object
			for i := range subArgs {
				subArgs[i], _ = unmarshalJsonObject(values[i])
			}
			return NIL
		},
	}
}

func typeBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			if len(args) != 1 {
				return NewError(line, ARGUMENTERROR, "1", len(args))
			}
			return NewString(fmt.Sprintf("%s", args[0].Type()))
		},
	}
}

func chanBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			if len(args) == 0 {
				return &ChanObject{ch: make(chan Object)}
			} else if len(args) == 1 {
				var v int64
				switch o := args[0].(type) {
				case *Integer:
					v = o.Int64
				case *UInteger:
					v = int64(o.UInt64)
				default:
					return NewError(line, PARAMTYPEERROR, "first", "chan", "*Integer|*UInteger", args[0].Type())
				}
				return &ChanObject{ch: make(chan Object, v)}
			}
			return NewError(line, ARGUMENTERROR, "Not 0|1", len(args))
		},
	}
}

func assertBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			if len(args) != 1 {
				return NewError(line, ARGUMENTERROR, "1", len(args))
			}

			v, ok := args[0].(*Boolean)
			if !ok {
				return NewError(line, PARAMTYPEERROR, "first", "assert", "*Boolean", args[0].Type())
			}

			if v.Bool == true {
				return NIL
			}

			return NewError(line, ASSERTIONERROR)
		},
	}
}

func reverseBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			if len(args) != 1 {
				return NewError(line, ARGUMENTERROR, "1", len(args))
			}

			switch input := args[0].(type) {
			case *String:
				word := []rune(input.String)
				reverse := []rune{}
				for i := len(word) - 1; i >= 0; i-- {
					reverse = append(reverse, word[i])
				}

				return NewString(string(reverse))
			case *Array:
				reverse := &Array{}
				for i := len(input.Members) - 1; i >= 0; i-- {
					reverse.Members = append(reverse.Members, input.Members[i])
				}
				return reverse
			case *Tuple:
				reverse := &Tuple{}
				for i := len(input.Members) - 1; i >= 0; i-- {
					reverse.Members = append(reverse.Members, input.Members[i])
				}
				return reverse
			case *Hash:
				hash := NewHash()
				for _, hk := range input.Order { //hk:hash key
					v, _ := input.Pairs[hk]
					hash.Push(line, v.Value, v.Key)
					//hash.Pairs[hashable.HashKey()] = HashPair{Key: v.Value, Value: v.Key}
				}
				return hash
			default:
				return NewError(line, PARAMTYPEERROR, "first", "reverse", "*Array|*String", args[0].Type())
			}
		},
	}
}

func iffBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			if len(args) != 3 {
				return NewError(line, ARGUMENTERROR, "3", len(args))
			}

			v, ok := args[0].(*Boolean)
			if !ok {
				return NewError(line, PARAMTYPEERROR, "first", "iff", "*Boolean", args[0].Type())
			}

			if v.Bool == true {
				return args[1]
			} else {
				return args[2]
			}
		},
	}
}

func newArrayBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			if len(args) < 0 {
				return NewError(line, ARGUMENTERROR, ">0", len(args))
			}

			var count int64
			switch o := args[0].(type) {
			case *Integer:
				count = o.Int64
			case *UInteger:
				count = int64(o.UInt64)
			default:
				return NewError(line, PARAMTYPEERROR, "first", "newArray", "*Integer|*UInteger", args[0].Type())
			}

			if count < 0 {
				return NewError(line, GENERICERROR, "Parameter of 'newArry' is less than zero.")
			}

			remainingArgs := args[1:]
			remainingLen := len(remainingArgs)

			var newLen int64 = 0
			for i := 0; i < remainingLen; i++ {
				if remainingArgs[i].Type() == ARRAY_OBJ {
					newLen += int64(len(remainingArgs[i].(*Array).Members))
				} else {
					newLen += 1
				}
			}

			ret := &Array{}
			for i := 0; i < remainingLen; i++ {
				if remainingArgs[i].Type() == ARRAY_OBJ {
					ret.Members = append(ret.Members, remainingArgs[i].(*Array).Members...)
				} else {
					ret.Members = append(ret.Members, remainingArgs[i])
				}
			}

			if count <= newLen {
				return ret
			}

			//count > newLen
			for i := newLen; i < count; i++ {
				ret.Members = append(ret.Members, NIL)
			}
			return ret
		},
	}
}

func dialTCPBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			if len(args) != 2 {
				return NewError(line, ARGUMENTERROR, "2", len(args))
			}
			netStr, ok := args[0].(*String)
			if !ok {
				return NewError(line, PARAMTYPEERROR, "first", "dialTCP", "*String", args[0].Type())
			}
			addrStr, ok := args[1].(*String)
			if !ok {
				return NewError(line, PARAMTYPEERROR, "second", "dialTCP", "*String", args[1].Type())
			}

			tcpAddr, err := net.ResolveTCPAddr(netStr.String, addrStr.String)
			if err != nil {
				return NewNil(err.Error())
			}

			conn, err := net.DialTCP(netStr.String, nil, tcpAddr)
			if err != nil {
				return NewNil(err.Error())
			}

			return &TcpConnObject{Conn: conn, Address: tcpAddr.String()}
		},
	}
}

func listenTCPBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			if len(args) != 2 {
				return NewError(line, ARGUMENTERROR, "2", len(args))
			}
			netStr, ok := args[0].(*String)
			if !ok {
				return NewError(line, PARAMTYPEERROR, "first", "listenTCP", "*String", args[0].Type())
			}
			addrStr, ok := args[1].(*String)
			if !ok {
				return NewError(line, PARAMTYPEERROR, "second", "listenTCP", "*String", args[1].Type())
			}

			tcpAddr, err := net.ResolveTCPAddr(netStr.String, addrStr.String)
			if err != nil {
				return NewNil(err.Error())
			}

			listener, err := net.ListenTCP(netStr.String, tcpAddr)
			if err != nil {
				return NewNil(err.Error())
			}

			return &TCPListenerObject{Listener: listener, Address: tcpAddr.String()}
		},
	}
}

func dialUDPBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			if len(args) != 2 {
				return NewError(line, ARGUMENTERROR, "2", len(args))
			}
			netStr, ok := args[0].(*String)
			if !ok {
				return NewError(line, PARAMTYPEERROR, "first", "dialUDP", "*String", args[0].Type())
			}
			addrStr, ok := args[1].(*String)
			if !ok {
				return NewError(line, PARAMTYPEERROR, "second", "dialUDP", "*String", args[1].Type())
			}

			udpAddr, err := net.ResolveUDPAddr(netStr.String, addrStr.String)
			if err != nil {
				return NewNil(err.Error())
			}

			conn, e := net.DialUDP(netStr.String, nil, udpAddr)
			if e != nil {
				return NewNil(err.Error())
			}

			return &UdpConnObject{Conn: conn, Address: udpAddr.String()}
		},
	}
}

func dialUnixBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			if len(args) != 2 {
				return NewError(line, ARGUMENTERROR, "2", len(args))
			}
			netStr, ok := args[0].(*String)
			if !ok {
				return NewError(line, PARAMTYPEERROR, "first", "dialUnix", "*String", args[0].Type())
			}
			addrStr, ok := args[1].(*String)
			if !ok {
				return NewError(line, PARAMTYPEERROR, "second", "dialUnix", "*String", args[1].Type())
			}

			unixAddr, err := net.ResolveUnixAddr(netStr.String, addrStr.String)
			if err != nil {
				return NewNil(err.Error())
			}

			conn, err := net.DialUnix(netStr.String, nil, unixAddr)
			if err != nil {
				return NewNil(err.Error())
			}

			return &UnixConnObject{Conn: conn, Address: unixAddr.String()}
		},
	}
}

func listenUnixBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			if len(args) != 2 {
				return NewError(line, ARGUMENTERROR, "2", len(args))
			}
			netStr, ok := args[0].(*String)
			if !ok {
				return NewError(line, PARAMTYPEERROR, "first", "listenTCP", "*String", args[0].Type())
			}
			addrStr, ok := args[1].(*String)
			if !ok {
				return NewError(line, PARAMTYPEERROR, "second", "listenTCP", "*String", args[1].Type())
			}

			unixAddr, err := net.ResolveUnixAddr(netStr.String, addrStr.String)
			if err != nil {
				return NewNil(err.Error())
			}

			listener, err := net.ListenUnix(netStr.String, unixAddr)
			if err != nil {
				return NewNil(err.Error())
			}

			return &UnixListenerObject{Listener: listener, Address: unixAddr.String()}
		},
	}
}

func dbOpenBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			if len(args) != 2 {
				return NewError(line, ARGUMENTERROR, "2", len(args))
			}
			driverName, ok := args[0].(*String)
			if !ok {
				return NewError(line, PARAMTYPEERROR, "first", "dbOpen", "*String", args[0].Type())
			}

			dataSourceName, ok := args[1].(*String)
			if !ok {
				return NewError(line, PARAMTYPEERROR, "second", "dbOpen", "*String", args[1].Type())
			}

			db, err := sql.Open(driverName.String, dataSourceName.String)
			if err != nil {
				return NewNil(err.Error())
			}

			return &SqlObject{Db: db, Name: fmt.Sprintf("%s:%s", driverName.String, dataSourceName.String)}
		},
	}
}

func newTimeBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			if len(args) != 0 && len(args) != 1 {
				return NewError(line, ARGUMENTERROR, "0|1", len(args))
			}

			if len(args) == 0 {
				return &TimeObj{Tm: time.Now(), Valid: true}
			}

			location, ok := args[0].(*Integer)
			if !ok {
				return NewError(line, PARAMTYPEERROR, "first", "newTime", "*Integer", args[0].Type())
			}
			if location.Int64 == UTC {
				return &TimeObj{Tm: time.Now().UTC(), Valid: true}
			}
			return &TimeObj{Tm: time.Now(), Valid: true}
		},
	}
}

/* accept a timestamp(second & nanosecond), and convert it to a TimeObj */
func unixTimeBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			if len(args) != 1 && len(args) != 2 {
				return NewError(line, ARGUMENTERROR, "1|2", len(args))
			}

			var ok bool
			var second *Integer
			var nsecond *Integer
			if len(args) == 1 {
				second, ok = args[0].(*Integer)
				if !ok {
					return NewError(line, PARAMTYPEERROR, "first", "unixTime", "*Integer", args[0].Type())
				}
			}

			if len(args) == 2 {
				nsecond, ok = args[1].(*Integer)
				if !ok {
					return NewError(line, PARAMTYPEERROR, "second", "unixTime", "*Integer", args[1].Type())
				}
			}

			if len(args) == 1 {
				return &TimeObj{Tm: time.Unix(second.Int64, 0), Valid: true}
			}
			return &TimeObj{Tm: time.Unix(second.Int64, nsecond.Int64), Valid: true}
		},
	}
}

//func Date(year int, month Month, day, hour, min, sec, nsec int, loc int)
func newDateBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			argLen := len(args)
			if argLen != 7 && argLen != 8 {
				return NewError(line, ARGUMENTERROR, "7|8", len(args))
			}

			year, ok1 := args[0].(*Integer)
			if !ok1 {
				return NewError(line, PARAMTYPEERROR, "first", "newDate", "*Integer", args[0].Type())
			}

			month, ok2 := args[1].(*Integer)
			if !ok2 {
				return NewError(line, PARAMTYPEERROR, "second", "newDate", "*Integer", args[1].Type())
			}

			day, ok3 := args[2].(*Integer)
			if !ok3 {
				return NewError(line, PARAMTYPEERROR, "third", "newDate", "*Integer", args[2].Type())
			}

			hour, ok4 := args[3].(*Integer)
			if !ok4 {
				return NewError(line, PARAMTYPEERROR, "fourth", "newDate", "*Integer", args[3].Type())
			}

			min, ok5 := args[4].(*Integer)
			if !ok5 {
				return NewError(line, PARAMTYPEERROR, "fifth", "newDate", "*Integer", args[4].Type())
			}

			sec, ok6 := args[5].(*Integer)
			if !ok6 {
				return NewError(line, PARAMTYPEERROR, "sixth", "newDate", "*Integer", args[5].Type())
			}

			nsec, ok7 := args[6].(*Integer)
			if !ok7 {
				return NewError(line, PARAMTYPEERROR, "seventh", "newDate", "*Integer", args[6].Type())
			}

			var location Object
			var ok8 bool
			if argLen == 8 {
				location, ok8 = args[7].(*Integer)
				if !ok8 {
					return NewError(line, PARAMTYPEERROR, "eighth", "newDate", "*Integer", args[7].Type())
				}
			}

			if argLen == 7 {
				return &TimeObj{Tm: time.Date(int(year.Int64), time.Month(month.Int64), int(day.Int64),
					int(hour.Int64), int(min.Int64), int(sec.Int64), int(nsec.Int64),
					time.Local), Valid: true}
			} else {
				var loc *time.Location
				if location.(*Integer).Int64 == LOCAL {
					loc = time.Local
				} else {
					loc = time.UTC
				}
				return &TimeObj{Tm: time.Date(int(year.Int64), time.Month(month.Int64), int(day.Int64),
					int(hour.Int64), int(min.Int64), int(sec.Int64), int(nsec.Int64),
					loc), Valid: true}
			}

		},
	}
}

func newCondBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			if len(args) != 1 {
				return NewError(line, ARGUMENTERROR, "1", len(args))
			}

			switch arg := args[0].(type) {
			case *SyncMutexObj:
				return &SyncCondObj{Cond: sync.NewCond(arg.Mutex)}
			case *SyncRWMutexObj:
				return &SyncCondObj{Cond: sync.NewCond(arg.RWMutex)}
			default:
				return NewError(line, PARAMTYPEERROR, "first", "newCond", "*SyncMutexObj|*SyncRWMutexObj", args[0].Type())
			}
		},
	}
}

func newOnceBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			if len(args) != 0 {
				return NewError(line, ARGUMENTERROR, "0", len(args))
			}
			return &SyncOnceObj{Once: new(sync.Once)}
		},
	}
}

func newMutexBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			if len(args) != 0 {
				return NewError(line, ARGUMENTERROR, "0", len(args))
			}
			return &SyncMutexObj{Mutex: new(sync.Mutex)}
		},
	}
}

func newRWMutexBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			if len(args) != 0 {
				return NewError(line, ARGUMENTERROR, "0", len(args))
			}
			return &SyncRWMutexObj{RWMutex: new(sync.RWMutex)}
		},
	}
}

func newWaitGroupBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			if len(args) != 0 {
				return NewError(line, ARGUMENTERROR, "0", len(args))
			}
			return &SyncWaitGroupObj{WaitGroup: new(sync.WaitGroup)}
		},
	}
}

func newPipeBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			if len(args) != 0 {
				return NewError(line, ARGUMENTERROR, "0", len(args))
			}
			r, w := io.Pipe()
			return &PipeObj{Reader: r, Writer: w}
		},
	}
}

func newLoggerBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			if len(args) != 3 && len(args) != 0 {
				return NewError(line, ARGUMENTERROR, "0|3", len(args))
			}

			if len(args) == 0 {
				logger := log.New(os.Stdout, "", log.LstdFlags)
				return &LoggerObj{Logger: logger}
			}

			out, ok := args[0].(Writable)
			if !ok {
				return NewError(line, PARAMTYPEERROR, "first", "newLogger", "Writable", args[0].Type())
			}

			prefix, ok := args[1].(*String)
			if !ok {
				return NewError(line, PARAMTYPEERROR, "second", "newLogger", "*String", args[1].Type())
			}

			flag, ok := args[2].(*Integer)
			if !ok {
				return NewError(line, PARAMTYPEERROR, "third", "newLogger", "*Integer", args[2].Type())
			}

			logger := log.New(out.IOWriter(), prefix.String, int(flag.Int64))
			return &LoggerObj{Logger: logger}
		},
	}
}

func newListBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			if len(args) != 0 {
				return NewError(line, ARGUMENTERROR, "0", len(args))
			}

			return &ListObject{List: list.New()}
		},
	}
}

func newDeepEqualBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			if len(args) != 2 {
				return NewError(line, ARGUMENTERROR, "2", len(args))
			}

			r := reflect.DeepEqual(args[0], args[1])
			if r {
				return TRUE
			}
			return FALSE
		},
	}
}

func newCsvReaderBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			argLen := len(args)
			if argLen != 1 {
				return NewError(line, ARGUMENTERROR, "1", argLen)
			}

			fname, ok := args[0].(*String)
			if !ok {
				return NewError(line, PARAMTYPEERROR, "first", "newCsv", "*String", args[0].Type())
			}

			f, err := os.Open(fname.String)
			if err != nil {
				return NewNil(err.Error())
			}

			return &CsvObj{Reader: csv.NewReader(f), ReaderFile: f}
		},
	}
}

func newCsvWriterBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			argLen := len(args)
			if argLen != 1 {
				return NewError(line, ARGUMENTERROR, "1", argLen)
			}

			writer, ok := args[0].(Writable)
			if !ok {
				return NewError(line, PARAMTYPEERROR, "first", "newCsvWriterBuiltin", "Writable", args[0].Type())
			}

			return &CsvObj{Writer: csv.NewWriter(writer.IOWriter())}
		},
	}
}

func instanceOfBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			argLen := len(args)
			if argLen != 2 {
				return NewError(line, ARGUMENTERROR, "2", argLen)
			}

			instance, ok := args[0].(*ObjectInstance)
			if !ok {
				return FALSE
			}

			switch class := args[1].(type) {
			case *String:
				return nativeBoolToBooleanObject(InstanceOf(class.String, instance))
			case *Class:
				return nativeBoolToBooleanObject(InstanceOf(class.Name, instance))
			}

			return NewError(line, GENERICERROR, "is_a/instanceOf expected a class or string for second argument")
		},
	}
}

func classOfBuiltin() *Builtin {
	return &Builtin{
		Fn: func(line string, scope *Scope, args ...Object) Object {
			argLen := len(args)
			if argLen != 1 {
				return NewError(line, ARGUMENTERROR, "1", argLen)
			}

			instance, ok := args[0].(*ObjectInstance)
			if !ok {
				return NewString("")
			}

			return NewString(instance.Class.Name)
		},
	}
}

func RegisterBuiltin(name string, f *Builtin) {
	builtins[strings.ToLower(name)] = f
}

func init() {
	builtins = map[string]*Builtin{
		"abs":      absBuiltin(),
		"range":    rangeBuiltin(),
		"addm":     addmBuiltin(),
		"chr":      chrBuiltin(),
		"newFile":  newFileBuiltin("newFile"),
		"open":     newFileBuiltin("open"),
		"int":      intBuiltin(),
		"uint":     uintBuiltin(),
		"float":    floatBuiltin(),
		"str":      strBuiltin(),
		"array":    arrayBuiltin(),
		"tuple":    tupleBuiltin(),
		"hash":     hashBuiltin(),
		"decimal":  decimalBuiltin(),
		"len":      lenBuiltin(),
		"methods":  methodsBuiltin(),
		"ord":      ordBuiltin(),
		"print":    printBuiltin(),
		"println":  printlnBuiltin(),
		"say":      printlnBuiltin(),
		"printf":   printfBuiltin(),
		"sprintf":  sprintfBuiltin(),
		"sscanf":   sscanfBuiltin(),
		"type":     typeBuiltin(),
		"chan":     chanBuiltin(),
		"assert":   assertBuiltin(),
		"reverse":  reverseBuiltin(),
		"iff":      iffBuiltin(),
		"newArray": newArrayBuiltin(),

		//net
		"dialTCP":    dialTCPBuiltin(),
		"listenTCP":  listenTCPBuiltin(),
		"dialUDP":    dialUDPBuiltin(),
		"dialUnix":   dialUnixBuiltin(),
		"listenUnix": listenUnixBuiltin(),

		//database
		"dbOpen": dbOpenBuiltin(),

		//time
		"newTime":  newTimeBuiltin(),
		"newDate":  newDateBuiltin(),
		"unixTime": unixTimeBuiltin(),

		//sync
		"newCond":      newCondBuiltin(),
		"newOnce":      newOnceBuiltin(),
		"newMutex":     newMutexBuiltin(),
		"newRWMutex":   newRWMutexBuiltin(),
		"newWaitGroup": newWaitGroupBuiltin(),

		//pipe
		"newPipe": newPipeBuiltin(),

		//Logger
		"newLogger": newLoggerBuiltin(),

		//container
		"newList": newListBuiltin(),

		//deepEqual
		"deepEqual": newDeepEqualBuiltin(),

		//csv
		"newCsvReader": newCsvReaderBuiltin(),
		"newCsvWriter": newCsvWriterBuiltin(),

		//class related
		"is_a":       instanceOfBuiltin(),
		"instanceOf": instanceOfBuiltin(),
		"classOf":    classOfBuiltin(),
	}
}

const BUILTINMETHOD_OBJ = "BUILTINMETHOD_OBJ"

type BuiltinMethodFunction func(line string, self *ObjectInstance, scope *Scope, args ...Object) Object

type BuiltinMethod struct {
	Fn       BuiltinMethodFunction
	Instance *ObjectInstance
}

func (b *BuiltinMethod) Inspect() string  { return "builtin method" }
func (b *BuiltinMethod) Type() ObjectType { return BUILTINMETHOD_OBJ }

//Could be used as class function
func (b *BuiltinMethod) classMethod() ast.ModifierLevel {
	return ast.ModifierPublic
}

func (b *BuiltinMethod) CallMethod(line string, scope *Scope, method string, args ...Object) Object {
	return NewError(line, NOMETHODERROR, method, b.Type())
}

func MakeBuiltinMethod(fn BuiltinMethodFunction) *BuiltinMethod {
	return &BuiltinMethod{Fn: fn}
}
