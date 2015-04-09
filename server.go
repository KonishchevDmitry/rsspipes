package rsspipes

import (
    "fmt"
    "net/http"

    "github.com/KonishchevDmitry/go-rss"

    "rsspipes/util"
)

var rootRegistered = false
var log = util.MustGetLogger("server")

func Serve(addressPort string) error {
    log.Info("Listening on %s...", addressPort)

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
        log.Info("%s %s", r.Method, r.RequestURI)
        handler(w, r)
    })

    if path == "/" {
        rootRegistered = true
    }
}

func generate(w http.ResponseWriter, r *http.Request, generator func() (*rss.Feed, error)) {
    var rssData []byte

    feed, err := generator()
    if err == nil {
        rssData, err = rss.Generate(feed)
    }

    if err != nil {
        error := fmt.Sprintf("Failed to generate the RSS feed: %s.", err)
        http.Error(w, error, http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/rss+xml")
    w.Write(rssData)
}