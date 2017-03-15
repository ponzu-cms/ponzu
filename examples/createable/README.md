# Createable

This example shows how to enable outside clients to submit content to your CMS.
All content submitted must be done through a POST request encoded as `multipart/form-data`
to the API endpoint `/api/content/create?type=<Type>`

## Song example
Imagine an app that lets users add Spotify music to a global playlist, and you need them
to supply songs in the format:
```go
type Song struct {
	item.Item

	Title      string `json:"title"`
	Artist     string `json:"artist"`
	Rating     int    `json:"rating"`
	Opinion    string `json:"opinion"`
	SpotifyURL string `json:"spotify_url"`
}
```

See the file `content/song.go` and read the comments to understand the various
methods needed to satisfy required interfaces for this kind of activity.

### Overview
1. Implement `api.Createable` with the `Create(http.ResponseWriter, *http.Request) error` method to allow outside POST requests
2. Implement `editor.Mergeable` with the `Approve(http.ResponseWriter, *http.Request) error` method so you can control the Approval / Rejection of submitted content OR
3. Implement `api.Trustable`  with the `AutoApprove(http.ResponseWriter, *http.Request) error` method to bypass `Approve` and auto-approve and publish submitted content

There are various validation and request checks shown in this example as well. 
Please feel free to modify and submit a PR for updates or bug fixes!