package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/flosch/pongo2"
	"github.com/gorilla/mux"
	"strconv"
)

var (
	tmpl = map[string]*pongo2.Template{
		"home":   pongo2.Must(pongo2.FromFile("templates/home.html")),
		"series": pongo2.Must(pongo2.FromFile("templates/series.html")),
	}
)

func main() {

	PORT := getPort()
	log.Print("Running server on " + PORT)

	r := mux.NewRouter().StrictSlash(true)
	r.HandleFunc("/", homeHandler)
	r.HandleFunc("/search/", searchHandler)
	r.HandleFunc("/s/{title}/", seriesHandler)
	r.HandleFunc("/s/{title}/{season:[0-9]+}/", seriesHandler)
	http.Handle("/", r)

	log.Fatal(http.ListenAndServe(":"+PORT, nil))
}

func getPort() (PORT string) {
	PORT = os.Getenv("PORT")
	if PORT == "" {
		PORT = "8080"
	}
	return
}

type Series struct {
	Response,
	Title,
	Plot,
	Year,
	TotalSeasons string
}

func seriesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	title := vars["title"]
	season := vars["season"]

	series, _ := getData(title)
	totalSeasons, err := strconv.ParseInt(series.TotalSeasons, 10, 0)
	if err != nil {
		log.Fatal(err)
	}
	if season != "" {
		number, err := strconv.ParseInt(season, 10, 0)
		if err != nil {
			log.Fatal(err)
		}
		if number > totalSeasons || number == 0 {
			w.Write([]byte("Nincs is annyi évad te kis trükkös"))
			return
		}
	}

	seasons := make([]int, totalSeasons, totalSeasons)
	for i := 0; i < int(totalSeasons); i++ {
		seasons[i] = i + 1
	}
	err = tmpl["series"].ExecuteWriter(pongo2.Context{
		"s":       series,
		"season":  season,
		"seasons": seasons,
	}, w)
	if err != nil {
		log.Fatal(err)
	}
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("search")
	series, err := getData(title)
	if err != nil {
		http.Redirect(w, r, r.Referer(), http.StatusFound)
	}

	title = titleToURL(series.Title)
	url := "/s/" + title + "/"
	http.Redirect(w, r, url, http.StatusFound)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	err := tmpl["home"].ExecuteWriter(nil, w)
	if err != nil {
		log.Fatal(err)
	}
}

func getData(search string) (Series, error) {
	title := titleToURL(search)
	res, err := http.Get("http://www.omdbapi.com/?type=series&t=" + title)
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

func titleToURL(search string) (title string) {
	title = search
	title = strings.Replace(title, " ", "+", -1)
	title = strings.ToLower(title)
	return
}
