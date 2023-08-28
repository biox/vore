package reaper

import (
	"testing"

	"git.j3s.sh/vore/rss"
	"git.j3s.sh/vore/sqlite"
)

func TestHasFeed(t *testing.T) {
	db := sqlite.New("go_test.db")
	r := New(db)
	f1 := rss.Feed{UpdateURL: "something"}
	f2 := rss.Feed{UpdateURL: "strange"}
	r.addFeed(&f1)
	r.addFeed(&f2)
	if r.HasFeed("banana") == true {
		t.Fatal("reaper should not have a banana")
	}
	if r.HasFeed("something") == false {
		t.Fatal("reaper should have something")
	}
	if r.HasFeed("strange") == false {
		t.Fatal("reaper should have strange")
	}
}
