package rsspipes

import (
    "io/ioutil"
    "github.com/SlyMarbo/rss"
)

func FetchFile(path string) (feed *rss.Feed, err error) {
    data, err := ioutil.ReadFile(path)
    if err != nil {
        return
    }
    return rss.Parse(data)
}