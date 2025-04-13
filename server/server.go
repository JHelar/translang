package server

import (
	"fmt"
	"net/http"
	"translang/template"

	"github.com/a-h/templ"
)

func ListenAndServe() {
	base := template.Home()

	http.Handle("/", templ.Handler(base))

	fmt.Println("Listening on :3000")
	http.ListenAndServe("127.0.0.1:3000", nil)
}
