// Package doc extracts source code documentation from a Magpie AST.
package doc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"originscript/ast"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"unicode"
)

var (
	regexpType            = regexp.MustCompile(`^\{(.+)\}$`)
	regExample            = regexp.MustCompile(`@example[\r\n]([^@]+)@[\r\n]`) //@examples
	regNote               = regexp.MustCompile(`@note[\r\n]([^@]+)@[\r\n]`)    //@note
	regWarning            = regexp.MustCompile(`@warn[\r\n]([^@]+)@[\r\n]`)    //@warn
	regExpShowSourceBegin = regexp.MustCompile(`(<p>SHOWSOURCE_PLACEHOLDER_LINE_BEGIN(.*?)</p>)`)

	//table of contents
	toc = `<p><div><a id="toc-button" class="toc-button" onclick="toggle_toc()"><span id="btn-text">&#x25BC;</span>&nbsp;Table of Contents</a><div id="table-of-contents" style="display:none;">`

	//PlaceHolder line, used only in html output.
	PlaceHolderTOC           = "<p>__TOC_PLACEHOLDER_LINE_END__</p>"
	PlaceHolderShowSourceEnd = "<p>__SHOWSOURCE_PLACEHOLDER_LINE_END__</p>"
	//Line number regexp. Mainly for adding line numbers for code blocks.
	regLinePlaceHolder = regexp.MustCompile(`</pre></div>[\r\n]<p>__LINENUMBER_PLACEHOLDER_LINE__(\d+)</p>`)

	Cfg = Config{}
)

type Config struct {
	ShowSrcComment int    //1:if show source comment
	GenHTML        int    //1: if generate html-style document
	CssStyle       int    // default css style to use.(See css.go for builtin css styles)
	CssContents    string //User supplied css file contents for styling generated html file
}

// File is the documentation for an entire magpie file.
type File struct {
	Name    string //FileName
	Classes []*Classes
	Enums   []*Value
	Lets    []*Value
	Consts  []*Value
	Funcs   []*Function
	GenHTML int
}

/* Classes is the documention for a class */
type Classes struct {
	Value *Value
	Props []*Value    //Properties
	Lets  []*Value    //Let-statements
	Funcs []*Function //Function
}

/* Classes is the documention for a function */
type Function struct {
	Value   *Value
	Params  []*FuncInfo
	Returns []*FuncInfo
}

//function information(@param/@return/@returns part)
type FuncInfo struct {
	Name string //parameter name if @param, or else ""
	Type string //type
	Desc string //description
}

//Value is the documentation for a (possibly grouped) enums, lets, functions, or class declaration.
type Value struct {
	Name string //name
	Doc  string //comment
	Text string //declaration text

	ShowSrc  int    //should source or not, 1: show
	Src      string //Source code text
	SrcLines int    //number of lines of `Src`.
	GenHTML  int    //1: if generate html-style document
}

// Request for github REST API
// URL : https://developer.github.com/v3/markdown/
type Request struct {
	Text    string `json:"text"`
	Mode    string `json:"mode"`
	Context string `json:"context"`
}

func New(name string, program *ast.Program) *File {
	var classes []*ast.ClassStatement
	var enums []*ast.EnumStatement
	var lets []*ast.LetStatement
	var consts []*ast.ConstStatement
	var funcs []*ast.FunctionStatement

	fh, _ := os.Open(name)
	defer fh.Close()
	for _, statement := range program.Statements {
		switch s := statement.(type) {
		case *ast.ClassStatement:
			if s.Doc != nil {
				classes = append(classes, s)
			}
		case *ast.EnumStatement:
			if s.Doc != nil {
				enums = append(enums, s)
			}
		case *ast.LetStatement:
			if s.Doc != nil {
				lets = append(lets, s)
			}
		case *ast.ConstStatement:
			if s.Doc != nil {
				consts = append(consts, s)
			}
		case *ast.FunctionStatement:
			if s.Doc != nil {
				funcs = append(funcs, s)
			}
		}
	}

	return &File{
		Name:    filepath.Base(name),
		Classes: sortedClasses(classes, fh),
		Enums:   sortedEnums(enums, fh),
		Lets:    sortedLets(lets, fh),
		Consts:  sortedConsts(consts, fh),
		Funcs:   sortedFuncs(funcs, fh),
		GenHTML: Cfg.GenHTML,
	}
}

// ----------------------------------------------------------------------------
// Markdown document generator

// MdDocGen generates markdown documentation from doc.File.
func MdDocGen(f *File) string {
	var buffer bytes.Buffer
	tmpl, _ := template.New("baseTpl").Funcs(funcs).Parse(templs[0])
	for _, templ := range templs[1:] {
		tmpl, _ = template.Must(tmpl.Clone()).Parse(templ)
	}
	err := tmpl.Execute(&buffer, f)
	if err != nil {
		fmt.Printf("Template executing error:%v\n", err)
	}
	return normalize(buffer.String())
}

func normalize(doc string) string {
	nlReplace := regexp.MustCompile(`\n(\s)+\n`)
	trimCodes := regexp.MustCompile("\n{2,}```")
	doc = nlReplace.ReplaceAllString(doc, "\n\n")
	doc = trimCodes.ReplaceAllString(doc, "\n```")

	return doc
}

// ----------------------------------------------------------------------------
// Html document generator(using github REST API)

// HtmlDocGen generates html documentation from a markdown file.
func HtmlDocGen(content string, file *File) string {
	buf, err := json.Marshal(Request{
		Text:    string(content),
		Mode:    "gfm",
		Context: "github/gollum",
	})
	if err != nil {
		fmt.Errorf("Marshaling request failed, reason=%v\n", err)
		return ""
	}

	resp, err := http.Post("https://api.github.com/markdown", "application/json", bytes.NewBuffer(buf))
	if err != nil {
		fmt.Errorf("Request for github failed, reason:%v\n", err)
		return ""
	}
	defer resp.Body.Close() //must close the 'Body'

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Errorf("Response read failed, reason:%v\n", err)
		return ""
	}

	var out bytes.Buffer
	//doc type
	out.WriteString("<!DOCTYPE html>")
	//head
	out.WriteString("<head><meta http-equiv=\"Content-Type\" content=\"text/html; charset=utf-8\">")
	out.WriteString(`<script type="text/javascript">
function toggle_toc() {
    var toc=document.getElementById('table-of-contents');
    var btn=document.getElementById('btn-text');
    toc.style.display=(toc.style.display=='none')?'block':'none';
    btn.innerHTML=(toc.style.display=='none')?'&#x25BC;':'&#x25B2;';
}
function toggle_source(name) {
    var showsource=document.getElementById('showsource-' + name);
    var btn=document.getElementById('btn-text-'+ name);
    showsource.style.display=(showsource.style.display=='none')?'block':'none';
    btn.innerHTML=(showsource.style.display=='none')?'&#x25BC;':'&#x25B2;';
}
</script>`)

	out.WriteString("</head>")

	//css style
	var css string
	out.WriteString("<style>")
	css = Cfg.CssContents
	if Cfg.CssContents == "" {
		css = fmt.Sprintf(cssGeneral, strArr2IntfArr(BuiltinCssStyle[Cfg.CssStyle])...)
	}
	out.WriteString(css)
	out.WriteString("</style>")

	//body
	out.WriteString(`<body><div class="readme"><article class="markdown-body">`)
	out.WriteString(string(body))
	out.WriteString("</article></div></body>")

	//The github returned html's inner linking is not working,
	//so we need to fix this.
	return postProcessingHtml(out.String(), file)
}

func postProcessingHtml(htmlStr string, file *File) string {
	html := htmlStr
	//--------------------------------------------
	// Fix link
	//--------------------------------------------
	for _, enum := range file.Enums {
		enumName := enum.Name
		src := fmt.Sprintf("<h3>%s</h3>", enumName)
		dest := fmt.Sprintf(`<h3 id="%s">%s</h3>`, SanitizedAnchorName(enumName), enumName)
		html = strings.Replace(html, src, dest, -1)
	}
	for _, let := range file.Lets {
		letName := let.Name
		src := fmt.Sprintf("<h3>%s</h3>", letName)
		dest := fmt.Sprintf(`<h3 id="%s">%s</h3>`, SanitizedAnchorName(letName), letName)
		html = strings.Replace(html, src, dest, -1)
	}
	for _, cons := range file.Consts {
		consName := cons.Name
		src := fmt.Sprintf("<h3>%s</h3>", consName)
		dest := fmt.Sprintf(`<h3 id="%s">%s</h3>`, SanitizedAnchorName(consName), consName)
		html = strings.Replace(html, src, dest, -1)
	}
	for _, fn := range file.Funcs {
		fnName := fn.Value.Name
		src := fmt.Sprintf("<h3>%s</h3>", fnName)
		dest := fmt.Sprintf(`<h3 id="%s">%s</h3>`, SanitizedAnchorName(fnName), fnName)
		html = strings.Replace(html, src, dest, -1)
	}

	for _, cls := range file.Classes {
		clsName := cls.Value.Name
		src := fmt.Sprintf("<h3>%s</h3>", clsName)
		dest := fmt.Sprintf(`<h3 id="%s">%s</h3>`, SanitizedAnchorName(clsName), clsName)
		html = strings.Replace(html, src, dest, -1)

		for _, prop := range cls.Props {
			propName := prop.Name
			src := fmt.Sprintf("<h5>%s</h5>", propName)
			dest := fmt.Sprintf(`<h5 id="%s">%s</h5>`, SanitizedAnchorName(propName), propName)
			html = strings.Replace(html, src, dest, -1)
		}
		for _, let := range cls.Lets {
			letName := let.Name
			src := fmt.Sprintf("<h5>%s</h5>", letName)
			dest := fmt.Sprintf(`<h5 id="%s">%s</h5>`, SanitizedAnchorName(letName), letName)
			html = strings.Replace(html, src, dest, -1)
		}
		for _, fn := range cls.Funcs {
			fnName := fn.Value.Name
			src := fmt.Sprintf("<h5>%s</h5>", fnName)
			dest := fmt.Sprintf(`<h5 id="%s">%s</h5>`, SanitizedAnchorName(fnName), fnName)
			html = strings.Replace(html, src, dest, -1)
		}
	}

	//--------------------------------------------
	// Replace placeholder
	//--------------------------------------------
	//Replace 'Table of Contents'
	html = strings.Replace(html, "<h1>Table of Contents</h1>", toc, 1)
	html = strings.Replace(html, PlaceHolderTOC, "</div></br>", 1)

	//Replace 'Show Source'
	if m := regExpShowSourceBegin.FindAllStringSubmatch(html, -1); m != nil {
		for _, match := range m {
			var buffer bytes.Buffer
			buffer.WriteString(`<p><div><a id="source-button" class="source-button" onclick="toggle_source('`)
			buffer.WriteString(match[2])
			buffer.WriteString(`')"><span id="btn-text-`)
			buffer.WriteString(match[2])
			buffer.WriteString(`">&#x25BC;</span>&nbsp;Show source</a>`)
			buffer.WriteString(`<div id="showsource-`)
			buffer.WriteString(match[2])
			buffer.WriteString(`" style="display:none;">`)
			html = strings.Replace(html, match[1], buffer.String(), 1)
		}
	}
	html = strings.Replace(html, PlaceHolderShowSourceEnd, "</div></br>", -1)
	//`<div class="highlight highlight-source-swift line-numbers"><codeblocks class="highlight-source-swift">`
	codeBlockBackColor := fmt.Sprintf("<div class=\"highlight highlight-source-swift line-numbers\" style=\"background:%s\"><codeblocks class=\"highlight-source-swift\">", BuiltinCssStyle[Cfg.CssStyle][1])
	html = strings.Replace(html, `<div class="highlight highlight-source-swift"><pre>`, codeBlockBackColor, -1)

	if m := regLinePlaceHolder.FindAllStringSubmatch(html, -1); m != nil {
		for _, match := range m {
			var buffer bytes.Buffer
			buffer.WriteString(genLineNumberRows(match[1]))
			buffer.WriteString("</codeblocks></pre></div>\n")
			html = strings.Replace(html, match[0], buffer.String(), 1)
		}
	}

	return html
}

// ----------------------------------------------------------------------------
// Sorting

type data struct {
	n    int
	swap func(i, j int)
	less func(i, j int) bool
}

func (d *data) Len() int           { return d.n }
func (d *data) Swap(i, j int)      { d.swap(i, j) }
func (d *data) Less(i, j int) bool { return d.less(i, j) }

// sortBy is a helper function for sorting
func sortBy(less func(i, j int) bool, swap func(i, j int), n int) {
	sort.Sort(&data{n, swap, less})
}

func sortedClasses(classes []*ast.ClassStatement, fh *os.File) []*Classes {
	list := make([]*Classes, len(classes))
	i := 0
	for _, c := range classes {

		funcs := make([]*ast.FunctionStatement, 0)
		for _, fn := range c.ClassLiteral.Methods {
			if fn.Doc != nil {
				funcs = append(funcs, fn)
			}
		}

		props := make([]*ast.PropertyDeclStmt, 0)
		for _, prop := range c.ClassLiteral.Properties {
			if prop.Doc != nil {
				props = append(props, prop)
			}
		}

		lets := make([]*ast.LetStatement, 0)
		for _, member := range c.ClassLiteral.Members {
			if member.Doc != nil {
				lets = append(lets, member)
			}
		}

		src, lineSrc := genSourceText(c, fh)
		list[i] = &Classes{
			Value: &Value{
				Name:     c.Name.Value,
				Doc:      preProcessCommentSpecial(c.Doc.Text()),
				Text:     c.Docs(),
				ShowSrc:  Cfg.ShowSrcComment,
				Src:      src,
				SrcLines: lineSrc,
				GenHTML:  Cfg.GenHTML,
			},
			Props: sortedProps(props, fh),
			Lets:  sortedLets(lets, fh),
			Funcs: sortedFuncs(funcs, fh),
		}
		i++
	}

	sortBy(
		func(i, j int) bool { return list[i].Value.Name < list[j].Value.Name },
		func(i, j int) {
			list[i].Value, list[j].Value = list[j].Value, list[i].Value
			list[i].Props, list[j].Props = list[j].Props, list[i].Props
			list[i].Lets, list[j].Lets = list[j].Lets, list[i].Lets
			list[i].Funcs, list[j].Funcs = list[j].Funcs, list[i].Funcs
		},
		len(list),
	)
	return list
}

func sortedLets(lets []*ast.LetStatement, fh *os.File) []*Value {
	list := make([]*Value, len(lets))
	i := 0
	for _, l := range lets {
		src, lineSrc := genSourceText(l, fh)
		list[i] = &Value{
			Name:     l.Names[0].Value,
			Doc:      preProcessCommentSpecial(l.Doc.Text()),
			Text:     l.Docs(),
			ShowSrc:  Cfg.ShowSrcComment,
			Src:      src,
			SrcLines: lineSrc,
			GenHTML:  Cfg.GenHTML,
		}
		i++
	}

	sortBy(
		func(i, j int) bool { return list[i].Name < list[j].Name },
		func(i, j int) { list[i], list[j] = list[j], list[i] },
		len(list),
	)
	return list
}

func sortedConsts(consts []*ast.ConstStatement, fh *os.File) []*Value {
	list := make([]*Value, len(consts))
	i := 0
	for idx, c := range consts {
		src, lineSrc := genSourceText(c, fh)
		list[i] = &Value{
			Name:     c.Name[idx].Value,
			Doc:      preProcessCommentSpecial(c.Doc.Text()),
			Text:     c.Docs(),
			ShowSrc:  Cfg.ShowSrcComment,
			Src:      src,
			SrcLines: lineSrc,
			GenHTML:  Cfg.GenHTML,
		}
		i++
	}

	sortBy(
		func(i, j int) bool { return list[i].Name < list[j].Name },
		func(i, j int) { list[i], list[j] = list[j], list[i] },
		len(list),
	)
	return list
}

func sortedEnums(enums []*ast.EnumStatement, fh *os.File) []*Value {
	list := make([]*Value, len(enums))
	i := 0
	for _, e := range enums {
		src, lineSrc := genSourceText(e, fh)
		list[i] = &Value{
			Name:     e.Name.Value,
			Doc:      preProcessCommentSpecial(e.Doc.Text()),
			Text:     e.Docs(),
			ShowSrc:  Cfg.ShowSrcComment,
			Src:      src,
			SrcLines: lineSrc,
			GenHTML:  Cfg.GenHTML,
		}
		i++
	}

	sortBy(
		func(i, j int) bool { return list[i].Name < list[j].Name },
		func(i, j int) { list[i], list[j] = list[j], list[i] },
		len(list),
	)
	return list
}

func sortedFuncs(funcs []*ast.FunctionStatement, fh *os.File) []*Function {
	list := make([]*Function, len(funcs))
	i := 0
	for _, f := range funcs {
		list[i] = parseFuncComment(f.Name.Value, preProcessCommentSpecial(f.Doc.Text()), f.Docs())
		list[i].Value.Src, list[i].Value.SrcLines = genSourceText(f, fh)
		list[i].Value.ShowSrc = Cfg.ShowSrcComment
		list[i].Value.GenHTML = Cfg.GenHTML
		i++
	}

	sortBy(
		func(i, j int) bool { return list[i].Value.Name < list[j].Value.Name },
		func(i, j int) {
			list[i].Value, list[j].Value = list[j].Value, list[i].Value
			list[i].Params, list[j].Params = list[j].Params, list[i].Params
			list[i].Returns, list[j].Returns = list[j].Returns, list[i].Returns
		},
		len(list),
	)
	return list
}

func sortedProps(props []*ast.PropertyDeclStmt, fh *os.File) []*Value {
	list := make([]*Value, len(props))
	i := 0
	for _, p := range props {
		src, lineSrc := genSourceText(p, fh)
		list[i] = &Value{
			Name:     p.Name.Value,
			Doc:      preProcessCommentSpecial(p.Doc.Text()),
			Text:     p.Docs(),
			ShowSrc:  Cfg.ShowSrcComment,
			Src:      src,
			SrcLines: lineSrc,
			GenHTML:  Cfg.GenHTML,
		}

		if strings.HasPrefix(p.Name.Value, "this") {
			list[i].Name = "this"
		} else {
			list[i].Name = p.Name.Value
		}
		i++
	}

	sortBy(
		func(i, j int) bool { return list[i].Name < list[j].Name },
		func(i, j int) { list[i], list[j] = list[j], list[i] },
		len(list),
	)
	return list
}

func parseFuncComment(name string, docComments string, text string) *Function {
	fn := &Function{
		Value: &Value{
			Name: name,
			Text: text,
		},
		Params:  make([]*FuncInfo, 0),
		Returns: make([]*FuncInfo, 0),
	}

	var buffer bytes.Buffer
	comments := strings.Split(docComments, "\n")
	for _, comment := range comments {
		if len(comment) > 0 && comment[0] == '@' {
			splitOnSpaces := strings.Split(comment, " ")
			label := splitOnSpaces[0]
			switch label {
			case "@param":
				funcParam := parseValue(splitOnSpaces[1:])
				fn.Params = append(fn.Params, funcParam)
			case "@return", "@returns":
				funcReturn := parseValue(splitOnSpaces[1:])
				fn.Returns = append(fn.Returns, funcReturn)
			}
		} else {
			buffer.WriteString(comment + "\n")
		}
	}
	fn.Value.Doc = buffer.String()

	return fn
}

func parseValue(splitOnSpaces []string) *FuncInfo {
	name := ""
	types := ""
	var description bytes.Buffer

	description.WriteString("")
	ret := &FuncInfo{Name: "", Type: "", Desc: ""}
	for _, item := range splitOnSpaces {
		if m := regexpType.FindStringSubmatch(item); m != nil {
			types = m[1]
		} else if len(name) == 0 {
			if len(item) > 0 && item[0] == '`' {
				name = item[1 : len(item)-1]
			} else {
				name = item
			}
		} else {
			description.WriteString(item + " ")
		}
	}

	if len(name) > 0 {
		ret.Name = name
	}
	if len(types) > 0 {
		ret.Type = types
	}
	ret.Desc = description.String()

	return ret
}

// SanitizedAnchorName returns a sanitized anchor name for the given text.
//copied from 'Blackfriday': a markdown processor for Go.
func SanitizedAnchorName(text string) string {
	var anchorName []rune
	futureDash := false
	for _, r := range text {
		switch {
		case unicode.IsLetter(r) || unicode.IsNumber(r):
			if futureDash && len(anchorName) > 0 {
				anchorName = append(anchorName, '-')
			}
			futureDash = false
			anchorName = append(anchorName, unicode.ToLower(r))
		default:
			futureDash = true
		}
	}
	return string(anchorName)
}

/* Process `@example block @` part, and replace this with
```swift
    block
```
*/
func preProcessCommentSpecial(comments string) string {
	retStr := comments
	//@example
	if m := regExample.FindAllStringSubmatch(comments, -1); m != nil {
		for _, match := range m {
			var buffer bytes.Buffer
			buffer.WriteString("\n```swift\n")
			buffer.WriteString(match[1])
			buffer.WriteString("\n```\n")

			//Add line numbers
			if Cfg.GenHTML == 1 {
				lines := strings.Split(match[1], "\n")
				if len(lines) > 1 {
					buffer.WriteString(fmt.Sprintf("<p>__LINENUMBER_PLACEHOLDER_LINE__%d</p>\n", len(lines)-1))
				}
			}

			retStr = replaceFirstString(regExample, retStr, buffer.String())
		}
		//fmt.Printf("retStr=<%s>\n", retStr) //debugging
	}

	//@note
	if m := regNote.FindAllStringSubmatch(comments, -1); m != nil {
		for _, match := range m {
			var buffer bytes.Buffer
			if Cfg.GenHTML == 0 {
				buffer.WriteString("#### Note\n")
				buffer.WriteString(match[1])
			} else {
				buffer.WriteString(`<div id="user-content-note">&nbsp;:bulb: Note<p>`)
				buffer.WriteString(match[1])
				buffer.WriteString("</p></div>")
			}
			retStr = replaceFirstString(regNote, retStr, buffer.String())
		}
		//fmt.Printf("retStr=<%s>\n", retStr) //debugging
	}

	//@warn
	if m := regWarning.FindAllStringSubmatch(comments, -1); m != nil {
		for _, match := range m {
			var buffer bytes.Buffer
			if Cfg.GenHTML == 0 {
				buffer.WriteString("#### Warning\n")
				buffer.WriteString(match[1])
			} else {
				buffer.WriteString(`<div id="user-content-warning">&nbsp;:warning: Warning<p>`)
				buffer.WriteString(match[1])
				buffer.WriteString("</p></div>")
			}
			retStr = replaceFirstString(regWarning, retStr, buffer.String())
		}
		//fmt.Printf("retStr=<%s>\n", retStr) //debugging
	}

	return retStr
}

func replaceFirstString(re *regexp.Regexp, srcStr, replStr string) string {
	src := []byte(srcStr)
	repl := []byte(replStr)

	if m := re.FindSubmatchIndex(src); m != nil {
		out := make([]byte, m[0])
		copy(out, src[0:m[0]])
		out = re.Expand(out, repl, src, m)
		if m[1] < len(src) {
			out = append(out, src[m[1]:]...)
		}
		return string(out)
	}
	out := make([]byte, len(src))
	copy(out, src)
	return string(out)
}

//Generate source text from AST
func genSourceText(src ast.Source, fh *os.File) (string, int) {
	tStart := src.SrcStart()
	tEnd := src.SrcEnd()

	buf := make([]byte, tEnd.Offset-tStart.Offset)
	fh.ReadAt(buf, int64(tStart.Offset))

	return string(buf), tEnd.Line - tStart.Line + 1
}

func genLineNumberRows(lineCnt string) string {
	iLineCnt, _ := strconv.Atoi(lineCnt)

	var buffer bytes.Buffer
	buffer.WriteString(`<span class="line-numbers-rows" style="left: -60px;">`)
	for i := 0; i < iLineCnt; i++ {
		buffer.WriteString(`<span></span>`)
	}

	return buffer.String()
}

//convert '[]string' to '[]interface{}'
func strArr2IntfArr(strArr []string) []interface{} {
	var intfArr []interface{}
	for _, str := range strArr {
		intfArr = append(intfArr, str)
	}
	return intfArr
}
