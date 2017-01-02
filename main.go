package main

import (
	"encoding/json"
	"errors"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"os"
)

func main() {
	PORT := os.Getenv("PORT")
	log.Print("Running server on " + PORT)
	http.HandleFunc("/s/", getMovieFromTitle)
	http.HandleFunc("/search/", searchHandler)
	http.HandleFunc("/", frontHandler)

	err := http.ListenAndServe(":" + PORT, nil)
	if err != nil {
		log.Fatal(err)
	}
}

type Series struct {
	Response, Title, Plot, Year string
}

func getMovieFromTitle(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[len("/s/"):]
	s, _ := getData(title)
	t, _ := template.ParseFiles("templates/series.html")
	t.Execute(w, s)
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
	t, _ := template.ParseFiles("templates/home.html")
	t.Execute(w, nil)
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
