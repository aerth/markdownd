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

var addr = flag.String("-http", ":8080", "address to listen on")
const version = "0.0.3"
const sig = "[markdownd v" + version + "]"
const serverheader = "markdownd/"+version

func init() {
	println(sig)
	if len(os.Args) != 2 {
		println("which directory to serve?")
		os.Exit(111)
	}
}

func main() {
	srv := new(Server)
	dir := os.Args[1]
	if dir == "." {
		dir = ""
	}
	if !strings.HasSuffix(dir, "/"){
		dir+="/"
	}
	srv.Root = http.Dir(dir)
	srv.RootString = dir
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
	w.Header().Add("Server", serverheader)
	ctx := context.WithValue(r.Context(), "RequestID", rfid())
	hosts[ctx.Value("RequestID")] = r.RemoteAddr
	r = r.WithContext(ctx)
	path := r.URL.Path[1:] // remove slash
	if path == "" {
		path = "index.md"
	}
	filesrc := s.RootString+path

	rfid := ctx.Value("RequestID")
	log.Println(rfid, r.RemoteAddr, r.Method, r.URL.Path, "->", filesrc)
	if path == "" || strings.HasSuffix(path, "/"){
		log.Printf("%s %s -> %sindex.md", rfid, filesrc, filesrc )
		filesrc += "index.md"
	}

	if strings.HasSuffix(filesrc, ".html"){
		trymd := strings.TrimSuffix(filesrc, ".html")+".md"
		_, err := os.Open(trymd)
		if err == nil {
			log.Println(rfid, filesrc, "->", trymd)
			filesrc = trymd
		}
	}


	defer log.Println(rfid, "closed")

	
	abs, err := filepath.Abs(filesrc)
	if err != nil {
		log.Println(rfid, "error resolving absolute path:", err)
		http.NotFound(w,r)
		return
	}

	_, err = os.Open(abs)
	if err != nil {
		log.Println(rfid, "error opening file:", err, abs)
		http.NotFound(w,r)
		return
	}

	if !fileisgood(abs) {
		log.Printf("%s error: %q is symlink. serving 404", rfid, abs)
		http.NotFound(w,r)
		return
	}
	
	b, err := ioutil.ReadFile(abs)
	if err != nil {
		log.Printf("%s error reading file: %q", rfid, filesrc)
		http.NotFound(w,r)
		return
	}

	// detect content type and encoding
	ct := http.DetectContentType(b)
	if strings.HasPrefix(ct, "text/html") {
		log.Println(rfid, "serving raw html:", abs)
		w.Write(b)
		return
	}

	// probably markdown
	if strings.HasPrefix(ct, "text/plain"){
		log.Println(rfid, "serving markdown:", abs)
		unsafe := blackfriday.MarkdownCommon(b)
		html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)
		w.Write(html)
		return
	}
	
	// fallthrough with http.ServeFile
	log.Printf("%s serving %s: %s", rfid, ct, abs)
	http.ServeFile(w, r, abs)
	w.Write([]byte(sig))
}

// returns false if symlink
// comparing absolute vs resolved path is quick and effective
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
