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
	"context"
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
	//	"github.com/shurcooL/github_flavored_markdown"
)

var addr = flag.String("http", ":8080", "address to listen on")
const version = "0.0.3"
const sig = "[markdownd v" + version + "]\nhttps://github.com/aerth/markdownd"
const serverheader = "markdownd/"+version

func init() {
	println(sig)
}

func main() {
	flag.Parse()
	if len(flag.Args()) != 1 {
		println("which directory to serve?")
		os.Exit(111)
	}
	dir := flag.Arg(0)
	if dir == "." {
		dir = ""
	}
	if !strings.HasSuffix(dir, "/"){
		dir+="/"
	}

	srv := &Server{
		Root: http.Dir(dir),
		RootString: dir,
	}
	
	println(http.ListenAndServe(*addr, srv).Error())
}

type Server struct {
	Root http.FileSystem
	RootString string
}

func rfid() string {
	return strconv.Itoa(rand.Int())
}

var hosts = make(map[interface{}]string)

func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		log.Println(r.RemoteAddr, r.Method, r.URL.Path, r.UserAgent())
		http.Error(w, http.StatusText(304), 304)
		return
	}
	w.Header().Add("Server", serverheader)
	requestid := rfid()
	ctx := context.WithValue(r.Context(), "RequestID", requestid)
	hosts[ctx.Value("RequestID")] = r.RemoteAddr

	r = r.WithContext(ctx)
	path := r.URL.Path[1:] // remove slash
	if path == "" {
		path = "index.md"
	}
	filesrc := s.RootString+path
	log.Println(requestid, r.RemoteAddr, r.Method, r.URL.Path, "->", filesrc)
	if path == "" || strings.HasSuffix(path, "/"){
		log.Printf("%s %s -> %sindex.md", requestid, filesrc, filesrc )
		filesrc += "index.md"
	}

	if strings.HasSuffix(filesrc, ".html"){
		trymd := strings.TrimSuffix(filesrc, ".html")+".md"
		_, err := os.Open(trymd)
		if err == nil {
			log.Println(requestid, filesrc, "->", trymd)
			filesrc = trymd
		}
	}

	defer log.Println(requestid, "closed")
	
	abs, err := filepath.Abs(filesrc)
	if err != nil {
		log.Println(requestid, "error resolving absolute path:", err)
		http.NotFound(w,r)
		return
	}

	_, err = os.Open(abs)
	if err != nil {
		log.Println(requestid, "error opening file:", err, abs)
		http.NotFound(w,r)
		return
	}

	if !fileisgood(abs) {
		log.Printf("%s error: %q is symlink. serving 404", requestid, abs)
		http.NotFound(w,r)
		return
	}

	// read bytes
	b, err := ioutil.ReadFile(abs)
	if err != nil {
		log.Printf("%s error reading file: %q", requestid, filesrc)
		http.NotFound(w,r)
		return
	}

	// detect content type and encoding
	ct := http.DetectContentType(b)

	// serve 
	if strings.HasPrefix(ct, "text/html") {
		log.Println(requestid, "serving raw html:", abs)
		w.Write(b)
		return
	}

	// probably markdown
	if strings.HasPrefix(ct, "text/plain"){
		if r.FormValue("raw") != "" || strings.Contains(r.URL.RawQuery,"?raw") {
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
	http.ServeFile(w, r, abs)
	w.Write([]byte(sig))
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
