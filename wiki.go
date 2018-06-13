package main

import (
  // "fmt"
  "html/template"
  "io/ioutil"
  "log"
  "net/http"
  "regexp"
)

// wiki consists of series of interconnected pages
// each of which has a title and body
type Page struct {
  Title string
  Body []byte
}

// panics when passes non-nil error
// otherwise returns *Template unaltered
var templates = template.Must(template.ParseFiles("edit.html", "view.html"))

// will panic vs MustCompile will return err as second param
// var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

// ============MIDDLEWARE========================
// if write successful, will return nil
// (the zero-value for pointers, interfaces, and some other types)
// 0600 read write permission for current user only
func (p *Page) save() error {
  filename := p.Title + ".txt"
  return ioutil.WriteFile(filename, p.Body, 0600)
}

// func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
//   m := validPath.FindStringSubmatch(r.URL.Path)
//   if m == nil {
//     http.NotFound(w, r)
//     return "", errors.New("Invalid Page Title")
//   }
//   return m[2], nil
// }

// load pages by returning new variable Page and pointing to it
func loadPage(title string) (*Page, error) {
  filename := title + ".txt"
  body, err := ioutil.ReadFile(filename)
  if err != nil {
    return nil, err
  }
  return &Page{Title: title, Body: body}, nil
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
  err := templates.ExecuteTemplate(w, tmpl+".html", p)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
  // t, err := template.ParseFiles(tmpl + ".html")
  // if err != nil {
  //   http.Error(w, err.Error(), http.StatusInternalServerError)
  //   return
  // }
  // err = t.Execute(w, p)
  // if err != nil {
  //   http.Error(w, err.Error(), http.StatusInternalServerError)
  // }
}

// =============HANDLERS=======================
// wrapper
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    m := validPath.FindStringSubmatch(r.URL.Path)
    if m == nil {
      http.NotFound(w, r)
      return
    }
    fn(w, r, m[2])
  }
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
  p, err := loadPage(title)
  if err != nil {
    http.Redirect(w, r, "/edit/"+title, http.StatusFound)
    return
  }
  renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
  p, err := loadPage(title)
  if err != nil {
    p = &Page{Title: title}
  }
  renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
  body := r.FormValue("body")
  p := &Page{Title: title, Body: []byte(body)}
  err := p.save()
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  http.Redirect(w, r, "/view/" + title, http.StatusFound)
}

// ==============MAIN===========================
// ListenAndServe will only return when unexpected error occurs
func main() {
  http.HandleFunc("/view/", makeHandler(viewHandler))
  http.HandleFunc("/edit/", makeHandler(editHandler))
  http.HandleFunc("/save/", makeHandler(saveHandler))
  log.Fatal(http.ListenAndServe(":8080", nil))
}

// func main() {
//   p1 := &Page{Title: "TestPage", Body: []byte("This is a sample Page.")}
//   p1.save()
//   p2, _ := loadPage("TestPage")
//   fmt.Println(string(p2.Body))
// }
