package main

import (
	"html/template"
	"io/ioutil"
	"net/http"
	"regexp"
)

var templates *template.Template = template.Must(template.ParseFiles("view.html", "edit.html"))
var validPath = regexp.MustCompile("^/(?:edit|save|view)/([a-zA-Z0-9]+)$")

const pagesDirRoot = "pages"

type Page struct {
	Title string
	Body  []byte
}

func (p *Page) save() error {
	filename := pagesDirRoot + "/" + p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := pagesDirRoot + "/" + title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func main() {
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	_ = http.ListenAndServe(":8080", nil)
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		m := validPath.FindStringSubmatch(request.URL.Path)
		if m == nil {
			http.NotFound(writer, request)
			return
		}
		title := m[1]
		fn(writer, request, title)

	}
}

func viewHandler(writer http.ResponseWriter, request *http.Request, title string) {
	page, err := loadPage(title)
	if err != nil {
		http.Redirect(writer, request, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(writer, "view", page)
}

func editHandler(writer http.ResponseWriter, _ *http.Request, title string) {
	page, err := loadPage(title)
	if err != nil {
		page = &Page{Title: title}
	}
	renderTemplate(writer, "edit", page)
}

func saveHandler(writer http.ResponseWriter, request *http.Request, title string) {
	body := request.FormValue("body")
	page := &Page{Title: title, Body: []byte(body)}
	err := page.save()
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(writer, request, "/view/"+title, http.StatusFound)
}

func renderTemplate(writer http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(writer, tmpl+".html", p)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}
