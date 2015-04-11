package rsspipes

import (
    "os"
    "github.com/KonishchevDmitry/go-rss"
)

type FutureFeedResult struct {
    Feed *rss.Feed
    Err error
}
type FutureFeed chan FutureFeedResult
type FetchFunc func(string) (*rss.Feed, error)

func FetchFile(path string) (feed *rss.Feed, err error) {
    file, err := os.Open(path)
    if err != nil {
        return
    }
    defer file.Close()
    return rss.Read(file)
}

func FutureFetch(fetchFunc FetchFunc, url string) FutureFeed {
    c := make(FutureFeed, 1)

    go func() {
        feed, err := fetchFunc(url)
        c <- FutureFeedResult{Feed: feed, Err: err}
    }()

    return c
}