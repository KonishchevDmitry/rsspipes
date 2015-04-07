package rss

import "testing"

var minimalRss = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
    <channel>
        <title>Feed title</title>
        <link>http://example.com/</link>
        <description>Feed description</description>
    </channel>
</rss>`

var fullRss = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
    <channel>
        <title>Feed title</title>
        <link>http://example.com/</link>
        <description>Feed description</description>
        <image>
            <url>http://example.com/logo.png</url>
            <title>Logo title</title>
            <link>http://example.com/</link>
            <width>100</width>
        </image>
        <language>en-us</language>
        <pubDate>Sat, 04 Apr 2015 00:00:00 GMT</pubDate>
        <category>feed-cat1</category>
        <category>feed-cat2</category>
        <item>
            <title>Item 1</title>
            <guid isPermaLink="true">http://example.com/item1</guid>
            <link>http://example.com/item1</link>
            <description>Item 1 description</description>
            <enclosure url="http://example.com/item1/podcast.mp3" type="audio/mpeg" length="123456789"></enclosure>
            <comments>http://example.com/item1/comments</comments>
            <pubDate>Sat, 04 Apr 2015 07:00:13 GMT</pubDate>
            <author>author1</author>
            <category>item-cat1</category>
            <category>item-cat2</category>
        </item>
        <item>
            <title>Item 2</title>
            <guid isPermaLink="false">2e17b013-f283-45e4-b010-5a03ad6776c6</guid>
        </item>
        <item>
            <title>Item 3</title>
            <guid>http://example.com/item3</guid>
        </item>
        <item></item>
    </channel>
</rss>`

func TestParseMinimal(t *testing.T) {
    testParse(t, minimalRss)
}

func TestParseFull(t *testing.T) {
    testParse(t, fullRss)
}

func testParse(t *testing.T, data string) {
    feed, err := Decode([]byte(data))
    if err != nil {
        t.Fatal(err)
    }

    generatedData, err := Encode(feed)
    if err != nil {
        t.Fatal(err)
    }

    if string(generatedData) != data {
        t.Fatalf("Feeds don't match:\n%s\nvs\n%s", generatedData, data)
    }
}