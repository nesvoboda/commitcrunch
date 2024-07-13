package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/Khan/genqlient/graphql"
	"github.com/charmbracelet/log"
	"github.com/go-chi/chi/v5"
)

type handler struct {
	client graphql.Client
	cache  cache
}

type authedTransport struct {
	key     string
	wrapped http.RoundTripper
}

func (t *authedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "bearer "+t.key)
	return t.wrapped.RoundTrip(req)
}

func main() {
	log.Info("starting")
	key := os.Getenv("GITHUB_TOKEN")
	if key == "" {
		panic(fmt.Errorf("must set GITHUB_TOKEN=<github token>"))
	}

	httpClient := http.Client{
		Transport: &authedTransport{
			key:     key,
			wrapped: http.DefaultTransport,
		},
	}
	graphqlClient := graphql.NewClient("https://api.github.com/graphql", &httpClient)

	handler := handler{
		client: graphqlClient,
		cache:  newCache(),
	}

	r := chi.NewRouter()

	r.HandleFunc("/contribs/{username}", func(w http.ResponseWriter, r *http.Request) {
		username := chi.URLParam(r, "username")

		if username == "" {
			log.Error("Empty username")
			http.Error(w, http.StatusText(404), 404)
			return
		}

		log.Info("Request", "url", r.URL.Path)

		days, ok := handler.cache.get(username)
		var err error
		if !ok {
			days, err = handler.getDays(username)
			if err != nil {
				log.Error("Error getting contributions", "error", err)
				http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
				return
			}
			handler.cache.set(username, days)
		}

		bytes, err := json.Marshal(days)

		if err != nil {
			log.Error("Error marshaling", "error", err)
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusOK)
		w.Write(bytes)
		log.Info("Request successful")
	})

	r.HandleFunc("/summary/{username}", func(w http.ResponseWriter, r *http.Request) {
		username := chi.URLParam(r, "username")

		if username == "" {
			log.Error("Empty username")
			http.Error(w, http.StatusText(404), 404)
			return
		}

		log.Info("Request", "url", r.URL.Path)

		summary, err := handler.getSummary(username)
		if err != nil {
			log.Error("Error getting contributions", "error", err)
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
			return
		}

		bytes, err := json.Marshal(summary)

		if err != nil {
			log.Error("Error marshaling", "error", err)
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusOK)
		w.Write(bytes)
		log.Info("Request successful")

	})

	log.Info("Starting the router")
	http.ListenAndServe(":8080", r)

}

type day struct {
	Date     string `json:"date"`
	Contribs int    `json:"contribs"`
}

type cache struct {
	mu   *sync.Mutex
	data map[string][]day
}

const MAX_CACHE_LEN = 100

func newCache() cache {
	c := cache{
		mu:   &sync.Mutex{},
		data: make(map[string][]day, MAX_CACHE_LEN),
	}

	go func() {
		time.Sleep(time.Hour)
		c.clear()
	}()

	return c
}

func (c cache) get(key string) ([]day, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, ok := c.data[key]
	if ok {
		log.Info("Cache hit", "key", key)
	} else {
		log.Info("Cache miss", "key", key)
	}
	return item, ok
}

func (c cache) set(key string, value []day) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.data) > MAX_CACHE_LEN {
		c.data = make(map[string][]day)
	}

	c.data[key] = value
}

func (c cache) clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	log.Info("Cache cleared")

	c.data = make(map[string][]day, MAX_CACHE_LEN)
}

func (h handler) getDays(username string) ([]day, error) {
	ctx := context.Background()

	resp, err := getContributions(ctx, h.client, username)

	if err != nil {
		return nil, err
	}

	days := make([]day, len(resp.User.ContributionsCollection.ContributionCalendar.Weeks)*7)

	for i := range resp.User.ContributionsCollection.ContributionCalendar.Weeks {
		week := resp.User.ContributionsCollection.ContributionCalendar.Weeks[i]
		for d := range week.ContributionDays {
			days[i*7+d] = day{
				Date:     week.ContributionDays[d].Date,
				Contribs: week.ContributionDays[d].ContributionCount,
			}
		}
	}

	return days, nil
}

type summary struct {
	Years []year `json:"years"`
}

type year struct {
	Current  bool `json:"current"`
	Contribs int  `json:"contribs"`
	Value    int  `json:"value"`
}

// todo batch years

func (h handler) getSummary(username string) (summary, error) {
	ctx := context.Background()

	years, err := getContributionYears(ctx, h.client, username)
	if err != nil {
		return summary{}, err
	}

	yrs := years.User.ContributionsCollection.ContributionYears

	sum := summary{
		Years: make([]year, len(yrs)),
	}

	for i, y := range yrs {
		start := time.Date(y, time.January, 1, 0, 0, 0, 0, time.UTC)
		sum.Years[i].Value = y

		data, err := getContributionsForYear(ctx, h.client, username, start)
		if err != nil {
			return summary{}, err
		}

		if y == time.Now().Year() {
			sum.Years[i].Current = true
		}

		sum.Years[i].Contribs = data.User.ContributionsCollection.ContributionCalendar.TotalContributions
	}
	return sum, nil
}
