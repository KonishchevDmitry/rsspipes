package rsspipes

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"

	"github.com/PuerkitoBio/goquery"

	"github.com/KonishchevDmitry/go-rss"
)

type FutureFeedResult struct {
	Feed *rss.Feed
	Err  error
}
type FutureFeed chan FutureFeedResult
type FetchFunc func(string) (*rss.Feed, error)

func FetchUrl(url string) (feed *rss.Feed, err error) {
	return FetchUrlWithParams(url, rss.GetParams{})
}

func FetchUrlWithParams(url string, params rss.GetParams) (feed *rss.Feed, err error) {
	defer func() { err = handleError(url, err) }()

	feed, err = rss.GetWithParams(url, params)
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

	client := rss.ClientFromParams(rss.GetParams{})

	response, err := client.Get(url)
	if err != nil {
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		if response.StatusCode >= 500 && response.StatusCode < 600 {
			err = &temporaryError{response.Status}
		} else {
			err = errors.New(response.Status)
		}
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

func FetchHtml(url string) (doc *goquery.Document, err error) {
	defer func() { err = handleError(url, err) }()

	_, data, err := FetchData(url, []string{"text/html"})
	if err != nil {
		return
	}

	htmlDoc, err := parseHtml(url, data)
	if err != nil {
		return
	}

	doc = goquery.NewDocumentFromNode(htmlDoc)

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

func parseHtml(url string, data string) (doc *html.Node, err error) {
	doc, err = html.Parse(strings.NewReader(data))
	if err != nil {
		return
	}

	encoding := strings.ToLower(getHtmlCharset(doc, url))
	if encoding == "" || encoding == "utf-8" && encoding == "utf8" {
		return
	}

	charsetReader, err := charset.NewReaderLabel(encoding, strings.NewReader(data))
	if err != nil {
		err = fmt.Errorf("The document has an unknown charset encoding: %s.", encoding)
		return
	}

	decodedBytes, err := ioutil.ReadAll(charsetReader)
	if err != nil {
		err = fmt.Errorf("Failed to decode the document using %s charset: %s", err)
		return
	}

	doc, err = html.Parse(bytes.NewReader(decodedBytes))

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
				log.Errorf("Got an invalid content type of '%s' from <meta http-equiv=\"Content-Type\"> tag: '%s'.",
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
		// Note: net/url.Error is also implements temporary interface
		err = &fetchError{Uri: uri, Err: err}
		log.Errorf("%s", err)
	}

	return err
}

type fetchError struct {
	Uri string
	Err error
}

func (e *fetchError) Temporary() bool {
	// MacOS doesn't differentiate "no such host" error from DNS lookup errors, so add this workaround here
	if runtime.GOOS == "darwin" {
		if urlError, ok := e.Err.(*url.Error); ok {
			if netErr, ok := urlError.Err.(*net.OpError); ok {
				if dnsErr, ok := netErr.Err.(*net.DNSError); ok && dnsErr.Err == "no such host" {
					return true
				}
			}
		}
	}

	err, ok := e.Err.(temporary)
	return ok && err.Temporary()
}

func (e *fetchError) Error() string {
	return fmt.Sprintf("Failed to fetch %s: %s", e.Uri, e.Err)
}

type temporaryError struct {
	message string
}

func (*temporaryError) Temporary() bool {
	return true
}

func (e *temporaryError) Error() string {
	return e.message
}
