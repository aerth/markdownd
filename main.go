/*
* The MIT License (MIT)
*
* Copyright (c) 2017  aerth <aerth@riseup.net>
*
* Permission is hereby granted, free of charge, to any person obtaining a copy
* of this software and associated documentation files (the "Software"), to deal
* in the Software without restriction, including without limitation the rights
* to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
* copies of the Software, and to permit persons to whom the Software is
* furnished to do so, subject to the following conditions:
*
* The above copyright notice and this permission notice shall be included in all
* copies or substantial portions of the Software.
*
* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
* IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
* FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
* AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
* LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
* OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
* SOFTWARE.
 */

// The markdownd command serves markdown, static, and html files.
package main

import (
	"flag"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/russross/blackfriday"
	"github.com/shurcool/github_flavored_markdown"
	"github.com/sourcegraph/syntaxhighlight"
)

// flags
var (
	addr          = flag.String("http", ":8080", "address to listen on format 'address:port',\n\tif address is omitted will listen on all interfaces")
	logfile       = flag.String("log", os.Stderr.Name(), "redirect logs to this file")
	indexPage     = flag.String("index", "index.md", "filename to use for paths ending in '/',\n\ttry something like '-index=README.md'")
	header        = flag.String("header", "", "html header filename for markdown requests")
	footer        = flag.String("footer", "", "html footer filename for markdown requests")
	toc           = flag.Bool("toc", false, "generate table of contents at the top of each markdown page")
	plain         = flag.Bool("plain", false, "disable github flavored markdown")
	syntaxEnabled = flag.Bool("syntax", false, "highlight syntax in .html")
)

// log to file
var logger = log.New(os.Stderr, "", 0)

const version = "0.0.10"
const sig = "[markdownd v" + version + "]\nhttps://github.com/aerth/markdownd"
const serverheader = "markdownd/" + version
const usage = `
USAGE

markdownd [flags] [directory]

EXAMPLES

Serve current directory on port 8080, log to stderr:
	markdownd -log /dev/stderr -http 127.0.0.1:8080 .

Serve 'docs' directory on port 8081, log to 'md.log':
	markdownd -log md.log -http :8081 docs

Serve docs with header, footer, and table of contents. Disable Logs:
	markdownd -log none -header bar.html -footer foo.html -toc docs

Serve docs only on localhost:
	markdownd -http 127.0.0.1:8080 docs
`

// redefine flag Usage
func init() {
	flag.Usage = func() {
		println(usage)
		println("FLAGS")
		flag.PrintDefaults()
	}
	rand.Seed(time.Now().UnixNano())
}

// Handler handles markdown requests
type Handler struct {
	Root           http.FileSystem // directory to serve
	RootString     string          // keep directory name for comparing prefix
	header, footer []byte          // for not-raw markdown requests
}

// markdown command
func main() {
	println(sig)
	flag.Parse()
	serve(flag.Args())
}

func serve(args []string) {
	// need only 1 argument, the directory to serve
	if len(args) != 1 {
		flag.Usage()
		os.Exit(111)
		return
	}

	// get absolute path of flag.Arg(0)
	dir := flag.Arg(0)
	dir = prepareDirectory(dir)

	_, err := os.Stat(dir + *indexPage)
	if err != nil {
		logger.Printf("warning: %q not found, did you forget '-index' flag?", *indexPage)
	}

	// new markdown handler
	mdhandler := &Handler{
		Root:       http.Dir(dir),
		RootString: dir,
	}

	h := http.DefaultServeMux
	h.Handle("/", mdhandler)
	// print absolute directory we are serving
	println("serving filesystem:", dir)

	// take care of opening log file
	openLogFile()
	println("logging to:", *logfile)

	if *header != "" {
		println("html header:", *header)
		b, err := ioutil.ReadFile(*header)
		if err != nil {
			println(err.Error())
			os.Exit(111)
		}
		mdhandler.header = b
	} else {
		mdhandler.header = []byte("<!DOCTYPE html>\n")
	}

	if *footer != "" {
		println("html footer:", *footer)
		b, err := ioutil.ReadFile(*footer)
		if err != nil {
			println(err.Error())
			os.Exit(111)
		}
		mdhandler.footer = b
	}

	// create a http server
	server := &http.Server{
		Addr:              *addr,
		Handler:           h,
		ErrorLog:          logger,
		MaxHeaderBytes:    (1 << 10), // 1KB
		ReadTimeout:       (time.Second * 5),
		WriteTimeout:      (time.Second * 5),
		ReadHeaderTimeout: (time.Second * 5),
		IdleTimeout:       (time.Second * 5),
	}

	// disable keepalives
	server.SetKeepAlivesEnabled(false)

	// trick to show listening port
	go func() { <-time.After(time.Second); println("listening:", *addr) }()

	// add date+time to log entries
	logger.SetFlags(log.LstdFlags)

	// start serving
	err = server.ListenAndServe()

	// print usage info, probably started wrong or port is occupied
	flag.Usage()

	// always non-nil
	logger.Println(err)

	// any exit is an error
	os.Exit(111)
}

// generate kind-of-unique string
func rfid() string {
	return strconv.Itoa(rand.Int())
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// all we want is GET
	if r.Method != "GET" {
		logger.Println("bad method:", r.RemoteAddr, r.Method, r.URL.Path, r.UserAgent())
		http.NotFound(w, r)
		return
	}

	// deny requests containing '../'
	if strings.Contains(r.URL.Path, "../") {
		logger.Println("bad path:", r.RemoteAddr, r.Method, r.URL.Path, r.UserAgent())
		http.NotFound(w, r)
		return
	}

	// start timing
	t1 := time.Now()

	// Add Server header
	w.Header().Add("Server", serverheader)

	// Prevent page from being displayed in an iframe
	w.Header().Add("X-Frame-Options", "DENY")

	// generate unique request id
	requestid := rfid()

	// log how long this takes
	defer func(t func() time.Time) {
		logger.Println(requestid, "closed after", t().Sub(t1))
	}(time.Now)

	if *syntaxEnabled && r.URL.Path == "/gh.css" {
		b, err := Asset("static/gh.css")
		if err == nil {
			w.Header().Add("Content-Type", "test/css")
			w.Write(b)
			return
		}
	}

	// abs is not absolute yet
	abs := r.URL.Path[1:] // remove slash
	if abs == "" {
		abs = *indexPage
	}

	// '/' suffix, add *index.Page
	if strings.HasSuffix(abs, "/") {
		abs += *indexPage
	}

	// still not absolute, prepend root directory to filesrc
	abs = h.RootString + abs

	// log now that we have filename
	logger.Println(requestid, r.RemoteAddr, r.Method, r.URL.Path, "->", abs)

	// get absolute path of requested file (could not exist)
	abs, err := filepath.Abs(abs)
	if err != nil {
		logger.Println(requestid, "error resolving absolute path:", err)
		http.NotFound(w, r)
		return
	}

	// .html suffix, but .md exists. choose to serve .md over .html
	if strings.HasSuffix(abs, ".html") {
		trymd := strings.TrimSuffix(abs, ".html") + ".md"
		_, err := os.Open(trymd)
		if err == nil {
			logger.Println(requestid, abs, "->", trymd)
			abs = trymd
		}
	}

	// check if exists, or give 404
	_, err = os.Open(abs)
	if err != nil {
		if strings.Contains(err.Error(), "no such file") {
			logger.Println(requestid, "404", abs)
			http.NotFound(w, r)
			return
		}

		// probably permissions
		logger.Println(requestid, "error opening file:", err, abs)
		http.NotFound(w, r)
		return
	}

	// check if symlink ( to avoid /proc/self/root style attacks )
	if !fileisgood(abs) {
		logger.Printf("%s error: %q is symlink. serving 404", requestid, abs)
		http.NotFound(w, r)
		return
	}

	// compare prefix (alternate way of checking symlink?)
	// above, we checked for abs vs symlink resolved,
	// here lets check if they have the special prefix of "s.Root"
	// probably redundant.
	if !strings.HasPrefix(abs, h.RootString) {
		logger.Println(requestid, "bad path", abs, "doesnt have prefix:", h.RootString)
		http.NotFound(w, r)
		return
	}

	// read bytes (for detecting content type )
	b, err := ioutil.ReadFile(abs)
	if err != nil {
		logger.Printf("%s error reading file: %q", requestid, abs)
		http.NotFound(w, r)
		return
	}

	// detect content type and encoding
	ct := http.DetectContentType(b)

	// serve raw html if exists
	if strings.HasSuffix(abs, ".html") && strings.HasPrefix(ct, "text/html") {
		logger.Println(requestid, "serving raw html:", abs)
		w.Header().Add("Content-Type", "text/html")
		w.Write(b)
		return
	}

	// probably markdown
	if strings.HasSuffix(abs, ".md") && strings.HasPrefix(ct, "text/plain") {
		if strings.Contains(r.URL.RawQuery, "raw") {
			logger.Println(requestid, "raw markdown request:", abs)
			w.Write(b)
			return
		}
		logger.Println(requestid, "serving markdown:", abs)

		md := markdown2html(b)
		if md == nil {
			w.WriteHeader(200)
			return
		}
		w.Header().Add("Content-Type", "text/html")
		w.Write(h.header)
		w.Write(md)
		w.Write(h.footer)
		return
	}

	// fallthrough with http.ServeFile
	logger.Printf("%s serving %s: %s", requestid, ct, abs)

	http.ServeFile(w, r, abs)
}

// fileisgood returns false if symlink
// comparing absolute vs resolved path is apparently quick and effective
func fileisgood(abs string) bool {

	// sanity check
	if abs == "" {
		return false
	}

	// is absolute really absolute?
	var err error
	if !filepath.IsAbs(abs) {
		abs, err = filepath.Abs(abs)
	}
	if err != nil {
		println(err.Error())
		return false
	}

	// get real path after eval symlinks
	realpath, err := filepath.EvalSymlinks(abs)
	if err != nil {
		println(err.Error())
		return false
	}

	// equality check
	return realpath == abs
}

// prepare root filesystem directory for serving
func prepareDirectory(dir string) string {
	// add slash to dot
	if dir == "." {
		dir += string(os.PathSeparator)
	}

	// become absolute
	var err error
	dir, err = filepath.Abs(dir)
	if err != nil {
		println(err.Error())
		os.Exit(111)
		return err.Error()
	}

	// add trailing slash (for comparing prefix)
	if !strings.HasSuffix(dir, string(os.PathSeparator)) {
		dir += string(os.PathSeparator)
	}

	return dir
}

func markdown2html(in []byte) []byte {
	if len(in) == 0 {
		return nil
	}
	if *plain {
		// default flags
		flags := 0
		if *toc {
			flags |= blackfriday.HTML_TOC
		}
		md := blackfriday.Markdown(
			in, blackfriday.HtmlRenderer(
				// html flags
				flags,
				"", ""),
			// extensions
			0)
		return md
	}

	return github_flavored_markdown.Markdown(in)
}

// use logfile flag and set logger Logger
func openLogFile() {
	switch *logfile {
	case os.Stderr.Name(), "stderr":
		// already stderr
		*logfile = os.Stderr.Name()
	case os.Stdout.Name(), "stdout":
		logger.SetOutput(os.Stdout)
		*logfile = os.Stdout.Name()
	case "none", "no", "null", "/dev/null", "nil", "disabled":
		logger.SetOutput(ioutil.Discard)
		*logfile = os.DevNull
	default:
		func() {
			logger.Printf("Opening log file: %q", *logfile)
			f, err := os.OpenFile(*logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0660)
			if err != nil {
				logger.Fatalf("cant open log file: %s", err)
			}
			logger.SetOutput(f)
		}()
	}

}

func highlightSyntaxHTML(in []byte) (out []byte) {
	out, err := syntaxhighlight.AsHTML(in)
	if err != nil {
		logger.Println("error highlighting syntax:", err)
		return in
	}
	return out
}
