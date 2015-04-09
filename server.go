package rsspipes

import (
    "fmt"
    "net/http"
    "github.com/KonishchevDmitry/go-rss"
)

func Register(path string, generator func() (*rss.Feed, error)) {
    http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
        generate(w, r, generator)
    })
}

func Serve() {
    http.ListenAndServe(":8080", nil)
}

func generate(w http.ResponseWriter, r *http.Request, generator func() (*rss.Feed, error)) {
    feed, err := generator()
    if err != nil {
        error := fmt.Sprintf("Failed to generate the RSS feed: %s.", err)
        http.Error(w, error, http.StatusInternalServerError)
        return
    }

    // FIXME
    rss, err := rss.Generate(feed)

    w.Header().Set("Content-Type", "application/rss+xml")
    w.Write([]byte(rss))
}