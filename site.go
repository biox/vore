package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"git.j3s.sh/vore/lib"
	"git.j3s.sh/vore/reaper"
	"git.j3s.sh/vore/rss"
	"git.j3s.sh/vore/sqlite"
	"golang.org/x/crypto/bcrypt"
)

type Site struct {
	// title of the website
	title string

	// contains every single feed
	reaper *reaper.Reaper

	// site database handle
	db *sqlite.DB
}

// New returns a fully populated & ready for action Site
func New() *Site {
	title := "vore"
	db := sqlite.New(title + ".db")
	s := Site{
		title:  title,
		reaper: reaper.New(db),
		db:     db,
	}
	return &s
}

// rootHandler is our "wildcard handler", so in addition to
// serving /, it also acts as a router for a few arbitrary
// patterns that can't be registered at starttime
// this includes /<username>, static files, and 404
func (s *Site) rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		s.indexHandler(w, r)
		return
	}
	// handles /<username>
	if s.db.UserExists(strings.TrimPrefix(r.URL.Path, "/")) {
		s.userHandler(w, r)
		return
	}
	// handles static files
	file := filepath.Join("files", "static", strings.TrimPrefix(r.URL.Path, "/"))
	if _, err := os.Stat(file); !errors.Is(err, os.ErrNotExist) {
		http.ServeFile(w, r, file)
		return
	}
	// 404
	http.NotFound(w, r)
}

func (s *Site) indexHandler(w http.ResponseWriter, r *http.Request) {
	if !s.methodAllowed(w, r, "GET") {
		return
	}
	if s.loggedIn(r) {
		http.Redirect(w, r, "/"+s.username(r), http.StatusSeeOther)
		return
	}
	s.renderPage(w, r, "index", nil)
}

func (s *Site) changelogHandler(w http.ResponseWriter, r *http.Request) {
	if !s.methodAllowed(w, r, "GET") {
		return
	}
	s.renderPage(w, r, "changelog", nil)
}

func (s *Site) loginHandler(w http.ResponseWriter, r *http.Request) {
	if !s.methodAllowed(w, r, "GET", "POST") {
		return
	}
	if r.Method == "GET" {
		if s.loggedIn(r) {
			http.Redirect(w, r, "/", http.StatusSeeOther)
		} else {
			s.renderPage(w, r, "login", nil)
		}
	}
	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")

		err := s.login(w, username, password)
		if err != nil {
			s.renderErr(w, err.Error(), http.StatusUnauthorized)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

// TODO: make this take a POST only in accordance w/ some spec
func (s *Site) logoutHandler(w http.ResponseWriter, r *http.Request) {
	if !s.methodAllowed(w, r, "GET", "POST") {
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:  "session_token",
		Value: "",
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Site) registerHandler(w http.ResponseWriter, r *http.Request) {
	if !s.methodAllowed(w, r, "POST") {
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")
	err := s.register(username, password)
	if err != nil {
		s.renderErr(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = s.login(w, username, password)
	if err != nil {
		s.renderErr(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Site) userHandler(w http.ResponseWriter, r *http.Request) {
	if !s.methodAllowed(w, r, "GET") {
		return
	}

	username := strings.TrimPrefix(r.URL.Path, "/")
	items := s.reaper.SortFeedItemsByDate(s.reaper.GetUserFeeds(username))
	data := struct {
		User  string
		Items []*rss.Item
	}{
		User:  username,
		Items: items,
	}

	s.renderPage(w, r, "user", data)

}

func (s *Site) settingsHandler(w http.ResponseWriter, r *http.Request) {
	if !s.methodAllowed(w, r, "GET") {
		return
	}
	if !s.loggedIn(r) {
		s.renderErr(w, "", http.StatusUnauthorized)
		return
	}

	var feeds []*rss.Feed
	feeds = s.reaper.GetUserFeeds(s.username(r))
	s.renderPage(w, r, "feeds", feeds)
}

// TODO:
//
//	show diff before submission (like tf plan)
//	check if feed exists in db already?
//	validate that title exists
func (s *Site) feedsSubmitHandler(w http.ResponseWriter, r *http.Request) {
	if !s.methodAllowed(w, r, "POST") {
		return
	}
	if !s.loggedIn(r) {
		s.renderErr(w, "", http.StatusUnauthorized)
		return
	}

	// validate user input
	var validatedURLs []string
	for _, inputURL := range strings.Split(r.FormValue("submit"), "\r\n") {
		inputURL = strings.TrimSpace(inputURL)
		if inputURL == "" {
			continue
		}

		// if the entry is already in reaper, don't validate
		if s.reaper.HasFeed(inputURL) {
			validatedURLs = append(validatedURLs, inputURL)
			continue
		}
		if _, err := url.ParseRequestURI(inputURL); err != nil {
			e := fmt.Sprintf("can't parse url '%s': %s", inputURL, err)
			s.renderErr(w, e, http.StatusBadRequest)
			return
		}
		validatedURLs = append(validatedURLs, inputURL)
	}

	// write to reaper + db
	for _, u := range validatedURLs {
		// if it's in reaper, it's in the db, safe to skip
		if s.reaper.HasFeed(u) {
			continue
		}
		err := s.reaper.Fetch(u)
		if err != nil {
			e := fmt.Sprintf("reaper: can't fetch '%s' %s", u, err)
			s.renderErr(w, e, http.StatusBadRequest)
			return
		}
		s.db.WriteFeed(u)
	}

	// subscribe to all listed feeds exclusively
	s.db.UnsubscribeAll(s.username(r))
	for _, url := range validatedURLs {
		s.db.Subscribe(s.username(r), url)
	}
	http.Redirect(w, r, "/feeds", http.StatusSeeOther)
}

// username fetches a client's username based
// on the sessionToken that user has set. username
// will return "" if there is no sessionToken.
func (s *Site) username(r *http.Request) string {
	cookie, err := r.Cookie("session_token")
	if err == http.ErrNoCookie {
		return ""
	}
	if err != nil {
		log.Println(err)
	}
	username := s.db.GetUsernameBySessionToken(cookie.Value)
	return username
}

func (s *Site) loggedIn(r *http.Request) bool {
	if s.username(r) == "" {
		return false
	}
	return true
}

// login compares the sqlite password field against the user supplied password and
// sets a session token against the supplied writer.
func (s *Site) login(w http.ResponseWriter, username string, password string) error {
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}
	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}
	if !s.db.UserExists(username) {
		return fmt.Errorf("user '%s' does not exist", username)
	}
	storedPassword := s.db.GetPassword(username)
	err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password))
	if err != nil {
		return fmt.Errorf("invalid password")
	}
	sessionToken := lib.GenerateSessionToken()
	err = s.db.SetSessionToken(username, sessionToken)
	if err != nil {
		log.Println(err)
	}

	http.SetCookie(w, &http.Cookie{
		Name: "session_token",
		// 18 years
		Expires: time.Now().Add(time.Hour * 24 * 365 * 18),
		Value:   sessionToken,
	})
	return nil
}

func (s *Site) register(username string, password string) error {
	if s.db.UserExists(username) {
		return fmt.Errorf("user '%s' already exists", username)
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	err = s.db.AddUser(username, string(hashedPassword))
	if err != nil {
		return err
	}
	return nil
}

// renderPage renders the given page and passes data to the
// template execution engine. it's normally the last thing a
// handler should do tbh.
func (s *Site) renderPage(w http.ResponseWriter, r *http.Request, page string, data any) {
	funcMap := template.FuncMap{
		"printDomain": s.printDomain,
		"timeSince":   s.timeSince,
		"trimSpace":   strings.TrimSpace,
	}

	tmplFiles := filepath.Join("files", "*.tmpl.html")
	tmpl := template.Must(template.New("whatever").Funcs(funcMap).ParseGlob(tmplFiles))

	// we read the stylesheet in order to render it inline
	cssFile := filepath.Join("files", "style.css")
	stylesheet, err := ioutil.ReadFile(cssFile)
	if err != nil {
		panic(err)
	}

	// fields on this anon struct are generally
	// pulled out of Data when they're globally required
	// callers should jam anything they want into Data
	pageData := struct {
		Title      string
		Username   string
		LoggedIn   bool
		StyleSheet string
		CutePhrase string
		Data       any
	}{
		Title:      page + " | " + s.title,
		Username:   s.username(r),
		LoggedIn:   s.loggedIn(r),
		StyleSheet: string(stylesheet),
		CutePhrase: s.randomCutePhrase(),
		Data:       data,
	}

	err = tmpl.ExecuteTemplate(w, page, pageData)
	if err != nil {
		s.renderErr(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// printDomain does a best-effort uri parse and
// prints the base domain, otherwise returning the
// unmodified string
func (s *Site) printDomain(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err == nil {
		return parsedURL.Hostname()
	}
	// do our best to trim it manually if url parsing fails
	trimmedStr := strings.TrimSpace(rawURL)
	trimmedStr = strings.TrimPrefix(trimmedStr, "http://")
	trimmedStr = strings.TrimPrefix(trimmedStr, "https://")

	return strings.Split(trimmedStr, "/")[0]
}

func (s *Site) timeSince(t time.Time) string {
	now := time.Now()
	duration := now.Sub(t)

	minutes := int(duration.Minutes())
	hours := int(duration.Hours())
	days := int(duration.Hours() / 24)
	weeks := int(duration.Hours() / (24 * 7))
	months := int(duration.Hours() / (24 * 7 * 4))
	years := int(duration.Hours() / (24 * 7 * 4 * 12))

	if years > 100 {
		return fmt.Sprintf("over 100 years ago ಠ_ಠ")
	} else if years > 1 {
		return fmt.Sprintf("%d years ago", years)
	} else if months > 1 {
		return fmt.Sprintf("%d months ago", months)
	} else if weeks > 1 {
		return fmt.Sprintf("%d weeks ago", weeks)
	} else if days > 1 {
		return fmt.Sprintf("%d days ago", days)
	} else if hours > 1 {
		return fmt.Sprintf("%d hours ago", hours)
	} else if minutes > 1 {
		return fmt.Sprintf("%d mins ago", minutes)
	} else {
		return fmt.Sprintf("just now")
	}
}

// renderErr sets the correct http status in the header,
// optionally decorates certain errors, then renders the err page
func (s *Site) renderErr(w http.ResponseWriter, error string, code int) {
	var prefix string
	switch code {
	case http.StatusBadRequest:
		prefix = "400 bad request\n"
	case http.StatusUnauthorized:
		prefix = "401 unauthorized\n"
	case http.StatusMethodNotAllowed:
		prefix = "405 method not allowed\n"
		prefix += "request method: "
	case http.StatusInternalServerError:
		prefix = "(╥﹏╥) oopsie woopsie, uwu\n"
		prefix += "we made a fucky wucky (╥﹏╥)\n\n"
		prefix += "500 internal server error\n"
	}
	log.Println(prefix + error)
	http.Error(w, prefix+error, code)
}

// methodAllowed takes an http w/r, and returns true if the
// http requests method is in teh allowedMethods list.
// if methodNotAllowed returns false, it has already
// written a request & it's on the caller to close it.
func (s *Site) methodAllowed(w http.ResponseWriter, r *http.Request, allowedMethods ...string) bool {
	allowed := false
	for _, m := range allowedMethods {
		if m == r.Method {
			allowed = true
		}
	}
	if allowed == false {
		w.Header().Set("Allow", strings.Join(allowedMethods, ", "))
		s.renderErr(w, r.Method, http.StatusMethodNotAllowed)
	}
	return allowed
}

func (s *Site) randomCutePhrase() string {
	phrases := []string{
		"devouring feeds since 2023",
		"a no-bullshit feed reader",
	}
	i := rand.Intn(len(phrases))
	return phrases[i]
}
