package reaper

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"git.j3s.sh/vore/sqlite"
	"github.com/SlyMarbo/rss"
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
	fmt.Println("reaper: starting")

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
		r.updateAll()
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
func (r *Reaper) updateAll() {
	start := time.Now()
	fmt.Printf("reaper: fetching %d feeds\n", len(r.feeds))

	var wg sync.WaitGroup
	wg.Add(len(r.feeds))
	for i := range r.feeds {
		go func(i int) {
			defer wg.Done()
			r.updateFeed(&r.feeds[i])
		}(i)
	}
	go func() {
		wg.Wait()
		fmt.Printf("reaper: fetched %d feeds in %s\n", len(r.feeds), time.Since(start))
	}()
}

// updateFeed triggers a fetch on the given feed,
// and sets a fetch error in the db if there is one.
func (r *Reaper) updateFeed(f *rss.Feed) {
	// return early if it's not time to refresh yet
	if !f.Refresh.After(time.Now()) {
		return
	}
	err := f.Update()
	if err != nil {
		fmt.Printf("[err] reaper: fetch failure url '%s' %s\n", f.UpdateURL, err)
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
		return f[i].Title < f[j].Title
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

// FetchFeed attempts to fetch a feed from a given url, marshal
// it into a feed object, and add it to Reaper.
func (r *Reaper) Fetch(url string) error {
	feed, err := rss.Fetch(url)
	if err != nil {
		return err
	}

	r.addFeed(*feed)

	return nil
}
