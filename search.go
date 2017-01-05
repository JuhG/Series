package main

import (
	"encoding/json"
	"errors"
	"github.com/flosch/pongo2"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

func searchHandler(w http.ResponseWriter, r *http.Request) {
	search := r.FormValue("search")

	var response string
	series, err := getAllResults(search)
	if err != nil {
		response = err.Error()
	}

	err = templates["search"].ExecuteWriter(pongo2.Context{
		"search":   search,
		"results":  series,
		"response": response,
	}, w)
	if err != nil {
		log.Fatal(err)
	}
}

// The sorting function
type ByPopularity []Series

func (s ByPopularity) Len() int {
	return len(s)
}
func (s ByPopularity) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByPopularity) Less(i, j int) bool {
	return s[i].Popularity > s[j].Popularity
}

func getAllResults(search string) (series []Series, err error) {
	data, err := getSearchResult(search, 1)
	if err == nil {
		series = extractResults(series, data)
	}
	if data.Total_pages > 1 {
		for i := 2; i < (data.Total_pages + 1); i++ {
			data, err := getSearchResult(search, i)
			if err == nil {
				series = extractResults(series, data)
			}
			if i > 5 {
				break
			}
		}
	}
	// Sorting by popularity
	sort.Sort(ByPopularity(series))
	return
}

func extractResults(series []Series, data apiResponse) []Series {
	for _, s := range data.Results {
		// Filter the results so that we only show results with more than one vote
		if s.Vote_count > 1 {
			series = append(series, s.cleanUp())
		}
	}
	return series
}

func getSearchResult(search string, page int) (apiResponse, error) {
	title := titleToURL(search)
	pageString := strconv.FormatInt(int64(page), 10)
	url := "https://api.themoviedb.org/3/search/tv?api_key=" + API_KEY + "&page=" + pageString + "&language=hu-HU&query=" + title
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
		return response, errors.New("No series was found!")
	}

	return response, nil
}

func titleToURL(search string) string {
	search = strings.Replace(search, " ", "+", -1)
	search = strings.ToLower(search)
	return search
}
