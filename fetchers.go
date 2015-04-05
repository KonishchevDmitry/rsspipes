package rsspipes

import (
    "io/ioutil"
    "github.com/SlyMarbo/rss"
)

type FutureFeedResult struct {
    Feed *rss.Feed
    Err error
}
type FutureFeed chan FutureFeedResult
type FetchFunc func(string) (*rss.Feed, error)

func FetchFile(path string) (feed *rss.Feed, err error) {
    data, err := ioutil.ReadFile(path)
    if err != nil {
        return
    }
    return rss.Parse(data)
}

func FutureFetch(fetchFunc FetchFunc, url string) FutureFeed {
    c := make(FutureFeed)

    go func() {
        feed, err := fetchFunc(url)
        c <- FutureFeedResult{Feed: feed, Err: err}
    }()

    return c
}