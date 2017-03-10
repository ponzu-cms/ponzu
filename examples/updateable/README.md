# Updateable

This example shows how to enable outside clients to update content to your CMS.
All content submitted must be done through a POST request encoded as `multipart/form-data`
to the API endpoint `/api/content/update?type=<Type>&id=<id>`

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
1. Implement `api.Updateable` with the `AcceptUpdate(http.ResponseWriter, *http.Request)` method to allow outside POST requests. 
2. Consistent with the externalable example, authentication can be validated in `BeforeAcceptUdate(http.ResponseWriter, *http.Request)`

There are various validation and request checks shown in this example as well. 
Please feel free to modify and submit a PR for updates or bug fixes!

