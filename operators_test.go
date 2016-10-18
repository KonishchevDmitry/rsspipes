package rsspipes

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/KonishchevDmitry/go-rss"
)

func TestFilter(t *testing.T) {
	feed := &rss.Feed{
		Items: []*rss.Item{
			newItem("1", 1),
			newItem("2", 2),
			newItem("3", 3),
			newItem("4", 4),
			newItem("5", 5),
		},
	}

	expectedFeed := &rss.Feed{
		Items: []*rss.Item{
			newItem("1", 1),
			newItem("4", 4),
		},
	}

	Filter(feed, func(item *rss.Item) bool {
		switch item.Guid.Id {
		case "2", "3", "5":
			return false
		default:
			return true
		}
	})

	checkResult(t, feed, expectedFeed, "Invalid filter result")
}

func TestFilterWithEmptyFeed(t *testing.T) {
	expectedFeed := &rss.Feed{}
	feed := &rss.Feed{}

	Filter(feed, func(item *rss.Item) bool { return true })
	checkResult(t, feed, expectedFeed, "Filter has modified an empty feed")

	Filter(feed, func(item *rss.Item) bool { return false })
	checkResult(t, feed, expectedFeed, "Filter has modified an empty feed")
}

func TestUnion(t *testing.T) {
	feed1, feed2, expectedFeed, resultFeed := newUnionTestData()
	Union(resultFeed, feed1, feed2)
	checkResult(t, resultFeed, expectedFeed, "Invalid result feed")
}

func TestUnionFutures(t *testing.T) {
	feed1, feed2, expectedFeed, resultFeed := newUnionTestData()

	futureFeed1 := make(FutureFeed, 1)
	futureFeed2 := make(FutureFeed, 1)

	go func() {
		futureFeed1 <- FutureFeedResult{Feed: feed1}
		futureFeed2 <- FutureFeedResult{Feed: feed2}
	}()

	Union(resultFeed, feed1, feed2)
	checkResult(t, resultFeed, expectedFeed, "Invalid result feed")
}

func TestUnionFuturesWithError(t *testing.T) {
	err := errors.New("Mocked error")
	_, feed2, _, resultFeed := newUnionTestData()

	expectedFeed := &rss.Feed{}
	*expectedFeed = *resultFeed

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

	checkResult(t, resultFeed, expectedFeed, "Result feed has been changed")
}

func newUnionTestData() (feed1 *rss.Feed, feed2 *rss.Feed, expectedFeed *rss.Feed, resultFeed *rss.Feed) {
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
		Title:       "Union title",
		Link:        "http://example.com/rss",
		Description: "Union description",

		Items: []*rss.Item{
			feed2.Items[1],
			feed1.Items[0],
			feed1.Items[1],
		},
	}

	resultFeed = &rss.Feed{
		Title:       expectedFeed.Title,
		Link:        expectedFeed.Link,
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

func TestLimit(t *testing.T) {
	feed := &rss.Feed{
		Items: []*rss.Item{
			newItem("1", 1),
			newItem("2", 2),
			newItem("3", 3),
			newItem("4", 4),
			newItem("5", 5),
		},
	}

	origFeed := &rss.Feed{
		Items: []*rss.Item{
			newItem("1", 1),
			newItem("2", 2),
			newItem("3", 3),
			newItem("4", 4),
			newItem("5", 5),
		},
	}

	expectedFeed := &rss.Feed{
		Items: []*rss.Item{
			newItem("3", 3),
			newItem("4", 4),
			newItem("5", 5),
		},
	}

	Limit(feed, 5)
	checkResult(t, feed, origFeed, "Limit has stripped the feed when not expected to")

	Limit(feed, 3)
	checkResult(t, feed, expectedFeed, "Invalid Limit() result")
}

func TestLimitWithEmptyFeed(t *testing.T) {
	feed := &rss.Feed{}
	expectedFeed := &rss.Feed{}

	Limit(feed, 1)
	checkResult(t, feed, expectedFeed, "Limit has modified an empty feed")

	Limit(feed, 0)
	checkResult(t, feed, expectedFeed, "Limit has modified an empty feed")
}

func checkResult(t *testing.T, feed *rss.Feed, expectedFeed *rss.Feed, message string) {
	if !reflect.DeepEqual(feed, expectedFeed) {
		t.Fatalf("%s:\n%s\nvs\n%s", message, feed, expectedFeed)
	}
}
