package eval

import (
	"fmt"
	"io"
	"originscript/ast"
	_ "os"
	"sync"
)

var BuiltinClasses = map[string]*Class{
	"object":   BASE_CLASS,
	"Override": OVERRIDE_ANNOCLASS,
	"NotNull":  NOTNULL_ANNOCLASS,
	"NotEmpty": NOTEMPTY_ANNOCLASS,
}

func NewScope(p *Scope, w io.Writer) *Scope {
	s := make(map[string]Object)
	r := make(map[string]bool)
	ret := &Scope{store: s, readonly: r, parentScope: p}
	if p == nil {
		ret.CallStack = &CallStack{Frames: []CallFrame{}} //creat a new empty CallStack
		ret.Writer = w
	} else {
		ret.Writer = p.Writer
		ret.CallStack = p.CallStack
	}

	return ret
}

//CallStack is a stack for CallFrame
type CallStack struct {
	Frames []CallFrame
}

type CallFrame struct {
	FuncScope   *Scope
	CurrentCall *ast.CallExpression // currently calling function
	defers      []func()            // function's defers
}

func (frame *CallFrame) runDefers(s *Scope) {
	// execute defers last-to-first
	defers := frame.defers
	for i := len(defers) - 1; i >= 0; i-- {
		defers[i]()
	}
}

type Scope struct {
	store       map[string]Object
	readonly    map[string]bool
	parentScope *Scope
	Writer      io.Writer
	CallStack   *CallStack

	//We need to use `Mutex`, because we added 'spawn'(multithread).
	//if not，when running `spawn`, there will be lot of errors, even core dump.
	//The reason is golang's map is not thread safe
	sync.RWMutex
}

func (s *Scope) Get(name string) (Object, bool) {
	s.RLock()
	defer s.RUnlock()

	//check the builtin class/annotation
	if val, ok := BuiltinClasses[name]; ok {
		return val, ok
	}

	obj, ok := s.store[name]
	if !ok && s.parentScope != nil {
		obj, ok = s.parentScope.Get(name)
	}
	return obj, ok
}

// Get all the keys of the scope.
func (s *Scope) GetKeys() []string {
	keys := make([]string, 0, len(s.store))
	for k := range s.store {
		keys = append(keys, k)
	}
	return keys
}

func (s *Scope) DebugPrint(indent string) {
	s.Lock()
	defer s.Unlock()

	for k, v := range s.store {
		fmt.Fprintf(s.Writer, "%s<%s> = <%s>  value.Type: %T\n", indent, k, v.Inspect(), v)
	}

	if s.parentScope != nil {
		fmt.Fprintf(s.Writer, "\n%sParentScope:\n", indent)
		s.parentScope.DebugPrint(indent + "  ")
	}

}

func (s *Scope) IsReadOnly(name string) bool {
	s.Lock()
	defer s.Unlock()

	//check if it is readonly
	_, ok := s.store[name]
	if ok {
		if s.readonly[name] {
			return true
		}
	} else {
		if s.parentScope != nil {
			return s.parentScope.IsReadOnly(name)
		}
	}

	return false
}

func (s *Scope) Set(name string, val Object) Object {
	s.Lock()
	defer s.Unlock()

	s.store[name] = val
	return val
}

func (s *Scope) SetConst(name string, val Object) Object {
	s.Lock()
	defer s.Unlock()

	s.store[name] = val
	s.readonly[name] = true //mark it as readonly

	return val
}

func (s *Scope) Reset(name string, val Object) (Object, bool) {
	s.Lock()
	defer s.Unlock()

	var ok bool
	_, ok = s.store[name]
	if ok {
		s.store[name] = val
	}

	if !ok && s.parentScope != nil {
		_, ok = s.parentScope.Reset(name, val)
	}

	if !ok {
		s.store[name] = val
		ok = true
	}
	return val, ok
}

func (s *Scope) CurrentFrame() *CallFrame {
	s.RLock()
	s.RUnlock()

	if s != nil {
		frames := s.CallStack.Frames
		if n := len(frames); n > 0 {
			return &frames[n-1] //return top
		}
	}
	return nil
}

// CallerFrame return caller's CallFrame
func (s *Scope) CallerFrame() *CallFrame {
	s.RLock()
	s.RUnlock()

	if s != nil {
		frames := s.CallStack.Frames
		if n := len(frames); n > 1 {
			return &frames[n-2] //return caller's frame
		}
	}
	return nil
}

var GlobalScopes map[string]Object = make(map[string]Object)
var GlobalMutex sync.RWMutex

func GetGlobalObj(name string) (Object, bool) {
	GlobalMutex.Lock()
	defer GlobalMutex.Unlock()

	obj, ok := GlobalScopes[name]
	return obj, ok
}

func SetGlobalObj(name string, Obj Object) {
	GlobalMutex.Lock()
	defer GlobalMutex.Unlock()

	GlobalScopes[name] = Obj
}
