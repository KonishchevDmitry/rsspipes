package rsspipes

import (
	"fmt"
	"net/http"
	"time"

	"github.com/KonishchevDmitry/go-rss"
)

var rootRegistered = false

func Serve(addressPort string) error {
	log.Infof("Listening on %s...", addressPort)

	if !rootRegistered {
		register("/", http.NotFound)
	}

	return http.ListenAndServe(addressPort, nil)
}

func Register(path string, generator func() (*rss.Feed, error)) {
	register(path, func(w http.ResponseWriter, r *http.Request) {
		generate(w, r, generator)
	})
}

func register(path string, handler func(http.ResponseWriter, *http.Request)) {
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		log.Infof("%s %s", r.Method, r.RequestURI)
		handler(w, r)
	})

	if path == "/" {
		rootRegistered = true
	}
}

func generate(w http.ResponseWriter, r *http.Request, generator func() (*rss.Feed, error)) {
	feed, err := generator()

	if err == nil {
		postprocessFeed(feed)
		writeFeed(w, feed)
	} else {
		log.Errorf("Failed to generate %s RSS feed: %s", r.RequestURI, err)

		if temporaryErr, ok := err.(temporary); ok && temporaryErr.Temporary() {
			writeError(w, err)
		} else {
			message := "rsspipes feed generation error"
			writeFeed(w, &rss.Feed{
				Title: message,
				Items: []*rss.Item{&rss.Item{
					Title:       message,
					Guid:        rss.Guid{Id: "rsspipes-error-" + time.Now().UTC().Format("02-01-2006")},
					Description: err.Error(),
				}},
			})
		}
	}
}

func postprocessFeed(feed *rss.Feed) {
	isPermaLink := true

	for _, item := range feed.Items {
		guid := &item.Guid
		if guid.Id == "" && item.Link != "" {
			guid.Id = item.Link
			guid.IsPermaLink = &isPermaLink
		}
	}
}

func writeFeed(w http.ResponseWriter, feed *rss.Feed) {
	data, err := rss.Generate(feed)

	if err == nil {
		w.Header().Set("Content-Type", rss.ContentType)
		w.Write(data)
	} else {
		writeError(w, err)
	}
}

func writeError(w http.ResponseWriter, err error) {
	http.Error(w, fmt.Sprintf("Failed to generate the RSS feed: %s", err), http.StatusInternalServerError)
}
