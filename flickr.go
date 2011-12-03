// Flickr library for Go.
// Created to be used primarily in Google App Engine.
package flickgo

import (
	"fmt"
	"http"
	"os"
	"strconv"
	"strings"
)

// Flickr API permission levels.  See
// http://www.flickr.com/services/api/auth.spec.html.
const (
	ReadPerm   = "read"
	WritePerm  = "write"
	DeletePerm = "delete"
)

// Debug logger.
type Debugfer interface {
	// Debugf formats its arguments according to the format, analogous to fmt.Printf,
	// and records the text as a log message at Debug level.
	Debugf(format string, args ...interface{})
}

// Flickr client.
type Client struct {
	// Auth token for acting on behalf of a user.
	AuthToken string

	// Logger to use.
	// Hint: App engine's Context implements this interface.
	Logger Debugfer

	// API key for your app.
	apiKey string

	// API secret for your app.
	secret string

	// Client to use for HTTP communication.
	httpClient *http.Client
}

// Creates a new Client object.  See
// http://www.flickr.com/services/api/misc.api_keys.html for learning about API
// key and secret.  For App Engine apps, you can create httpClient by calling
// urlfetch.Client function; other apps can pass http.DefaultClient.
func New(apiKey string, secret string, httpClient *http.Client) *Client {
	return &Client{
		apiKey:     apiKey,
		secret:     secret,
		httpClient: httpClient,
	}
}

// Returns the URL for requesting authorisation to access the user's Flickr
// account.  List of possible permissions are defined at
// http://www.flickr.com/services/api/auth.spec.html.  You can also use one of
// the following constants:
//     ReadPerm
//     WritePerm
//     DeletePerm
func (c *Client) AuthURL(perms string) string {
	args := map[string]string{}
	args["perms"] = perms
	return signedURL(c.secret, c.apiKey, "auth", args)
}

// Returns the signed URL for Flickr's flickr.auth.getToken request.
func getTokenURL(c *Client, frob string) string {
	return makeURL(c, "flickr.auth.getToken", map[string]string{"frob": frob}, true)
}

type flickrError struct {
	Code string `xml:"attr"`
	Msg  string `xml:"attr"`
}

func (e *flickrError) Error() os.Error {
	return fmt.Errorf("Flickr error code %s: %s", e.Code, e.Msg)
}

// Exchanges a temporary frob for a token that's valid forever.
// See http://www.flickr.com/services/api/auth.howto.web.html.
func (c *Client) GetToken(frob string) (string, *User, os.Error) {
	r := struct {
		Stat string `xml:"attr"`
		Err  flickrError
		Auth struct {
			Token string
			User  User
		}
	}{}
	if err := flickrGet(c, getTokenURL(c, frob), &r); err != nil {
		return "", nil, err
	}
	if r.Stat != "ok" {
		return "", nil, r.Err.Error()
	}
	return r.Auth.Token, &r.Auth.User, nil
}

// Returns URL for Flickr photo search.
func searchURL(c *Client, args map[string]string) string {
	argsCopy := clone(args)
	argsCopy["extras"] += ",url_t"
	return makeURL(c, "flickr.photos.search", argsCopy, true)
}

// Searches for photos.  args contains search parameters as described in
// http://www.flickr.com/services/api/flickr.photos.search.html.
func (c *Client) Search(args map[string]string) (*SearchResponse, os.Error) {
	r := struct {
		Stat   string `xml:"attr"`
		Err    flickrError
		Photos SearchResponse
	}{}
	if err := flickrGet(c, searchURL(c, args), &r); err != nil {
		return nil, err
	}
	if r.Stat != "ok" {
		return nil, r.Err.Error()
	}

	for i, ph := range r.Photos.Photos {
		h, hErr := strconv.Atof64(ph.Height_T)
		w, wErr := strconv.Atof64(ph.Width_T)
		if hErr == nil && wErr == nil {
			// ph is apparently just a copy of r.Photos.Photos[i], so we are
			// updating the original.
			r.Photos.Photos[i].Ratio = w / h
		}
	}
	return &r.Photos, nil
}

// Initiates an asynchronous photo upload and returns the ticket ID.  See
// http://www.flickr.com/services/api/upload.async.html for details.
func (c *Client) Upload(name string, photo []byte,
args map[string]string) (ticketID string, err os.Error) {
	req, uErr := uploadRequest(c, name, photo, args)
	if uErr != nil {
		return "", wrapErr("request creation failed", uErr)
	}

	resp := struct {
		Stat     string `xml:"attr"`
		Err      flickrError
		TicketID string
	}{}
	if err := flickrPost(c, req, &resp); err != nil {
		return "", wrapErr("uploading failed", err)
	}
	if resp.Stat != "ok" {
		return "", resp.Err.Error()
	}
	return resp.TicketID, nil
}

// Returns URL for flickr.photos.upload.checkTickets request.
func checkTicketsURL(c *Client, tickets []string) string {
	args := make(map[string]string)
	args["tickets"] = strings.Join(tickets, ",")
	return makeURL(c, "flickr.photos.upload.checkTickets", args, false)
}

// Asynchronous photo upload status response.
type TicketStatus struct {
	ID       string `xml:"attr"`
	Complete string `xml:"attr"`
	Invalid  string `xml:"attr"`
	PhotoID  string `xml:"attr"`
}

// Checks the status of async upload tickets (returned by Upload method, for
// example).  Interface for
// http://www.flickr.com/services/api/flickr.photos.upload.checkTickets.html
// API method.
func (c *Client) CheckTickets(tickets []string) (statuses []TicketStatus, err os.Error) {
	r := struct {
		Stat    string `xml:"attr"`
		Err     flickrError
		Tickets []TicketStatus `xml:"uploader>ticket"`
	}{}
	if err := flickrGet(c, checkTicketsURL(c, tickets), &r); err != nil {
		return nil, err
	}
	if r.Stat != "ok" {
		return nil, r.Err.Error()
	}
	return r.Tickets, nil
}

// Returns URL for flickr.photosets.getList request.
func getPhotoSetsURL(c *Client, userID string) string {
	args := make(map[string]string)
	args["user_id"] = userID
	return makeURL(c, "flickr.photosets.getList", args, true)
}

// Returns the list of photo sets of the specified user.
func (c *Client) GetSets(userID string) ([]PhotoSet, os.Error) {
	r := struct {
		Stat string `xml:"attr"`
		Err  flickrError
		Sets []PhotoSet `xml:"photosets>photoset"`
	}{}
	if err := flickrGet(c, getPhotoSetsURL(c, userID), &r); err != nil {
		return nil, err
	}
	if r.Stat != "ok" {
		return nil, r.Err.Error()
	}
	return r.Sets, nil
}

func addToSetURL(c *Client, photoID, setID string) string {
	args := make(map[string]string)
	args["photo_id"] = photoID
	args["photoset_id"] = setID
	return makeURL(c, "flickr.photosets.addPhoto", args, true)
}

// Adds a photo to a photoset.
func (c *Client) AddPhotoToSet(photoID, setID string) os.Error {
	r := struct {
		Stat string `xml:"attr"`
		Err  flickrError
	}{}
	if err := flickrGet(c, addToSetURL(c, photoID, setID), &r); err != nil {
		return err
	}
	if r.Stat != "ok" {
		return r.Err.Error()
	}
	return nil
}
