package reaper

import (
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"git.j3s.sh/vore/rss"
	"git.j3s.sh/vore/sqlite"
)

type Reaper struct {
	// internal list of all rss feeds where the map
	// key represents the primary id of the key in the db
	feeds []rss.Feed

	// this mutex is used for locking writes to Feeds
	mu sync.Mutex

	db *sqlite.DB
}

func Summon(db *sqlite.DB) *Reaper {
	var feeds []rss.Feed

	reaper := Reaper{
		feeds: feeds,
		mu:    sync.Mutex{},
		db:    db,
	}
	return &reaper
}

func (r *Reaper) Start() {
	// Make initial url list
	urls := r.db.GetAllFeedURLs()

	for _, url := range urls {
		// Setting UpdateURL lets us defer fetching
		feed := rss.Feed{
			UpdateURL: url,
		}
		r.feeds = append(r.feeds, feed)
	}

	for {
		r.refreshAll()
		time.Sleep(15 * time.Minute)
	}
}

// Add the given rss feed to Reaper for maintenance.
// If the given feed is already in reaper.Feeds, Add does nothing.
func (r *Reaper) addFeed(f rss.Feed) {
	if !r.HasFeed(f.UpdateURL) {
		r.mu.Lock()
		r.feeds = append(r.feeds, f)
		r.mu.Unlock()
	}
}

// UpdateAll fetches every feed & attempts updating them
// asynchronously, then prints the duration of the sync
func (r *Reaper) refreshAll() {
	start := time.Now()
	var wg sync.WaitGroup
	for i := range r.feeds {
		if r.staleFeed(&r.feeds[i]) {
			wg.Add(1)
			go func(f *rss.Feed) {
				defer wg.Done()
				r.refreshFeed(f)
			}(&r.feeds[i])
		}
	}
	wg.Wait()
	fmt.Printf("reaper: refresh complete in %s\n", time.Since(start))
}

func (r *Reaper) staleFeed(f *rss.Feed) bool {
	if f.Refresh.After(time.Now()) {
		return false
	}
	return true
}

// updateFeed triggers a fetch on the given feed,
// and sets a fetch error in the db if there is one.
func (r *Reaper) refreshFeed(f *rss.Feed) {
	err := f.UpdateByFunc(reaperFetchFunc)
	if err != nil {
		fmt.Printf("[err] reaper: fetch failure '%s': %s\n", f.UpdateURL, err)
		r.db.SetFeedFetchError(f.UpdateURL, err.Error())
	}
}

// Have checks whether a given url is represented
// in the reaper cache.
func (r *Reaper) HasFeed(url string) bool {
	for i := range r.feeds {
		if r.feeds[i].UpdateURL == url {
			return true
		}
	}
	return false
}

// GetUserFeeds returns a list of feeds
func (r *Reaper) GetUserFeeds(username string) []rss.Feed {
	urls := r.db.GetUserFeedURLs(username)
	var result []rss.Feed
	for i := range r.feeds {
		for _, url := range urls {
			if r.feeds[i].UpdateURL == url {
				result = append(result, r.feeds[i])
			}
		}
	}

	r.SortFeeds(result)
	return result
}

func (r *Reaper) SortFeeds(f []rss.Feed) {
	sort.Slice(f, func(i, j int) bool {
		return f[i].UpdateURL < f[j].UpdateURL
	})
}

func (r *Reaper) SortFeedItemsByDate(f []rss.Feed) []rss.Item {
	var posts []rss.Item
	for _, f := range f {
		for _, i := range f.Items {
			posts = append(posts, *i)
		}
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date.After(posts[j].Date)
	})
	return posts
}

var reaperFetchFunc = func(url string) (*http.Response, error) {
	client := http.DefaultClient
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// escape un-parseable HTML elements
	body := string(bodyBytes)
	body = strings.ReplaceAll(body, "&ndash;", "")
	//	body = strings.ReplaceAll(body, "&&amp;", "&#8211;")

	//	body = strings.ReplaceAll(body, "&reg;", "")
	//	fmt.Println(b.String())
	//	if err != nil {
	//		return nil, err
	//	}
	// body := html.UnescapeString(string(bodyBytes))

	// re-encode as XML
	// err = xml.EscapeText(&b, []byte(body))

	// overwrite the body
	resp.Body.Close()
	resp.Body = io.NopCloser(strings.NewReader(body))

	return resp, nil
}

// FetchFeed attempts to fetch a feed from a given url, marshal
// it into a feed object, and add it to Reaper.
func (r *Reaper) Fetch(url string) error {
	feed, err := rss.FetchByFunc(reaperFetchFunc, url)
	if err != nil {
		return err
	}

	r.addFeed(*feed)

	return nil
}
