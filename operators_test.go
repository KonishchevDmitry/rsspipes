package rsspipes

import (
    "reflect"
    "time"
    "testing"
    "github.com/SlyMarbo/rss"
)

func TestUnion(t *testing.T) {
    feed1 := &rss.Feed{
        Items: []*rss.Item{
            &rss.Item{ID: "2", Date: time.Time{}.Add(2 * time.Second)},
            &rss.Item{ID: "non-unique", Date: time.Time{}.Add(3 * time.Second)},
        },
    }

    feed2 := &rss.Feed{
        Items: []*rss.Item{
            &rss.Item{ID: "non-unique", Date: time.Time{}.Add(3 * time.Second)},
            &rss.Item{ID: "1", Date: time.Time{}.Add(1 * time.Second)},
        },
    }

    expectedFeed := &rss.Feed{
        Title: "Union title",
        Link: "http://example.com/rss",
        Description: "Union description",

        Items: []*rss.Item{
            feed2.Items[1],
            feed1.Items[0],
            feed1.Items[1],
        },
    }

    result := Union(&UnionFeedParams{
        Title: expectedFeed.Title,
        Link: expectedFeed.Link,
        Description: expectedFeed.Description,
    }, feed1, feed2)

    if !reflect.DeepEqual(result, expectedFeed) {
        t.Fatalf("Items:\n%v\nvs\n%v", result, expectedFeed)
    }
}