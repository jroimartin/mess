// md helps to preview markdown files.
package main

import (
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"text/template"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
)

//go:embed main.tmpl
var mainTemplateHTML string

var (
	mainTemplate = template.Must(template.New("main").Parse(mainTemplateHTML))

	md = goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
	)
)

func main() {
	httpAddr := flag.String("http", "127.0.0.1:0", "HTTP service address")
	openURL := flag.Bool("open", false, "open the URL in the default web browser")

	flag.Usage = usage
	flag.Parse()

	var path string
	switch flag.NArg() {
	case 0:
		path = "."
	case 1:
		path = flag.Arg(0)
	default:
		usage()
		os.Exit(2)
	}

	var err error
	if path == "-" {
		err = render(path)
	} else {
		err = serve(path, *httpAddr, *openURL)
	}

	if err != nil {
		log.Fatalf("error: %v", err)
	}
}

func render(path string) error {
	source, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("read all: %w", err)
	}

	var buf bytes.Buffer
	if err := md.Convert(source, &buf); err != nil {
		return fmt.Errorf("convert md: %w", err)
	}

	mainTemplate.Execute(os.Stdout, buf.String())

	return nil
}

func serve(path, httpAddr string, openURL bool) error {
	dir, file, err := splitPath(path)
	if err != nil {
		return fmt.Errorf("split path: %w", err)
	}

	log.Printf("Working directory: %v", dir)

	if file != "" {
		log.Printf("File: %v", file)
	}

	l, err := net.Listen("tcp", httpAddr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	fileURL, err := url.Parse("http://" + l.Addr().String())
	if err != nil {
		return fmt.Errorf("parse url: %w", err)
	}
	fileURL = fileURL.JoinPath(file)

	log.Printf("Open your web browser and visit %v", fileURL)

	if openURL {
		if err := browse(fileURL.String()); err != nil {
			return fmt.Errorf("browse: %w", err)
		}
	}

	mh := MarkdownHandler{
		h:   http.FileServer(http.Dir(dir)),
		dir: dir,
	}

	http.Handle("/", mh)

	if err := http.Serve(l, nil); err != nil {
		return fmt.Errorf("serve: %w", err)
	}

	return nil
}

type MarkdownHandler struct {
	h   http.Handler
	dir string
}

func (mh MarkdownHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("%v %v %v", r.RemoteAddr, r.Method, r.URL.Path)

	if filepath.Ext(r.URL.Path) != ".md" {
		mh.h.ServeHTTP(w, r)
		return
	}

	source, err := os.ReadFile(filepath.Join(mh.dir, r.URL.Path))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error: read file: %v\n", err)
		return
	}

	var buf bytes.Buffer
	if err := md.Convert(source, &buf); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "convert md: %v\n", err)
		return
	}

	mainTemplate.Execute(w, buf.String())
}

func splitPath(path string) (dir, file string, err error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return "", "", fmt.Errorf("stat: %v", err)
	}

	if fileInfo.IsDir() {
		return path, "", nil
	} else {
		dir, file = filepath.Split(path)
		if dir == "" {
			dir = "."
		}
		return dir, file, nil
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %v [flags] [path]\n", os.Args[0])
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nIf path is - stdin and stdout are used.\n")
}
