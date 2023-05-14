package reaper

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"git.j3s.sh/vore/rss"
	"git.j3s.sh/vore/sqlite"
)

type Reaper struct {
	// internal list of all rss feeds where the map
	// key represents the url of the feed (which should be unique)
	feeds map[string]*rss.Feed

	db *sqlite.DB
}

func New(db *sqlite.DB) *Reaper {
	r := &Reaper{
		feeds: make(map[string]*rss.Feed),
		db:    db,
	}

	go r.start()

	return r
}

// Start initializes the reaper by populating a list of feeds from the database
// and periodically refreshes all feeds every 15 minutes, if the feeds are
// stale.
// reaper should only ever be started once (in New)
func (r *Reaper) start() {
	urls := r.db.GetAllFeedURLs()

	for _, url := range urls {
		// Setting UpdateURL lets us defer fetching
		feed := &rss.Feed{
			UpdateURL: url,
		}
		r.feeds[url] = feed
	}

	for {
		r.refreshAllFeeds()
		time.Sleep(15 * time.Minute)
	}
}

// Add the given rss feed to Reaper for maintenance.
func (r *Reaper) addFeed(f *rss.Feed) {
	r.feeds[f.UpdateURL] = f
}

// UpdateAll fetches every feed & attempts updating them
// asynchronously, then prints the duration of the sync
func (r *Reaper) refreshAllFeeds() {
	start := time.Now()
	var wg sync.WaitGroup
	for i := range r.feeds {
		if r.feeds[i].Stale() {
			wg.Add(1)
			go func(f *rss.Feed) {
				defer wg.Done()
				r.refreshFeed(f)
			}(r.feeds[i])
		}
	}
	wg.Wait()
	fmt.Printf("reaper: refresh complete in %s\n", time.Since(start))
}

// refreshFeed triggers a fetch on the given feed,
// and sets a fetch error in the db if there is one.
func (r *Reaper) refreshFeed(f *rss.Feed) {
	err := f.Update()
	if err != nil {
		r.handleFeedFetchFailure(f.UpdateURL, err)
	}
}

func (r *Reaper) handleFeedFetchFailure(url string, err error) {
	fmt.Printf("[err] reaper: fetch failure '%s': %s\n", url, err)
	err = r.db.SetFeedFetchError(url, err.Error())
	if err != nil {
		fmt.Printf("[err] reaper: could not set feed fetch error '%s'\n", err)
	}
}

// Have checks whether a given url is represented
// in the reaper cache.
func (r *Reaper) HasFeed(url string) bool {
	for _, f := range r.feeds {
		if f.UpdateURL == url {
			return true
		}
	}
	return false
}

// GetUserFeeds returns a list of feeds
func (r *Reaper) GetUserFeeds(username string) []*rss.Feed {
	urls := r.db.GetUserFeedURLs(username)
	var result []*rss.Feed
	for _, u := range urls {
		// feeds in the db are guaranteed to be in reaper
		result = append(result, r.feeds[u])
	}

	r.SortFeeds(result)
	return result
}

func (r *Reaper) SortFeeds(f []*rss.Feed) {
	sort.Slice(f, func(i, j int) bool {
		return f[i].UpdateURL < f[j].UpdateURL
	})
}

func (r *Reaper) SortFeedItemsByDate(feeds []*rss.Feed) []*rss.Item {
	var posts []*rss.Item
	for _, f := range feeds {
		for _, i := range f.Items {
			posts = append(posts, i)
		}
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date.After(posts[j].Date)
	})
	return posts
}

// Fetch attempts to fetch a feed from a given url, marshal
// it into a feed object, and manage it via reaper.
func (r *Reaper) Fetch(url string) error {
	feed, err := rss.Fetch(url)
	if err != nil {
		return err
	}

	r.addFeed(feed)

	return nil
}
