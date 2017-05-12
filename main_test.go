package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestPrepareDirectory(t *testing.T) {
	dir := "docs"
	dir = prepareDirectory(dir)
	if dir == "docs" {
		t.Log("Wanted absolute pathfrom 'docs', still got 'docs'")
		t.FailNow()
	}

	// give absolute dir
	dir2 := prepareDirectory(dir)
	if dir != dir2 {
		t.Logf("Wanted no change if given absolute path, got: %q != %q", dir, dir2)
		t.FailNow()
	}
}

func TestFileIsGood(t *testing.T) {
	dir := prepareDirectory("docs")
	req := dir + "index.md" // index.md exists
	if !fileisgood(req) {
		// docs is a normal directory, not a symlink. should never be false
		t.Logf("%s should be good, got bad", req)
		t.FailNow()
	}
}

func TestRefuseSymlinks(t *testing.T) {
	dir := prepareDirectory("docs")
	os.Remove(dir + "index.link")
	err := os.Symlink(dir+"index.md", dir+"index.link")
	defer os.Remove(dir + "index.link")
	if err != nil {
		t.Log("Error creating symlink:", err)
		t.FailNow()
	}
	req := dir + "index.link"
	if fileisgood(req) {
		// index.link is a symlink. should never be true
		t.Logf("%s should be good, got bad", req)
		t.FailNow()
	}

}

func TestRefuseDotDots(t *testing.T) {
	dir := prepareDirectory("docs")
	req, _ := http.NewRequest("GET", "/../main.go", nil)
	w := httptest.NewRecorder()
	h := &Handler{
		Root:       http.Dir(dir),
		RootString: dir,
	}
	h.ServeHTTP(w, req)
	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusNotFound {
		t.Log("Expected 404, got:", resp.StatusCode)
		t.FailNow()
	}

	if len(body) != 19 {
		t.Log("Expected 404 body, got:", len(body), string(body))
		t.FailNow()
	}

	if string(body) != "404 page not found\n" {
		t.Logf("Expected %q, got: %q", "404 page not found\n", string(body))
		t.FailNow()
	}
}

func TestHTMLHeaderFooter(t *testing.T) {
	dir := prepareDirectory("docs")
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	h := &Handler{
		Root:       http.Dir(dir),
		RootString: dir,
		header:     []byte("001"),
		footer:     []byte("002"),
	}
	h.ServeHTTP(w, req)
	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		t.Log("Expected 200, got:", resp.StatusCode)
		t.FailNow()
	}

	bodystr := string(body)

	if !strings.HasPrefix(bodystr, "001") {
		t.Log("Expected prefix of '001'")
		t.Fail()
	}

	if !strings.HasSuffix(bodystr, "002") {
		t.Log("Expected suffix of '002'")
		t.Fail()
	}
}
