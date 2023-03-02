// md previews a given markdown file.
//
// It starts an HTTP server that renders the specified markdown file every time
// "/" is requested.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
)

func main() {
	httpAddr := flag.String("http", "127.0.0.1:0", "HTTP service address")
	openURL := flag.Bool("open", false, "open the URL in the default web browser")

	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 1 {
		usage()
		os.Exit(2)
	}
	filename := flag.Arg(0)

	l, err := net.Listen("tcp", *httpAddr)
	if err != nil {
		log.Fatalf("error: listen: %v", err)
	}

	url := fmt.Sprintf("http://%v", l.Addr())

	log.Printf("Open your web browser and visit %v", url)

	if *openURL {
		if err := browse(url); err != nil {
			log.Fatalf("error: browse: %v", err)
		}
	}

	http.Handle("/", mdHandler(filename))

	if err := http.Serve(l, nil); err != nil {
		log.Fatalf("error: serve: %v", err)
	}
}

func mdHandler(filename string) http.HandlerFunc {
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
	)

	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%v %v %v", r.Method, r.URL.Path, r.RemoteAddr)

		source, err := os.ReadFile(filename)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "error: read file: %v\n", err)
			return
		}

		var buf bytes.Buffer
		if err := md.Convert(source, &buf); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "error: convert md: %v\n", err)
			return
		}

		fmt.Fprintln(w, "<html>")
		fmt.Fprintln(w, "<head><title>md</title></head>")
		fmt.Fprintf(w, "<body>\n%v\n</body>\n", buf.String())
		fmt.Fprintln(w, "</html>")
	}
}

func usage() {
	fmt.Printf("usage: %v [flags] file.md\n", os.Args[0])
	flag.PrintDefaults()
}
