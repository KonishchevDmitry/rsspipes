package rsspipes

import (
    "errors"
    "fmt"
    "io/ioutil"
    "mime"
    "net/http"
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
    if err != nil {
        logError(url, err)
        return
    }
    sortItems(feed)
    return
}

func FetchFile(path string) (feed *rss.Feed, err error) {
    defer func() { logError(path, err) }()

    file, err := os.Open(path)
    if err != nil {
        return
    }
    defer file.Close()

    feed, err = rss.Read(file)
    if err != nil {
        return
    }

    sortItems(feed)
    return
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

    for i, futureFeed := range futureFeeds {
        futureResult := <-futureFeed
        feeds[i] = futureResult.Feed

        if futureResult.Err != nil {
            err = futureResult.Err
        }
    }

    return
}

func FetchData(url string, allowedMediaTypes []string) (mediaType string, data string, err error) {
    client := &http.Client{
        Timeout: rss.DefaultGetParams.Timeout,
    }

    response, err := client.Get(url)
    if err != nil {
        return
    }
    defer response.Body.Close()

    if response.StatusCode != http.StatusOK {
        err = errors.New(response.Status)
        return
    }

    mediaType, err = checkMediaType(response, allowedMediaTypes)
    if err != nil {
        return
    }

    body, err := ioutil.ReadAll(response.Body)
    if err != nil {
        return
    }

    data = string(body)

    return
}

func checkMediaType(response *http.Response, allowedMediaTypes []string) (mediaType string, err error) {
    contentType := response.Header.Get("Content-Type")
    mediaType, _, err = mime.ParseMediaType(contentType)
    if err != nil {
        err = fmt.Errorf("The document has an invalid Content-Type: %s", err)
        return
    }

    if allowedMediaTypes != nil {
        allowed := false
        for _, allowedMediaType := range allowedMediaTypes {
            if mediaType == allowedMediaType {
                allowed = true
                break
            }
        }

        if !allowed {
            err = fmt.Errorf("The document has an invalid media type (%s).", mediaType)
            return
        }
    }

    return
}

func logError(uri string, err error) {
    if err != nil {
        log.Error("Failed to fetch %s: %s", uri, err)
    }
}