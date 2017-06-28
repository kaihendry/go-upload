package main

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/pkg/profile"
)

func upload(w http.ResponseWriter, r *http.Request) {

	r.ParseMultipartForm(32 << 20) // Not quite sure what this should be

	file, handler, err := r.FormFile("movie")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer file.Close()

	f, err := ioutil.TempFile("", "upload")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if _, err := io.Copy(f, file); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if err := f.Close(); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	log.Println(handler)
	fmt.Printf("Upload writen to %v\n", f.Name())
	w.Write([]byte("OK"))
}

func main() {

	defer profile.Start().Stop()

	mux := http.NewServeMux()
	mux.HandleFunc("/", index)
	mux.HandleFunc("/upload", upload)

	log.Println("Listening on :3001")
	err := http.ListenAndServe(":3001", mux)
	log.Fatal(err)

}

func index(w http.ResponseWriter, r *http.Request) {

	t, err := template.New("foo").Parse(`<!DOCTYPE html>
<html>
<head>
<title>Simplest upload example</title>
<meta charset="utf-8" />
<meta name=viewport content="width=device-width, initial-scale=1">
</head>
<body>

<form action="/upload" enctype="multipart/form-data" method="post">
<input type="file" name="movie" />
<input type="submit" value="Send" />
</form>

</body>
</html>`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t.Execute(w, t)

}
