package wayback

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
)

type Client struct {
	httpClient *http.Client
}

const userAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.97 Safari/537.36"

var (
	host = "archive.org"
	dest = "https://web." + host
	base = "https://web.archive.org/save/"

	endpoint = "https://archive.org/wayback/available"
)

// Wayback is the handle of saving webpages to archive.org
func (wbrc *Client) Archive(ctx context.Context, u string) (result string, err error) {
	if wbrc.httpClient == nil {
		wbrc.httpClient = &http.Client{
			CheckRedirect: noRedirect,
		}
	}

	result, err = wbrc.archive(ctx, u)
	if err != nil {
		return
	}
	return
}

func (wbrc *Client) archive(ctx context.Context, uri string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+uri, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("User-Agent", userAgent)
	resp, err := wbrc.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var loc string
	loc = resp.Header.Get("Content-Location")

	if len(loc) > 0 {
		return loc, nil
	}

	loc = resp.Header.Get("Location")
	if len(loc) > 0 {
		return loc, nil
	}

	links := resp.Header.Get("Link")
	re := regexp.MustCompile(`(?m)http[s]?:\/\/web\.archive\.org/web/[-a-zA-Z0-9@:%_\+.~#?&//=]*`)
	if match := re.FindAllString(links, -1); len(match) > 0 {
		loc = match[len(match)-1]
		return fmt.Sprintf("%v", loc), nil
	}

	loc = resp.Request.URL.String()
	if match := re.FindAllString(loc, -1); len(match) > 0 {
		return fmt.Sprintf("%v", loc), nil
	}

	loc, err = wbrc.latest(ctx, uri)
	if err != nil {
		loc = base + uri
	}

	// HTTP 509 Bandwidth Limit Exceeded
	if resp.StatusCode == 509 {
		return fmt.Sprint(loc), nil
	}

	if resp.StatusCode != 200 {
		return fmt.Sprint(loc), nil
	}

	return loc, nil
}

func (wbrc *Client) latest(_ context.Context, u string) (string, error) {
	// https://web.archive.org/*/https://example.org
	result := fmt.Sprintf("%s/*/%s", dest, u)

	uri := endpoint + "?url=" + u
	resp, err := wbrc.httpClient.Get(uri)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var dat map[string]interface{}
	if err := json.Unmarshal(data, &dat); err != nil {
		return "", err
	}

	if archived, ok := dat["archived_snapshots"].(map[string]interface{}); ok {
		if closest, ok := archived["closest"].(map[string]interface{}); ok {
			if closest["available"].(bool) && closest["status"] == "200" {
				return closest["url"].(string), nil
			}
		}
	}

	return result, fmt.Errorf("Not found")
}

func noRedirect(req *http.Request, via []*http.Request) error {
	return http.ErrUseLastResponse
}
