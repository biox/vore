package reaper

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"
	"unsafe"

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
		log.Println("reaper: refreshing all feeds")
		r.refreshAllFeeds()
		log.Println("reaper: refreshed all feeds, sleeping 😴")
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
	ch := make(chan *rss.Feed)
	var wg sync.WaitGroup
	// i chose 20 workers somewhat arbitrarily
	for i := 20; i > 0; i-- {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for f := range ch {
				start := time.Now()
				r.refreshFeed(f)
				log.Printf("reaper: %s refreshed in %s\n", f.UpdateURL, time.Since(start))
			}
		}()
	}

	for i := range r.feeds {
		if r.feeds[i].Stale() {
			ch <- r.feeds[i]
		}
	}

	close(ch)
	wg.Wait()
}

// refreshFeed triggers a fetch on the given feed,
// and sets a fetch error in the db if there is one.
func (r *Reaper) refreshFeed(f *rss.Feed) {
	err := f.Update()
	if err != nil {
		r.handleFeedFetchFailure(f.UpdateURL, err)
	}
	fmt.Printf("examining %s (size %d) - items(%d):\n", f.UpdateURL, unsafe.Sizeof(f), len(f.Items))
	for _, c := range f.Categories {
		fmt.Printf("  category %s: %d\n", c, unsafe.Sizeof(c))
	}
	for _, c := range f.Items {
		fmt.Printf("  item struct %s: %d\n", c.Title, unsafe.Sizeof(c))
		fmt.Printf("  item Title: %d\n", unsafe.Sizeof(c.Title))
		fmt.Printf("  item Summary: %d\n", unsafe.Sizeof(c.Summary))
	}
}

func (r *Reaper) handleFeedFetchFailure(url string, err error) {
	log.Printf("reaper: failed to fetch %s: %s\n", url, err)
	err = r.db.SetFeedFetchError(url, err.Error())
	if err != nil {
		log.Printf("reaper: could not set feed fetch error '%s'\n", err)
	}
}

// HasFeed checks whether a given url is represented
// in the reaper cache.
func (r *Reaper) HasFeed(url string) bool {
	if _, ok := r.feeds[url]; ok {
		return true
	}
	return false
}

func (r *Reaper) GetFeed(url string) *rss.Feed {
	return r.feeds[url]
}

// GetItem recurses through all rss feeds, returning the first
// found feed by matching against the provided link
func (r *Reaper) GetItem(url string) (*rss.Item, error) {
	for _, f := range r.feeds {
		for _, i := range f.Items {
			if i.Link == url {
				return i, nil
			}
		}
	}
	return &rss.Item{}, errors.New("item not found")
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

// SortFeeds sorts reaper feeds chronologically by date
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
