package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"os"

	"github.com/flosch/pongo2"
)

var (
	tmpl = map[string]*pongo2.Template{
		"home": pongo2.Must(pongo2.FromFile("templates/home.html")),
		"series": pongo2.Must(pongo2.FromFile("templates/series.html")),
	}
)

func main() {

	PORT := getPort()
	log.Print("Running server on " + PORT)
	http.HandleFunc("/s/", getMovieFromTitle)
	http.HandleFunc("/search/", searchHandler)
	http.HandleFunc("/", frontHandler)

	log.Fatal(http.ListenAndServe(":" + PORT, nil))
}

func getPort() (PORT string) {
	PORT = os.Getenv("PORT")
	if PORT == "" {
		PORT = "8080"
	}
	return
}

type Series struct {
	Response, Title, Plot, Year string
}

func getMovieFromTitle(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[len("/s/"):]
	s, _ := getData(title)
	err := tmpl["series"].ExecuteWriter(pongo2.Context{
		"s": s,
	}, w)
	if err != nil {
		log.Fatal(err)
	}
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("search")
	s, err := getData(title)
	if err != nil {
		http.Redirect(w, r, r.Referer(), http.StatusFound)
	}
	title = strings.Replace(s.Title, " ", "+", -1)
	http.Redirect(w, r, "/s/"+title, http.StatusFound)
}

func frontHandler(w http.ResponseWriter, r *http.Request) {
	err := tmpl["home"].ExecuteWriter(nil, w)
	if err != nil {
		log.Fatal(err)
	}
}

func getData(search string) (Series, error) {
	// encode the spaces
	title := strings.Replace(search, " ", "+", -1)

	res, err := http.Get("http://www.omdbapi.com/?type=series&tomatoes=true&t=" + title)
	if err != nil {
		log.Fatal(err)
	}
	data, err := ioutil.ReadAll(res.Body)
	res.Body.Close()

	var s Series
	err = json.Unmarshal(data, &s)
	if err != nil {
		log.Fatal(err)
	}

	if "False" == s.Response {
		return s, errors.New("No response found")
	}
	return s, nil
}
