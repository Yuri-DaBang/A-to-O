package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"originscript/eval"
	"originscript/lexer"
	"originscript/message"
	"originscript/parser"

	// "originscript/repl"
	"math/rand"
	"io"
	"os"
	"path/filepath"
	"strings"
	"regexp"
	"runtime"
	"time"

	"bytes"
	"math"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"errors"
	"net"
	"net/http"
	"net/url"
)

func runProgram(debug bool, filename string) {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	f, err := ioutil.ReadFile(wd + "/" + filename)
	if err != nil {
		fmt.Println("OriginScript: ", err.Error())
		os.Exit(1)
	}

	input := string(f)
	l := lexer.New(filename, input)

	p := parser.New(l, wd)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		for _, err := range p.Errors() {
			fmt.Println(err)
		}
		os.Exit(1)
	}
	scope := eval.NewScope(nil, os.Stdout)
	RegisterGoGlobals()

	if debug {
		eval.REPLColor = true
		eval.Dbg = eval.NewDebugger()
		eval.Dbg.SetFunctions(p.Functions)
		eval.Dbg.ShowBanner()

		dbgInfosArr := parser.SplitSlice(parser.DebugInfos) //[][]ast.Node
		eval.Dbg.SetDbgInfos(dbgInfosArr)
		//for idx, dbgInfos := range dbgInfosArr {
		//	for _, dbgInfo := range dbgInfos {
		//		fmt.Printf("idx:%d, Line:<%d-%d>, node.Type=%T, node=<%s>\n", idx, dbgInfo.Pos().Line, dbgInfo.End().Line, dbgInfo, dbgInfo.String())
		// 	}
		//}

		eval.MsgHandler = message.NewMessageHandler()
		eval.MsgHandler.AddListener(eval.Dbg)

	}

	result := eval.Eval(program, scope)
	if result.Type() == eval.ERROR_OBJ {
		fmt.Println(result.Inspect())
	}

	//	e := eval.Eval(program, scope)
	//	if e.Inspect() != "nil" {
	//		fmt.Println(e.Inspect())
	//	}
}
func runProgramDebugList(debug bool, filename string) {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	f, err := ioutil.ReadFile(wd + "/" + filename)
	if err != nil {
		fmt.Println("OriginScript: ", err.Error())
		os.Exit(1)
	}

	input := string(f)
	l := lexer.New(filename, input)

	p := parser.New(l, wd)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		for _, err := range p.Errors() {
			fmt.Println(err)
		}
		os.Exit(1)
	}
	scope := eval.NewScope(nil, os.Stdout)
	RegisterGoGlobals()

	if debug {
		eval.REPLColor = true
		eval.Dbg = eval.NewDebugger()
		eval.Dbg.SetFunctions(p.Functions)
		eval.Dbg.ShowBanner()

		dbgInfosArr := parser.SplitSlice(parser.DebugInfos) //[][]ast.Node
		eval.Dbg.SetDbgInfos(dbgInfosArr)
		for idx, dbgInfos := range dbgInfosArr {
			for _, dbgInfo := range dbgInfos {
				fmt.Printf("Idx:%d, Line:<%d-%d>, note.Type=%T, note=<%s>\n", idx, dbgInfo.Pos().Line, dbgInfo.End().Line, dbgInfo, dbgInfo.String())
		 	}
		}

		eval.MsgHandler = message.NewMessageHandler()
		eval.MsgHandler.AddListener(eval.Dbg)

	}

	result := eval.Eval(program, scope)
	if result.Type() == eval.ERROR_OBJ {
		fmt.Println(result.Inspect())
	}

	//	e := eval.Eval(program, scope)
	//	if e.Inspect() != "nil" {
	//		fmt.Println(e.Inspect())
	//	}
}

// Register go package methods/types
// Note here, we use 'gfmt', 'glog', 'gos' 'gtime', because in magpie
// we already have built in module 'fmt', 'log' 'os', 'time'.
// And Here we demonstrate the use of import go language's methods.
func RegisterGoGlobals() {
	eval.RegisterFunctions("console", map[string]interface{}{
		"Errorf":   fmt.Errorf,
		"Println":  fmt.Println,
		"Print":    fmt.Print,
		"Printf":   fmt.Printf,
		"Fprint":   fmt.Fprint,
		"Fprintln": fmt.Fprintln,
		"Fscan":    fmt.Fscan,
		"Fscanf":   fmt.Fscanf,
		"Fscanln":  fmt.Fscanln,
		"Scan":     fmt.Scan,
		"Scanf":    fmt.Scanf,
		"Scanln":   fmt.Scanln,
		"Sscan":    fmt.Sscan,
		"Sscanf":   fmt.Sscanf,
		"Sscanln":  fmt.Sscanln,
		"Sprint":   fmt.Sprint,
		"Sprintf":  fmt.Sprintf,
		"Sprintln": fmt.Sprintln,
	})

	eval.RegisterFunctions("log", map[string]interface{}{
		"Fatal":     log.Fatal,
		"Fatalf":    log.Fatalf,
		"Fatalln":   log.Fatalln,
		"Flags":     log.Flags,
		"Panic":     log.Panic,
		"Panicf":    log.Panicf,
		"Panicln":   log.Panicln,
		"Print":     log.Print,
		"Printf":    log.Printf,
		"Println":   log.Println,
		"SetFlags":  log.SetFlags,
		"SetOutput": log.SetOutput,
		"SetPrefix": log.SetPrefix,
	})

	eval.RegisterFunctions("os", map[string]interface{}{
		"Chdir":    os.Chdir,
		"Chmod":    os.Chmod,
		"Chown":    os.Chown,
		"Exit":     os.Exit,
		"Getpid":   os.Getpid,
		"Hostname": os.Hostname,
		"Environ":  os.Environ,
		"Getenv":   os.Getenv,
		"Setenv":   os.Setenv,
		"Create":   os.Create,
		"Open":     os.Open,
	})

	argsStart := 1
	if len(os.Args) > 2 {
		argsStart = 2
	}
	eval.RegisterVars("os", map[string]interface{}{
		"Args": os.Args[argsStart:],
	})

	eval.RegisterVars("runtime", map[string]interface{}{
		"OS":   runtime.GOOS,
		"ARCH": runtime.GOARCH,
	})

	eval.RegisterVars("time", map[string]interface{}{
		"Duration": time.Duration(0),
		"Ticker":   time.Ticker{},
		"Time":     time.Time{},
	})
	eval.RegisterFunctions("time", map[string]interface{}{
		"After":           time.After,
		"Sleep":           time.Sleep,
		"Tick":            time.Tick,
		"Since":           time.Since,
		"FixedZone":       time.FixedZone,
		"LoadLocation":    time.LoadLocation,
		"NewTicker":       time.NewTicker,
		"Date":            time.Date,
		"Now":             time.Now,
		"Parse":           time.Parse,
		"ParseDuration":   time.ParseDuration,
		"ParseInLocation": time.ParseInLocation,
		"Unix":            time.Unix,
		"AfterFunc":       time.AfterFunc,
		"NewTimer":        time.NewTimer,
		"Nanosecond":      time.Nanosecond,
		"Microsecond":     time.Microsecond,
		"Millisecond":     time.Millisecond,
		"Second":          time.Second,
		"Minute":          time.Minute,
		"Hour":            time.Hour,
	})

	eval.RegisterFunctions("bufio", map[string]interface{}{
		"NewWriter":     bufio.NewWriter,
		"NewReader":     bufio.NewReader,
		"NewReadWriter": bufio.NewReadWriter,
		"NewScanner":    bufio.NewScanner,
	})
	eval.RegisterFunctions("regex", map[string]interface{}{
		"Match":            regexp.Match,
		"MatchReader":      regexp.MatchReader,
		"MatchString":      regexp.MatchString,
		"QuoteMeta":        regexp.QuoteMeta,
		"Compile":          regexp.Compile,
		"CompilePOSIX":     regexp.CompilePOSIX,
		"MustCompile":      regexp.MustCompile,
		"MustCompilePOSIX": regexp.MustCompilePOSIX,
	})
	eval.RegisterFunctions("strings", map[string]interface{}{
		"Compare":        strings.Compare,
		"Contains":       strings.Contains,
		"ContainsAny":    strings.ContainsAny,
		"ContainsRune":   strings.ContainsRune,
		"Count":          strings.Count,
		"EqualFold":      strings.EqualFold,
		"Fields":         strings.Fields,
		"FieldsFunc":     strings.FieldsFunc,
		"HasPrefix":      strings.HasPrefix,
		"HasSuffix":      strings.HasSuffix,
		"Index":          strings.Index,
		"IndexAny":       strings.IndexAny,
		"IndexByte":      strings.IndexByte,
		"IndexFunc":      strings.IndexFunc,
		"IndexRune":      strings.IndexRune,
		"Join":           strings.Join,
		"LastIndex":      strings.LastIndex,
		"LastIndexAny":   strings.LastIndexAny,
		"LastIndexByte":  strings.LastIndexByte,
		"LastIndexFunc":  strings.LastIndexFunc,
		"Map":            strings.Map,
		"Repeat":         strings.Repeat,
		"Replace":        strings.Replace,
		"ReplaceAll":     strings.ReplaceAll,
		"Split":          strings.Split,
		"SplitAfter":     strings.SplitAfter,
		"SplitAfterN":    strings.SplitAfterN,
		"SplitN":         strings.SplitN,
		"Title":          strings.Title,
		"ToLower":        strings.ToLower,
		"ToLowerSpecial": strings.ToLowerSpecial,
		"ToTitle":        strings.ToTitle,
		"ToTitleSpecial": strings.ToTitleSpecial,
		"ToUpper":        strings.ToUpper,
		"ToUpperSpecial": strings.ToUpperSpecial,
		"Trim":           strings.Trim,
		"TrimFunc":       strings.TrimFunc,
		"TrimLeft":       strings.TrimLeft,
		"TrimLeftFunc":   strings.TrimLeftFunc,
		"TrimPrefix":     strings.TrimPrefix,
		"TrimRight":      strings.TrimRight,
		"TrimRightFunc":  strings.TrimRightFunc,
		"TrimSpace":      strings.TrimSpace,
		"TrimSuffix":     strings.TrimSuffix,
	})

	eval.RegisterFunctions("bytes", map[string]interface{}{
		"Compare":         bytes.Compare,
		"Contains":        bytes.Contains,
		"ContainsAny":     bytes.ContainsAny,
		"ContainsRune":    bytes.ContainsRune,
		"Count":           bytes.Count,
		"Equal":           bytes.Equal,
		"EqualFold":       bytes.EqualFold,
		"Fields":          bytes.Fields,
		"FieldsFunc":      bytes.FieldsFunc,
		"HasPrefix":       bytes.HasPrefix,
		"HasSuffix":       bytes.HasSuffix,
		"Index":           bytes.Index,
		"IndexAny":        bytes.IndexAny,
		"IndexByte":       bytes.IndexByte,
		"IndexFunc":       bytes.IndexFunc,
		"IndexRune":       bytes.IndexRune,
		"Join":            bytes.Join,
		"LastIndex":       bytes.LastIndex,
		"LastIndexAny":    bytes.LastIndexAny,
		"LastIndexByte":   bytes.LastIndexByte,
		"LastIndexFunc":   bytes.LastIndexFunc,
		"Repeat":          bytes.Repeat,
		"Replace":         bytes.Replace,
		"Split":           bytes.Split,
		"SplitAfter":      bytes.SplitAfter,
		"SplitAfterN":     bytes.SplitAfterN,
		"SplitN":          bytes.SplitN,
		"Title":           bytes.Title,
		"ToLower":         bytes.ToLower,
		"ToTitle":         bytes.ToTitle,
		"ToUpper":         bytes.ToUpper,
		"Trim":            bytes.Trim,
		"TrimFunc":        bytes.TrimFunc,
		"TrimLeft":        bytes.TrimLeft,
		"TrimLeftFunc":    bytes.TrimLeftFunc,
		"TrimPrefix":      bytes.TrimPrefix,
		"TrimRight":       bytes.TrimRight,
		"TrimRightFunc":   bytes.TrimRightFunc,
		"TrimSpace":       bytes.TrimSpace,
		"TrimSuffix":      bytes.TrimSuffix,
	})

	eval.RegisterFunctions("cryptomd5", map[string]interface{}{
		"New": md5.New,
		"Sum": md5.Sum,
	})

	eval.RegisterFunctions("cryptosha1", map[string]interface{}{
		"New": sha1.New,
		"Sum": sha1.Sum,
	})

	eval.RegisterFunctions("cryptosha256", map[string]interface{}{
		"New": sha256.New,
		"Sum": sha256.Sum256,
	})

	eval.RegisterFunctions("cryptosha512", map[string]interface{}{
		"New": sha512.New,
		"Sum": sha512.Sum512,
	})

	eval.RegisterFunctions("base64", map[string]interface{}{
		"NewDecoder": base64.NewDecoder,
		"NewEncoder": base64.NewEncoder,
		"StdEncoding": base64.StdEncoding,
		"URLEncoding": base64.URLEncoding,
	})

	eval.RegisterFunctions("json", map[string]interface{}{
		"Marshal":        json.Marshal,
		"Unmarshal":      json.Unmarshal,
		"NewDecoder":     json.NewDecoder,
		"NewEncoder":     json.NewEncoder,
		"MarshalIndent":  json.MarshalIndent,
	})

	eval.RegisterFunctions("xml", map[string]interface{}{
		"Marshal":        xml.Marshal,
		"Unmarshal":      xml.Unmarshal,
		"NewDecoder":     xml.NewDecoder,
		"NewEncoder":     xml.NewEncoder,
		"MarshalIndent":  xml.MarshalIndent,
	})

	eval.RegisterFunctions("errors", map[string]interface{}{
		"New": errors.New,
		"Is":  errors.Is,
		"As":  errors.As,
	})

	eval.RegisterFunctions("io", map[string]interface{}{
		"Copy":           io.Copy,
		"CopyBuffer":     io.CopyBuffer,
		"CopyN":          io.CopyN,
		"ReadAtLeast":    io.ReadAtLeast,
		"ReadFull":       io.ReadFull,
		"WriteString":    io.WriteString,
	})

	eval.RegisterFunctions("ioutil", map[string]interface{}{
		"ReadAll":  ioutil.ReadAll,
		"ReadFile": ioutil.ReadFile,
		"ReadDir": ioutil.ReadDir,
		"WriteFile": ioutil.WriteFile,
		"TempDir": ioutil.TempDir,
		"TempFile": ioutil.TempFile,
	})

	eval.RegisterFunctions("math", map[string]interface{}{
		"Abs": math.Abs,
		"Acos": math.Acos,
		"Asin": math.Asin,
		"Atan": math.Atan,
		"Atan2": math.Atan2,
		"Ceil": math.Ceil,
		"Cos": math.Cos,
		"Cosh": math.Cosh,
		"Exp": math.Exp,
		"Exp2": math.Exp2,
		"Expm1": math.Expm1,
		"Floor": math.Floor,
		"Log": math.Log,
		"Log10": math.Log10,
		"Log1p": math.Log1p,
		"Log2": math.Log2,
		"Max": math.Max,
		"Min": math.Min,
		"Pow": math.Pow,
		"Sin": math.Sin,
		"Sinh": math.Sinh,
		"Sqrt": math.Sqrt,
		"Tan": math.Tan,
		"Tanh": math.Tanh,
	})

	eval.RegisterFunctions("rand", map[string]interface{}{
	    "New":         rand.New,
		"NewSource":   rand.NewSource,
		"ExpFloat64":  rand.ExpFloat64,	
		"Float32": rand.Float32,
		"Float64": rand.Float64,
		"Int": rand.Int,
		"Int31": rand.Int31,
		"Int31n": rand.Int31n,
		"Int63": rand.Int63,
		"Int63n": rand.Int63n,
		"Intn": rand.Intn,
		"NormFloat64": rand.NormFloat64,
		"Perm": rand.Perm,
		"Seed": rand.Seed,
		"Uint32": rand.Uint32,
		"Uint64": rand.Uint64,
		
	})

	eval.RegisterFunctions("net", map[string]interface{}{
		"CIDRMask": net.CIDRMask,
		"Dial": net.Dial,
		"DialTimeout": net.DialTimeout,
		"FileConn": net.FileConn,
		"FileListener": net.FileListener,
		"FilePacketConn": net.FilePacketConn,
		"IPv4": net.IPv4,
		"IPv4Mask": net.IPv4Mask,
		"Listen": net.Listen,
		"LookupAddr": net.LookupAddr,
		"LookupCNAME": net.LookupCNAME,
		"LookupHost": net.LookupHost,
		"ParseCIDR": net.ParseCIDR,
		"ParseIP": net.ParseIP,
		"SplitHostPort": net.SplitHostPort,
		"JoinHostPort": net.JoinHostPort,
	})

	eval.RegisterFunctions("nethttp", map[string]interface{}{
		"clone": http.Get,
		"Post": http.Post,
		"PostForm": http.PostForm,
		"Head": http.Head,
		"ListenAndServe": http.ListenAndServe,
		"NewRequest": http.NewRequest,
		"ReadResponse": http.ReadResponse,
		"Redirect": http.Redirect,
		"Serve": http.Serve,
		"ServeContent": http.ServeContent,
		"ServeFile": http.ServeFile,
		"SetCookie": http.SetCookie,
	})

	eval.RegisterFunctions("neturl", map[string]interface{}{
		"Parse": url.Parse,
		"ParseRequestURI": url.ParseRequestURI,
		"QueryEscape": url.QueryEscape,
		"QueryUnescape": url.QueryUnescape,
	})

	eval.RegisterFunctions("filepath", map[string]interface{}{
		"Abs": filepath.Abs,
		"Base": filepath.Base,
		"Clean": filepath.Clean,
		"Dir": filepath.Dir,
		"Ext": filepath.Ext,
		"FromSlash": filepath.FromSlash,
		"Glob": filepath.Glob,
		"IsAbs": filepath.IsAbs,
		"Join": filepath.Join,
		"Match": filepath.Match,
		"Rel": filepath.Rel,
		"Split": filepath.Split,
		"ToSlash": filepath.ToSlash,
		"VolumeName": filepath.VolumeName,
		"Walk": filepath.Walk,
	})
}
func showHelp(topic string) {
    showHelpRecursive(topic)
}
func showHelpRecursive(item string) {

	// items := ["errors|error","keywords","libs"]
	if item == "***"{
		//fmt.Printf("showing help of %s\n",item)
		fmt.Println("   pack:")
		fmt.Println("\tDescription:")
		fmt.Println("\t   TO MANAGE PACKAGES AND FILES RELATED TO `/opkg/` DIR.")
		fmt.Println("\tUsage:")
		fmt.Println(fmt.Sprintf("\t   %-10s : %-62s : Usage == $EXE pack init", "pack init", "Init the pack file."))
		fmt.Println(fmt.Sprintf("\t   %-10s : %-62s : Usage == $EXE pack list", "pack list", "List of modules with descriptions."))
		fmt.Println(fmt.Sprintf("\t   %-10s : %-62s : Usage == $EXE pack clone $MODULE_NAME", "pack clone", "To clone a package in your env."))
		fmt.Println(fmt.Sprintf("\t   %-10s : %-62s : Usage == $EXE pack del $MODULE_NAME", "pack del", "To remove a package from your env."))
		fmt.Println(fmt.Sprintf("\t   %-10s : %-62s : Usage == $EXE pack check", "pack check", "To check consistency between origion.pack and opkg directory."))
	
		fmt.Println("   Run:")
		fmt.Println("\tDescription:")
		fmt.Println("\t   TO RUN A ORIGINSCRIPT FILE WHICH IS PRESENT IN `/opkg/` DIR.")
		fmt.Println("\tUsage:")
		fmt.Println("\t   run $FILE_NAME       : Run the aeroscript codefile.                : Usage == $EXE run $FILE_NAME")
		fmt.Println("\t   run debug $FILE_NAME : Run the aeroscript codefile in debug mode.  : Usage == $EXE run debug $FILE_NAME")

		fmt.Println("   Lun:")
		fmt.Println("\tDescription:")
		fmt.Println("\t   TO RUN A ORIGINSCRIPT FILE WHICH CAN BE LOCATED LOCALLY ANYWARE ON THE PC.")
		fmt.Println("\tUsage:")
		fmt.Println("\t   lun $FILE_NAME       : Run the aeroscript codefile.                : Usage == $EXE lun $FILE_NAME")
		fmt.Println("\t   lun debug $FILE_NAME : Run the aeroscript codefile in debug mode.  : Usage == $EXE lun debug $FILE_NAME")

		fmt.Println("   Others:")
		fmt.Println("\t-h error|errors : List of errors with descriptions.  : Usage == $EXE -h errors")

		fmt.Println("\t   Errors:")
		fmt.Println("\t      e3209 : Import Error  : Indicates that, the lib you imported, could not be imported!")
		fmt.Println("\t      e3301 : Syntax Error  : Indicates that, your code has a syntax mistake!")
		fmt.Println("\t      eUDE  : Eval Error    : This is an Eval error, occurred at runtime!")
	} else if item == "except" || item == "excepts" {
		fmt.Printf("showing list of errors\n")
		fmt.Println("\te3209 : Import Error  : Indicates that, the lib you imported, could not be imported!")
		fmt.Println("\te3301 : Syntax Error  : Indicates that, your code has a syntax mistake!")
		fmt.Println("\teUDE  : Eval Error    : This is an Eval error, occurred at runtime!")
	} else if item == "pack" {
		fmt.Println("   Usage of pack:")
		fmt.Println(fmt.Sprintf("\t%-10s : %-62s : Usage == $EXE pack init", "pack init", "Init the pack file."))
		fmt.Println(fmt.Sprintf("\t%-10s : %-62s : Usage == $EXE pack list", "pack list", "List of modules with descriptions."))
		fmt.Println(fmt.Sprintf("\t%-10s : %-62s : Usage == $EXE pack clone $MODULE_NAME", "pack clone", "To clone a package in your env."))
		fmt.Println(fmt.Sprintf("\t%-10s : %-62s : Usage == $EXE pack del $MODULE_NAME", "pack del", "To remove a package from your env."))
		fmt.Println(fmt.Sprintf("\t%-10s : %-62s : Usage == $EXE pack check", "pack check", "To check consistency between origion.pack and opkg directory."))
	} else if item == "lun" {
		fmt.Println("   Lun:")
		fmt.Println("\tDescription:")
		fmt.Println("\t   TO RUN A ORIGINSCRIPT FILE WHICH CAN BE LOCATED LOCALLY ANYWARE ON THE PC.")
		fmt.Println("\tUsage:")
		fmt.Println("\t   lun $FILE_NAME       : Run the aeroscript codefile.                : Usage == $EXE lun $FILE_NAME")
		fmt.Println("\t   lun debug $FILE_NAME : Run the aeroscript codefile in debug mode.  : Usage == $EXE lun debug $FILE_NAME")
	} else if item == "run" {
		fmt.Println("   Run:")
		fmt.Println("\tDescription:")
		fmt.Println("\t   TO RUN A ORIGINSCRIPT FILE WHICH IS PRESENT IN `/opkg/` DIR.")
		fmt.Println("\tUsage:")
		fmt.Println("\t   run $FILE_NAME       : Run the aeroscript codefile.                : Usage == $EXE run $FILE_NAME")
		fmt.Println("\t   run debug $FILE_NAME : Run the aeroscript codefile in debug mode.  : Usage == $EXE run debug $FILE_NAME")
	} else {
		showHelp("***")
		//fmt.Println("OriginScript: Usage: $AERO_SCRIPT_EXE_PATH -h $THING\n hint: type `$AERO_SCRIPT_EXE_PATH -h /list/` for list of items.")
	}
}

///
/// pack STUFF
/// 

// initializeModFile creates a new 'origion.pack' file if it does not exist.
func initializeModFile() {
    filename := "origion.pack"
    if _, err := os.Stat(filename); err == nil {
        fmt.Println("OriginScript: pack: Error: origion.pack file already exists.")
        return
    }

    file, err := os.Create(filename)
    if err != nil {
        fmt.Println("OriginScript: pack: Error creating origion.pack file.\n\t└─Error Message:", err)
        return
    }
    defer file.Close()

    // Write "origion.pack" into the file
    if _, err := file.WriteString(""); err != nil {
        fmt.Println("OriginScript: pack: Error writing to origion.pack file.\n\t└─Error Message:", err)
        return
    }

    fmt.Println("origion.pack file created and initialized successfully.")
}
// listScripts reads and prints the content of 'origion.pack' file.
func listScripts() {
	filename := "origion.pack"
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("OriginScript: pack: Error opening origion.pack file.\n\t└─Error Message:", err)
		return
	}
	defer file.Close()

	fmt.Println("   Scripts:")
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fmt.Println("\t"+scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("OriginScript: pack: Error reading origion.pack file. \n\t└─Error Message:", err)
	}
}
// isScriptInModFile checks if a script is already listed in the origion.pack file.
func isScriptInModFile(modFilename, scriptName string) bool {
	file, err := os.Open(modFilename)
	if err != nil {
		fmt.Println("OriginScript: pack: Error opening origion.pack file.\n\t└─Error Message:", err)
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if scanner.Text() == scriptName {
			return true
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("OriginScript: pack: Error reading origion.pack file.\n\t└─Error Message:", err)
	}
	return false
}
func addScript(scriptPattern string) {
    modFilename := "origion.pack"
    scriptsDir := "opkg"
    sourceDir := "C:/Program Files/OrigionScript/libs"

    // Glob for files matching the pattern
    sourceFiles, err := filepath.Glob(filepath.Join(sourceDir, scriptPattern))
    if err != nil {
        fmt.Println("OriginScript: pack: Error finding scripts.\n\t└─Error Message:", err)
        return
    }

    // Process each matched script
    for _, sourcePath := range sourceFiles {
        // Determine scriptName relative to sourceDir
        scriptName, err := filepath.Rel(sourceDir, sourcePath)
        if err != nil {
            fmt.Println("OriginScript: pack: Error getting script name.\n\t└─Error Message:", err)
            continue
        }

        // Path to the destination directory
        destDir := filepath.Join(scriptsDir, filepath.Dir(scriptName))

        // Create destination directory if it doesn't exist
        if _, err := os.Stat(destDir); os.IsNotExist(err) {
            if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
                fmt.Println("OriginScript: pack: Error creating destination directory.\n\t└─Error Message:", err)
                continue
            }
        }

        // Path to the script in the destination directory
        destPath := filepath.Join(scriptsDir, scriptName)

        // Copy the script to opkg directory
        if err := copyFile(sourcePath, destPath); err != nil {
            fmt.Println("OriginScript: pack: Error copying script.\n\t└─Error Message:", err)
            output := fmt.Sprintf("%s/*", sourcePath)
			addScript(output)
			continue
        }

        // Append the script name to origion.pack
        file, err := os.OpenFile(modFilename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
        if err != nil {
            fmt.Println("OriginScript: pack: Error opening origion.pack file.\n\t└─Error Message:", err)
            continue
        }
        defer file.Close()

        if _, err := file.WriteString(scriptName + "\n"); err != nil {
            fmt.Println("OriginScript: pack: Error writing to origion.pack file.\n\t└─Error Message:", err)
            continue
        }

        fmt.Println("OriginScript: pack: Script added successfully:", scriptName)
    }
}
// copyFile copies a file from src to dst.
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
// removeScript removes a script or directory from 'origion.pack' and deletes it from 'opkg' directory.
func removeScript(scriptName string) {
    modFilename := "origion.pack"
    scriptsDir := "opkg"

    // Path to the script or directory in the opkg directory
    scriptPath := filepath.Join(scriptsDir, scriptName)
    if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
        fmt.Println("OriginScript: pack: Error: Script or directory does not exist.\n\t└─Error Message:", err)
        return
    }

    // Remove the script or directory from opkg directory
    if err := os.RemoveAll(scriptPath); err != nil {
        fmt.Println("OriginScript: pack: Error removing script or directory.\n\t└─Error Message:", err)
        return
    }

    // Remove the script name from origion.pack
    input, err := os.ReadFile(modFilename)
    if err != nil {
        fmt.Println("OriginScript: pack: Error reading origion.pack file.\n\t└─Error Message:", err)
        return
    }

    lines := strings.Split(string(input), "\n")
    var output []string
    for _, line := range lines {
        if line != scriptName && line != "" {
            output = append(output, line)
        }
    }

    if err := os.WriteFile(modFilename, []byte(strings.Join(output, "\n")), 0644); err != nil {
        fmt.Println("OriginScript: pack: Error writing to origion.pack file.\n\t└─Error Message:", err)
        return
    }

    fmt.Println("OriginScript: pack: Script or directory removed successfully.")
}
// checkScripts compares the scripts listed in 'origion.pack' with those present in the 'aeromod' directory.
func checkScripts() {
	modFilename := "origion.pack"
	scriptsDir := "opkg"

	// Read scripts listed in origion.pack
	modScripts := make(map[string]bool)
	modFile, err := os.Open(modFilename)
	if err != nil {
		fmt.Println("OriginScript: pack: Error opening origion.pack file.\n\t└─Error Message:", err)
		return
	}
	defer modFile.Close()

	modScanner := bufio.NewScanner(modFile)
	for modScanner.Scan() {
		modScripts[modScanner.Text()] = true
	}
	if err := modScanner.Err(); err != nil {
		fmt.Println("OriginScript: pack: Error reading origion.pack file.\n\t└─Error Message:", err)
		return
	}

	// Read scripts present in aeromod directory
	dirScripts := make(map[string]bool)
	dirEntries, err := os.ReadDir(scriptsDir)
	if err != nil {
		fmt.Println("OriginScript: pack: Error reading aeromod directory.\n\t└─Error Message:", err)
		return
	}

	for _, entry := range dirEntries {
		dirScripts[entry.Name()] = true
	}

	// Check for mismatches
	var modMissing, dirMissing bool
	for script := range modScripts {
		if script == "origion.pack" && !dirScripts[script] {
			fmt.Println("OriginScript: pack: Error: Script origion.pack listed in origion.pack but not found in aeromod directory.")
			modMissing = true
		}
	}

	fmt.Println("OriginScript: pack:")
	for script := range dirScripts {
		if script != "origion.pack" && !modScripts[script] {
			fmt.Printf("                   Error: Script %s found in aeromod directory but not listed in origion.pack\n", script)
			dirMissing = true
		}
	}

	if !modMissing && !dirMissing {
		fmt.Println("                   All scripts match between origion.pack and aeromod directory.")
	}
}

///
/// MAIN STUFF
///

func main() {
	version := "0.1i"
	args := os.Args[1:]
	//We must reset `os.Args`, or the `flag` module will not functioning correctly
	os.Args = os.Args[1:]
	if len(args) == 0 {

		fmt.Println("OriginScript: version[`",version,"`] , Usage[`pack`,`--debug`,`--help`,`--lun`,`--run`,`--pack`]")

		showHelp("***")
		//repl.Start(os.Stdout, true)
	} else {
		if len(args) >= 1 {
			if args[0] == "-d" || args[0] == "--debug" { // debug
				if len(args) >= 2 {
					if args[1] != "" {
						runProgramDebugList(true, args[1])
					} else {
						fmt.Println("OriginScript: Usage: $AERO_SCRIPT_EXE_PATH %s $FILE_NAME.aero",args[0])
						os.Exit(1)
					}
				} else {
					fmt.Printf("OriginScript: Usage: $AERO_SCRIPT_EXE_PATH %s $FILE_NAME.aero\n",args[0])
					//os.Exit(1)
				}
			} else if args[0] == "-h" || args[0] == "--help" || args[0] == "help" || args[0] == "/h" || args[0] == "\\h" { 
				fmt.Println("OriginScript: version[`",version,"`]")
				if len(args) >= 2 {
					if args[1] != "" {
						showHelp(args[1])
					} else {
						fmt.Printf("OriginScript: Usage: $AERO_SCRIPT_EXE_PATH %s $THING\n hint: type `$AERO_SCRIPT_EXE_PATH %s /list/` for list of items.\n",args[0],args[0])
						//os.Exit(1)
					}
				} else {
					showHelp("***")
					//fmt.Printf("OriginScript: Usage: $AERO_SCRIPT_EXE_PATH %s $THING\n hint: type `$AERO_SCRIPT_EXE_PATH %s /list/` for list of items.\n",args[0],args[0])
					//os.Exit(1)
				}
			} else if args[0] == "-l" || args[0] == "--lun" { 
				if len(args) >= 1 {
					if args[1] == "debug" {
						runProgram(true, args[2])
					} else {
						runProgram(false, args[1])
						//fmt.Printf("OriginScript: Usage: $AERO_SCRIPT_EXE_PATH %s $FILE_NAME.aero\n",args[0])
						//os.Exit(1)
					}
				} else {
					//showHelp("***")
					fmt.Printf("OriginScript: Usage: $AERO_SCRIPT_EXE_PATH %s $FILE_NAME.aero\\n",args[0])
					//os.Exit(1)
				}
			} else if args[0] == "-r" || args[0] == "--run" { 
				if len(args) >= 1 {
					if args[1] == "debug" {
						formatted := fmt.Sprintf("/opkg/%s", args[2])
						runProgram(true, formatted)
					} else {
						formatted := fmt.Sprintf("/opkg/%s", args[1])
						runProgram(false, formatted)
						//fmt.Printf("OriginScript: Usage: $AERO_SCRIPT_EXE_PATH %s $FILE_NAME.aero\n",args[0])
						//os.Exit(1)
					}
				} else {
					//showHelp("***")
					fmt.Printf("OriginScript: Usage: $AERO_SCRIPT_EXE_PATH %s $FILE_NAME.aero\\n",args[0])
					//os.Exit(1)
				}
			} else if args[0] == "-p" || args[0] == "--pack" {
				fmt.Println("OriginScript: version<",version,">")
				if len(args) < 2 {
					showHelp("pack")
					return
				}

				switch args[1] {
					case "init":
						initializeModFile()

					case "list":
						listScripts()

					case "clone":
						if len(args) < 3 {
							fmt.Println("OriginScript: Error: Script name is required for 'clone' command")
							return
						}
						scriptName := args[2]
						addScript(scriptName)

					case "del":
						if len(args) < 3 {
							fmt.Println("OriginScript: Error: Script name is required for 'del' command")
							return
						}
						scriptName := args[2]
						removeScript(scriptName)

					case "check":
						checkScripts()

					default:
						fmt.Println("OriginScript: Error: Unknown subcommand")
						showHelp("pack")
				}
			} else {
				showHelp("***")
				//fmt.Println("OriginScript: Usage: $AERO_SCRIPT_EXE_PATH -$OPTION $EXTRA_ARGS_IF_REQUIRED")
			}
		} else {
			runProgram(false, args[0])
		}
	}
}


