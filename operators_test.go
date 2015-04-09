package rsspipes

import (
    "errors"
    "reflect"
    "time"
    "testing"

    "github.com/KonishchevDmitry/go-rss"
)

func TestUnion(t *testing.T) {
    feed1, feed2, expectedFeed, resultFeed := newTestData()

    Union(resultFeed, feed1, feed2)

    if !reflect.DeepEqual(resultFeed, expectedFeed) {
        t.Fatalf("Invalid result feed:\n%s\nvs\n%s", resultFeed, expectedFeed)
    }
}

func TestUnionFutures(t *testing.T) {
    feed1, feed2, expectedFeed, resultFeed := newTestData()

    futureFeed1 := make(FutureFeed, 1)
    futureFeed2 := make(FutureFeed, 1)

    go func() {
        futureFeed1 <- FutureFeedResult{Feed: feed1}
        futureFeed2 <- FutureFeedResult{Feed: feed2}
    }()

    Union(resultFeed, feed1, feed2)

    if !reflect.DeepEqual(resultFeed, expectedFeed) {
        t.Fatalf("Invalid result feed:\n%s\nvs\n%s", resultFeed, expectedFeed)
    }
}

func TestUnionFuturesWithError(t *testing.T) {
    err := errors.New("Mocked error")
    _, feed2, _, resultFeed := newTestData()

    resultFeedCopy := &rss.Feed{}
    *resultFeedCopy = *resultFeed

    futureFeed1 := make(FutureFeed, 1)
    futureFeed2 := make(FutureFeed, 1)

    go func() {
        futureFeed1 <- FutureFeedResult{Err: err}
        futureFeed2 <- FutureFeedResult{Feed: feed2}
    }()

    unionErr := UnionFutures(resultFeed, futureFeed1, futureFeed2)
    if unionErr != err {
        t.Fatalf("Invalid union error: %s.", unionErr)
    }

    if !reflect.DeepEqual(resultFeed, resultFeedCopy) {
        t.Fatalf("Result feed has been changed:\n%s\nvs\n%s", resultFeed, resultFeedCopy)
    }
}

func newTestData() (feed1 *rss.Feed, feed2 *rss.Feed, expectedFeed *rss.Feed, resultFeed *rss.Feed) {
    feed1 = &rss.Feed{
        Items: []*rss.Item{
            newItem("2", 2),
            newItem("non-unique", 3),
        },
    }

    feed2 = &rss.Feed{
        Items: []*rss.Item{
            newItem("non-unique", 3),
            newItem("1", 1),
        },
    }

    expectedFeed = &rss.Feed{
        Title: "Union title",
        Link: "http://example.com/rss",
        Description: "Union description",

        Items: []*rss.Item{
            feed2.Items[1],
            feed1.Items[0],
            feed1.Items[1],
        },
    }

    resultFeed = &rss.Feed{
        Title: expectedFeed.Title,
        Link: expectedFeed.Link,
        Description: expectedFeed.Description,
    }

    return
}

func newItem(id string, date time.Duration) *rss.Item {
    return &rss.Item{
        Guid: rss.Guid{Id: id},
        Date: rss.Date{time.Time{}.Add(date * time.Second)},
    }
}