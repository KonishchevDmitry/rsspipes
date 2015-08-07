package rsspipes

import (
    "errors"
    "fmt"
    "io/ioutil"
    "mime"
    "net/http"
    "os"
    "strings"

    "golang.org/x/net/html"

    "code.google.com/p/go-charset/charset"
    _ "code.google.com/p/go-charset/data"

    "github.com/KonishchevDmitry/go-rss"
)

type FutureFeedResult struct {
    Feed *rss.Feed
    Err error
}
type FutureFeed chan FutureFeedResult
type FetchFunc func(string) (*rss.Feed, error)

func FetchUrl(url string) (feed *rss.Feed, err error) {
    defer func() { err = handleError(url, err) }()

    feed, err = rss.Get(url)
    if err != nil {
        return
    }

    sortItems(feed)
    return
}

func FetchFile(path string) (feed *rss.Feed, err error) {
    defer func() { err = handleError(path, err) }()

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
    defer func() { err = handleError(url, err) }()

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

func FetchHtml(url string) (data string, err error) {
    defer func() { err = handleError(url, err) }()

    _, data, err = FetchData(url, []string{"text/html"})
    if err != nil {
        return
    }

    doc, err := html.Parse(strings.NewReader(data))
    if err != nil {
        return
    }

    encoding := getHtmlCharset(doc, url)
    if encoding == "" {
        return
    }

    charsetReader, err := charset.NewReader(encoding, strings.NewReader(data))
    if err != nil {
        err = fmt.Errorf("The document has an unknown charset encoding: %s.", encoding)
        return
    }

    decodedBytes, err := ioutil.ReadAll(charsetReader)
    if err != nil {
        err = fmt.Errorf("Failed to decode the document using %s charset: %s", err)
        return
    }

    data = string(decodedBytes)

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

func getHtmlCharset(doc *html.Node, uri string) (charset string) {
    node := findHtmlNode(doc, "html")
    if node != nil {
        node = findHtmlNode(node, "head")
    }
    if node == nil {
        return
    }

    isHttpCharset := true

    for node := node.FirstChild; node != nil; node = node.NextSibling {
        if node.Type != html.ElementNode || node.Data != "meta" {
            continue
        }

        attrs := make(map[string]string)
        for _, attr := range node.Attr {
            attrs[strings.ToLower(attr.Key)] = strings.ToLower(attr.Val)
        }

        if attrs["http-equiv"] == "content-type" {
            _, params, err := mime.ParseMediaType(attrs["content"])

            if err != nil {
                log.Error("Got an invalid content type of '%s' from <meta http-equiv=\"Content-Type\"> tag: '%s'.",
                          uri, attrs["content"])
            } else if params["charset"] != "" && isHttpCharset {
                charset = params["charset"]
            }
        }

        if attrs["charset"] != "" {
            charset = attrs["charset"]
            isHttpCharset = false
        }
    }

    return
}

func findHtmlNode(node *html.Node, name string) *html.Node {
    for node = node.FirstChild; node != nil; node = node.NextSibling {
        if node.Type == html.ElementNode && node.Data == name {
            return node
        }
    }

    return nil
}

func handleError(uri string, err error) error {
    if err != nil {
        err = fmt.Errorf("Failed to fetch %s: %s", uri, err)
        log.Error("%s", err)
    }

    return err
}