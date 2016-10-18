package rsspipes

import (
	"os"
	"reflect"
	"testing"
)

var test_rss_file = "testdata/test.rss"
var test_invalid_rss_file = "testdata/invalid.rss"

func TestFetchFile(t *testing.T) {
	feed, err := FetchFile(test_rss_file)
	if feed == nil || err != nil {
		t.Fatalf("Failed to fetch test %s file: %s.", test_rss_file, err)
	}
}

func TestFetchFileUnexisting(t *testing.T) {
	path := "invalid-unexisting-file"

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("File %s exists.")
	}

	feed, err := FetchFile(path)
	if feed != nil || err == nil {
		t.Fatalf("Successfully fetched an unexising file: %v.", feed)
	}
}

func TestFetchFileInvalid(t *testing.T) {
	if _, err := os.Stat(test_invalid_rss_file); err != nil {
		t.Fatalf("Test invalid file is broken: %s.", err)
	}

	feed, err := FetchFile(test_invalid_rss_file)
	if feed != nil || err == nil {
		t.Fatalf("Successfully fetched an invalid RSS file: %v.", feed)
	}
}

func TestFutureFetch(t *testing.T) {
	feed, err := FetchFile(test_rss_file)
	result := <-FutureFetch(FetchFile, test_rss_file)

	if !reflect.DeepEqual(result.Feed, feed) {
		t.Errorf("Feed:\n%s\nvs\n%s", result.Feed, feed)
	}

	if err != result.Err {
		t.Errorf("Error:\n%s\nvs\n%s", result.Err, err)
	}
}
