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
		// Setting UpdateURL lets us defer the actual fetching
		feed := rss.Feed{
			UpdateURL: url,
		}
		r.feeds = append(r.feeds, feed)
	}

	for {
		r.UpdateAll()
		time.Sleep(2 * time.Hour)
	}
}

// Add fetches the given feed url and appends it to r.Feeds
// If the given URL is already in reaper.Feeds, Add will do nothing
func (r *Reaper) Add(url string) error {
	for i := range r.feeds {
		if r.feeds[i].UpdateURL == url {
			return nil
		}
	}

	feed, err := rss.Fetch(url)
	if err != nil {
		return err
	}

	r.mu.Lock()
	r.feeds = append(r.feeds, *feed)
	r.mu.Unlock()

	return nil
}

// UpdateAll fetches every feed & attempts updating them
func (r *Reaper) UpdateAll() {
	start := time.Now()
	fmt.Printf("reaper: fetching %d feeds\n", len(r.feeds))
	for i := range r.feeds {
		err := r.feeds[i].Update()
		if err != nil {
			fmt.Println(err)
			// TODO: write err to db?
		}
	}
	fmt.Printf("reaper: fetched %d feeds in %s\n",
		len(r.feeds), time.Since(start))
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
	return result
}

func (r *Reaper) SortFeedItems(f []rss.Feed) []rss.Item {
	var posts []rss.Item
	for _, f := range f {
		for _, i := range f.Items {
			posts = append(posts, *i)
		}
	}

	// magick slice sorter by date
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date.After(posts[j].Date)
	})
	return posts
}
