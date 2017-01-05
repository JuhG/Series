package main

import (
	"log"
	"net/http"
	"os"

	"encoding/json"
	"errors"
	"fmt"
	"github.com/flosch/pongo2"
	"github.com/gorilla/mux"
	"io/ioutil"
	"sort"
	"strings"
)

var (
	templates = map[string]*pongo2.Template{
		"home":   pongo2.Must(pongo2.FromFile("templates/home.html")),
		"series": pongo2.Must(pongo2.FromFile("templates/series.html")),
		"search": pongo2.Must(pongo2.FromFile("templates/search.html")),
	}
)

const API_KEY = "65b7998bdc0bc00e7fffa6e1d05519dc"

func main() {

	PORT := getPort()
	log.Print("Running server on " + PORT)

	r := mux.NewRouter().StrictSlash(true)
	r.HandleFunc("/", homeHandler)
	r.HandleFunc("/search/", searchHandler)
	r.HandleFunc("/search/{page}/", searchHandler)
	r.HandleFunc("/s/{id:[0-9]+}/", seriesHandler)
	//r.HandleFunc("/s/{id:[0-9]+}/{season:[0-9]+}/", seriesHandler)
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
	Id                 int
	Vote_count         int
	Popularity         float32
	Vote_average       float32
	In_production      bool
	Number_of_seasons  int
	Number_of_episodes int
	Vote_average_string,
	Name,
	Original_name,
	First_air_date,
	Overview string
}

func (s Series) cleanUp() Series {

	// Get the year from the date
	if len(s.First_air_date) > 3 {
		s.First_air_date = s.First_air_date[:4]
	}
	// Change the decimal places to 2
	s.Vote_average_string = fmt.Sprintf("%.2f", s.Vote_average)

	return s
}

type apiResponse struct {
	Page,
	Total_results,
	Total_pages int
	Results []Series
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	err := templates["home"].ExecuteWriter(nil, w)
	if err != nil {
		log.Fatal(err)
	}
}

func seriesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	season := vars["season"]

	related, err := getRelated(id)
	if err != nil {
		related = []Series{}
	}

	series, _ := getDetails(id)
	err = templates["series"].ExecuteWriter(pongo2.Context{
		"s":      series.cleanUp(),
		"season": season,
		//"seasons": seasons,
		"related": related,
	}, w)
	if err != nil {
		log.Fatal(err)
	}
}

/*
TODO: list all the seasons
func seriesHandler(w http.ResponseWriter, r *http.Request) {
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
	if err != nil {
		log.Fatal(err)
	}
}*/

func getDetails(id string) (Series, error) {
	url := "https://api.themoviedb.org/3/tv/" + id + "?api_key=" + API_KEY + "&language=hu-HU"
	payload := strings.NewReader("{}")
	req, _ := http.NewRequest("GET", url, payload)
	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	var series Series
	err := json.Unmarshal(body, &series)
	if err != nil {
		log.Fatal(err)
	}

	// TODO: handle error when no series was found
	//if len(response.Results) == 0 {
	//	return response, errors.New("No series was found!")
	//}

	return series, nil
}

func getRelated(id string) ([]Series, error) {
	url := "https://api.themoviedb.org/3/tv/" + id + "/similar?api_key=" + API_KEY + "&language=hu-HU"
	payload := strings.NewReader("{}")
	req, _ := http.NewRequest("GET", url, payload)
	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	var response apiResponse
	err := json.Unmarshal(body, &response)
	if err != nil {
		log.Fatal(err)
	}

	if len(response.Results) == 0 {
		return []Series{}, errors.New("No series was found!")
	}
	series := response.Results
	sort.Sort(ByPopularity(series))

	return series, nil
}
