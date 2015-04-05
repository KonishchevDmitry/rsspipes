package rsspipes

import (
    "fmt"
    "net/http"
    "github.com/SlyMarbo/rss"
    "github.com/gorilla/feeds"
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

    resultItems := make([]*feeds.Item, len(feed.Items))

    for i, item := range feed.Items {
        resultItems[i] = &feeds.Item{
            Id:          item.ID,
            Title:       item.Title,
            Link:        &feeds.Link{Href: item.Link},
            Description: item.Content,
            Created:     item.Date,
        }
    }

    resultFeed := feeds.Feed{
        Title:       feed.Title,
        Link:        &feeds.Link{Href: feed.Link},
        Description: feed.Description,
        Items:       resultItems,
    }

    rss, err := resultFeed.ToRss()

    w.Header().Set("Content-Type", "application/rss+xml")
    w.Write([]byte(rss))
}