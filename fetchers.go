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

func FetchUrl(url string) (feed *rss.Feed, err error) {
    feed, err = rss.Get(url)
    logError(url, err)
    return
}

func FetchFile(path string) (feed *rss.Feed, err error) {
    defer func() { logError(path, err) }()

    file, err := os.Open(path)
    if err != nil {
        return
    }
    defer file.Close()

    return rss.Read(file)
}

func FutureFetch(fetchFunc FetchFunc, uri string) FutureFeed {
    c := make(FutureFeed, 1)

    go func() {
        feed, err := fetchFunc(uri)
        c <- FutureFeedResult{Feed: feed, Err: err}
    }()

    return c
}

func GetFutures(futureFeeds ...FutureFeed) (feeds []*rss.Feed, err error) {
    feeds = make([]*rss.Feed, len(futureFeeds))

    for i, futureFeed := range(futureFeeds) {
        futureResult := <-futureFeed
        feeds[i] = futureResult.Feed

        if futureResult.Err != nil {
            err = futureResult.Err
        }
    }

    return
}

func logError(uri string, err error) {
    if err != nil {
        log.Error("Failed to fetch %s: %s", uri, err)
    }
}