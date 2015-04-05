package rsspipes

import (
    "sort"
    "github.com/SlyMarbo/rss"
)


type UnionFeedParams struct {
    Title       string
    Description string
    Link        string
    Image       *rss.Image
}


type sortByDate []*rss.Item

func (items sortByDate) Len() int {
    return len(items)
}
func (items sortByDate) Swap(i, j int) {
    items[i], items[j] = items[j], items[i]
}
func (items sortByDate) Less(i, j int) bool {
    return items[i].Date.Unix() < items[j].Date.Unix()
}


func Union(params *UnionFeedParams, feeds ...*rss.Feed) *rss.Feed {
    result := &rss.Feed{
        Title: params.Title,
        Description: params.Description,
        Link: params.Link,
        Image: params.Image,
    }

    uniqueItems := make(map[string]*rss.Item)
    items := make([]*rss.Item, 0)

    for _, feed := range feeds {
        for _, item := range feed.Items {
            if item.ID != "" {
                uniqueItems[item.ID] = item
            } else if item.Link != "" {
                uniqueItems[item.Link] = item
            } else {
                items = append(items, item)
            }
        }
    }

    for _, item := range uniqueItems {
        items = append(items, item)
    }

    var sortedItems sortByDate = items
    sort.Sort(sortedItems)

    result.Items = items
    return result
}


func UnionFutures(params *UnionFeedParams, futureFeeds ...FutureFeed) (result *rss.Feed, err error) {
    feeds := make([]*rss.Feed, len(futureFeeds))

    for i, futureFeed := range(futureFeeds) {
        futureResult := <-futureFeed
        feeds[i] = futureResult.Feed

        if futureResult.Err != nil {
            err = futureResult.Err
        }
    }

    if err != nil {
        result = Union(params, feeds...)
    }

    return
}