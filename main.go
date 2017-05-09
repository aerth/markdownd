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

// Command markdownd serves markdown, static, and html files.
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

	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
)

var (
	addr    = flag.String("http", ":8080", "address to listen on")
	logfile = flag.String("log", os.Stderr.Name(), "redirect logs to this file")
)

const version = "0.0.4"
const sig = "[markdownd v" + version + "]\nhttps://github.com/aerth/markdownd"
const serverheader = "markdownd/" + version

func init() {
	flag.Usage = func() {
		println(usage)
		println("FLAGS")
		flag.PrintDefaults()
	}
}

const usage = `
USAGE

markdownd [flags] [directory]

EXAMPLES

Serve current directory on port 8080, log to stderr
	markdownd -log /dev/stderr -http 127.0.0.1:8080 .

Serve 'docs' directory on port 8081, log to 'md.log'
	markdownd -log md.log -http :8081`

func init() {
	println(sig)
}

func main() {

	flag.Parse()
	if len(flag.Args()) != 1 {
		println(usage)
		os.Exit(111)
		return
	}

	// get absolute path of flag.Arg(0)
	dir := flag.Arg(0)
	if dir == "." {
		dir = "./"
	}
	var err error
	dir, err = filepath.Abs(dir)

	if err != nil {
		println(err.Error())
		os.Exit(111)
		return
	}

	if !strings.HasSuffix(dir, "/") {
		dir += "/"
	}

	srv := &Server{
		Root:       http.Dir(dir),
		RootString: dir,
	}
	println("http filesystem:", dir)

	if *logfile != os.Stderr.Name() {
		f, err := os.OpenFile(*logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0660)
		if err != nil {
			log.Fatalf("cant open log file: %s", err)
		}
		log.SetOutput(f)
	}
	println("log output:", *logfile)

	log.Println(http.ListenAndServe(*addr, srv).Error())
	return
}

type Server struct {
	Root       http.FileSystem
	RootString string
}

func rfid() string {
	return strconv.Itoa(rand.Int())
}

func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" || strings.Contains(r.URL.Path, "../") {
		log.Println("bad request:", r.RemoteAddr, r.Method, r.URL.Path, r.UserAgent())
		http.NotFound(w, r)
		return
	}

	basedir := filepath.Base(s.RootString)

	w.Header().Add("Server", serverheader)

	// generate unique request id
	requestid := rfid()

	filesrc := r.URL.Path[1:] // remove slash

	if filesrc == "" {
		filesrc = "index.md"
	}

	filesrc = s.RootString + filesrc

	log.Println(requestid, r.RemoteAddr, r.Method, r.URL.Path, "->", filesrc)

	if strings.HasSuffix(filesrc, "/") {
		log.Printf("%s %s -> %sindex.md", requestid, filesrc, filesrc)
		filesrc += "index.md"
	}

	if strings.HasSuffix(filesrc, ".html") {
		trymd := strings.TrimSuffix(filesrc, ".html") + ".md"
		_, err := os.Open(trymd)
		if err == nil {
			log.Println(requestid, filesrc, "->", trymd)
			filesrc = trymd
		}
	}

	defer log.Println(requestid, "closed")

	// get absolute path of requested file (could not exist)
	abs, err := filepath.Abs(filesrc)
	if err != nil {
		log.Println(requestid, "error resolving absolute path:", err)
		http.NotFound(w, r)
		return
	}

	// check if exists, or give 404
	_, err = os.Open(abs)
	if err != nil {
		if strings.Contains(err.Error(), "no such file") {
			log.Println(requestid, "404", abs)
			http.NotFound(w, r)
			return
		}
		log.Println(requestid, "error opening file:", err, abs)
		http.NotFound(w, r)
		return
	}

	// check if symlink ( to avoid /proc/self/root style attacks )
	if !fileisgood(abs) {
		log.Printf("%s error: %q is symlink. serving 404", requestid, abs)
		http.NotFound(w, r)
		return
	}

	// read bytes
	b, err := ioutil.ReadFile(abs)
	if err != nil {
		log.Printf("%s error reading file: %q", requestid, filesrc)
		http.NotFound(w, r)
		return
	}

	// detect content type and encoding
	ct := http.DetectContentType(b)

	// serve raw html if exists
	if strings.HasSuffix(abs, ".html") && strings.HasPrefix(ct, "text/html") {
		log.Println(requestid, "serving raw html:", abs)
		w.Write(b)
		return
	}

	// probably markdown
	if strings.HasSuffix(abs, ".md") && strings.HasPrefix(ct, "text/plain") {
		if strings.Contains(r.URL.RawQuery, "raw") {
			log.Println(requestid, "raw markdown request:", abs)
			w.Write(b)
			return
		}
		log.Println(requestid, "serving markdown:", abs)
		unsafe := blackfriday.MarkdownCommon(b)
		html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)
		w.Write(html)
		return
	}

	// fallthrough with http.ServeFile
	log.Printf("%s serving %s: %s", requestid, ct, abs)
	if !strings.HasPrefix(filepath.Base(abs), basedir) {
		log.Println(requestid, "bad path", abs)
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, abs)
}

// fileisgood returns false if symlink
// comparing absolute vs resolved path is apparently quick and effective
func fileisgood(abs string) bool {
	if abs == "" {
		return false
	}

	var err error
	if !filepath.IsAbs(abs) {
		abs, err = filepath.Abs(abs)
	}

	if err != nil {
		println(err.Error())
		return false
	}

	realpath, err := filepath.EvalSymlinks(abs)
	if err != nil {
		println(err.Error())
		return false
	}
	return realpath == abs
}
