package rsspipes

import (
    "sort"
    "github.com/KonishchevDmitry/go-rss"
)


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


func Filter(feed *rss.Feed, filter func(*rss.Item) (allow bool)) {
    size := 0
    items := feed.Items

    for _, item := range(items) {
        if filter(item) {
            items[size] = item
            size++
        }
    }

    if size != len(items) {
        feed.Items = items[:size]
    }
}


func Union(result *rss.Feed, feeds ...*rss.Feed) {
    items := result.Items
    uniqueItems := make(map[string]*rss.Item)

    for _, feed := range feeds {
        for _, item := range feed.Items {
            if item.Guid.Id != "" {
                uniqueItems[item.Guid.Id] = item
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
}

func UnionFutures(result *rss.Feed, futureFeeds ...FutureFeed) error {
    feeds, err := GetFutures(futureFeeds...)
    if err != nil {
        return err
    }

    Union(result, feeds...)
    return nil
}