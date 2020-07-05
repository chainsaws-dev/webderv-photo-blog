package main

import (
	"crypto/sha1"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	uuid "github.com/satori/go.uuid"
)

var tpl *template.Template

func init() {
	tpl = template.Must(template.ParseGlob("templates/*"))
}

func main() {
	http.HandleFunc("/", index)
	http.Handle("/favicon.ico", http.NotFoundHandler())
	http.Handle("/public/", http.StripPrefix("/public", http.FileServer(http.Dir("./public"))))

	if checkExists("cert.pem") && checkExists("key.pem") {
		http.ListenAndServeTLS(":10443", "cert.pem", "key.pem", nil)
	} else {
		http.ListenAndServe(":8080", nil)
	}

}

func checkExists(filename string) bool {

	if _, err := os.Stat(filename); err == nil {
		return true
	} else {
		return false
	}
}

func index(w http.ResponseWriter, req *http.Request) {
	c := getCookie(w, req)

	if req.Method == http.MethodPost {
		f, fh, err := req.FormFile("ufile")

		if err != nil {
			log.Fatalln(err)
		}
		defer f.Close()

		ext := strings.Split(fh.Filename, ".")[1]

		fn := sha1.New()

		io.Copy(fn, f)

		filename := fmt.Sprintf("%x", fn.Sum(nil)) + "." + ext

		path := filepath.Join(".", "public", "pics", filename)

		nf, err := os.Create(path)

		if err != nil {
			log.Fatalln(err)
		}

		defer nf.Close()
		f.Seek(0, 0)

		io.Copy(nf, f)

		c = appendCookie(w, c, path)

	}

	xs := strings.Split(c.Value, "|")

	xscorrect := xs[1:]

	tpl.ExecuteTemplate(w, "index.gohtml", xscorrect)
}

func getCookie(w http.ResponseWriter, req *http.Request) *http.Cookie {
	c, err := req.Cookie("session")

	if err != nil {
		sID, err := uuid.NewV4()

		if err != nil {
			log.Fatalln(err)
		}

		c = &http.Cookie{
			Name:  "session",
			Value: sID.String(),
		}

		http.SetCookie(w, c)
	}

	return c
}

func appendCookie(w http.ResponseWriter, c *http.Cookie, newfn string) *http.Cookie {
	s := c.Value

	if !strings.Contains(s, newfn) {
		s += "|" + newfn
	}

	c.Value = s

	http.SetCookie(w, c)

	return c
}
