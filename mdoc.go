/* Documentor generator for magpie. */
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"magpie/docs"
	"magpie/lexer"
	"magpie/parser"
	"path/filepath"
	"strings"
	"os"
)

func genDocs(path string, cfg doc.Config, isDir bool) {
	if !isDir { //single file
		genDoc(path, cfg)
		return
	}

	//processing directory
	fd, err := os.Open(path)
	if err != nil {
		fmt.Errorf("[!] ","Open directory '%s' failed, reason:%v\n", path, err)
		return
	}
	defer fd.Close()

	list, err := fd.Readdir(-1)
	if err != nil {
		fmt.Errorf("[!] ","Read directory '%s' failed, reason:%v\n", path, err)
		return
	}

	for _, d := range list {
		if strings.HasSuffix(d.Name(), ".mp") {
			filename := filepath.Join(path, d.Name())
			genDoc(filename, cfg)
		}
	}
}

func genDoc(filename string, cfg doc.Config) {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	f, err := ioutil.ReadFile(wd + "/" + filename)
	if err != nil {
		fmt.Println("[!] ","AeroScript: ", err.Error())
		os.Exit(1)
	}

	contents := string(f)
	l := lexer.New(filename, contents)
	parser.FileLines = strings.Split(contents, "\n")
	p := parser.NewWithDoc(l, wd)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		for _, err := range p.Errors() {
			fmt.Println(err)
		}
		os.Exit(1)
	}

	//generate markdown docs
	file := doc.New(filename, program)
	md := doc.MdDocGen(file)

	//remove the '.mp' extension
	genFilename := strings.TrimSuffix(filename, filepath.Ext(filename))

	//create markdown file
	mdFile := genFilename + ".md"
	outMd, err := os.Create(mdFile)
	if err != nil {
		fmt.Printf("[!] ","Error creating '%s' file, reason:%v\n", mdFile, err)
		os.Exit(1)
	}

	//generate markdown file
	fmt.Fprintln(outMd, md)
	outMd.Close()

	if cfg.GenHTML == 0 {
		return
	}

	//create html file
	htmlFile := genFilename + ".html"
	outHtml, err := os.Create(htmlFile)
	if err != nil {
		fmt.Printf("[!] ","Error creating '%s' file, reason:%v\n", htmlFile, err)
		os.Exit(1)
	}
	defer outHtml.Close()

	html := doc.HtmlDocGen(md, file)
	fmt.Fprintln(outHtml, html)

	err = os.Remove(mdFile)
	if err != nil {
		fmt.Printf("[!] ","Error remove file '%s', reason : %v\n", mdFile, err)
	}
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [magpie file]\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(0)
	}

	var htmlFlag bool
	flag.BoolVar(&htmlFlag, "html", false, "Generate html file using github REST API.")

	var showSrcFlag bool
	flag.BoolVar(&showSrcFlag, "showsource", false, "Show class and function source code in Generated file.")

	var cssStyle int
	msg := fmt.Sprintf("[!] ","Set css style(Avialable: 0-%d) to use for html output.", len(doc.BuiltinCssStyle))
	flag.IntVar(&cssStyle, "css", 0, msg)

	var cssFile string
	flag.StringVar(&cssFile, "cssfile", "", "Css file to use for generating html file.")

	//parse the command line options
	flag.Parse()

	if len(flag.Args()) != 1 {
		fmt.Fprintln(os.Stderr, "Invalid number of arguments!")
		flag.Usage()
	}

	path := flag.Arg(0)
	fi, err := os.Stat(path)
	if err != nil {
		fmt.Println(err)
		return
	}

	if htmlFlag {
		doc.Cfg.GenHTML = 1

		if cssFile != "" {
			cssContents, err := ioutil.ReadFile(cssFile)
			if err != nil {
				fmt.Println("[!] ","Error reading css file: ", err.Error())
				//do not exits
			} else {
				doc.Cfg.CssContents = string(cssContents)
			}
		}

		if cssStyle > len(doc.BuiltinCssStyle)-1 || cssStyle < 0 {
			cssStyle = 0 //default
		}
		doc.Cfg.CssStyle = cssStyle

	}

	if showSrcFlag {
		doc.Cfg.ShowSrcComment = 1
	}
	switch mode := fi.Mode(); {
	case mode.IsDir():
		genDocs(path, doc.Cfg, true)
	case mode.IsRegular():
		genDocs(path, doc.Cfg, false)
	}
	
}
