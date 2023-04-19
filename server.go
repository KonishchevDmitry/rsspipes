package rsspipes

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	logging "github.com/KonishchevDmitry/go-easy-logging"
	"github.com/KonishchevDmitry/go-rss"
)

var rootRegistered = false

func Serve(ctx context.Context, addressPort string) error {
	logging.L(ctx).Infof("Listening on %s...", addressPort)

	if !rootRegistered {
		register("/", http.NotFound)
	}

	server := http.Server{
		Addr: addressPort,
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}

	return server.ListenAndServe()
}

func Register(path string, generator func(context.Context) (*rss.Feed, error)) {
	register(path, func(w http.ResponseWriter, r *http.Request) {
		generate(w, r, generator)
	})
}

func register(path string, handler func(http.ResponseWriter, *http.Request)) {
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logging.L(ctx).Infof("%s %s", r.Method, r.RequestURI)
		handler(w, r)
	})

	if path == "/" {
		rootRegistered = true
	}
}

func generate(w http.ResponseWriter, r *http.Request, generator func(context.Context) (*rss.Feed, error)) {
	ctx := r.Context()
	feed, err := generator(ctx)

	if err == nil {
		postprocessFeed(feed)
		writeFeed(w, feed)
	} else {
		logging.L(ctx).Errorf("Failed to generate %s RSS feed: %s", r.RequestURI, err)

		if temporaryErr, ok := err.(temporary); ok && temporaryErr.Temporary() {
			writeError(w, err)
		} else {
			message := "rsspipes feed generation error"
			writeFeed(w, &rss.Feed{
				Title: message,
				Items: []*rss.Item{{
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
