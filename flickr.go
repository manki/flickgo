// Copyright 2011 Muthukannan T <manki@manki.in>. All Rights Reserved.

// Flickr library for Go.
// Created to be used primarily in Google App Engine.
package flickgo

import (
  "http"
  "io"
  "io/ioutil"
  "os"
)

// Flickr API permission levels.  See
// http://www.flickr.com/services/api/auth.spec.html.
const (
  ReadPerm = "read"
  WritePerm = "write"
  DeletePerm = "delete"
)

// Flickr client.
type Client struct {
  // Auth token for acting on behalf of a user.
  AuthToken string

  // API key for your app.
  apiKey string

  // API secret for your app.
  secret string

  // Client to use for HTTP communication.
  httpClient *http.Client

  // Indirection for ioutil.ReadAll; tests should stub out this field if they
  // use a fake io.Reader.
  readFn func(r io.Reader) (buf []byte, err os.Error)
}

// Creates a new Client object.  See
// http://www.flickr.com/services/api/misc.api_keys.html for learning about API
// key and secret.  For App Engine apps, you can create httpClient by calling
// urlfetch.Client function; other apps can pass http.DefaultClient.
func New(apiKey string, secret string, httpClient *http.Client) *Client {
  return &Client{
       apiKey: apiKey,
       secret: secret,
       httpClient: httpClient,
       readFn: ioutil.ReadAll,
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
  return url(c, "flickr.auth.getToken", map[string]string{ "frob": frob })
}

// Exchanges a temporary frob for a token that's valid forever.
// See http://www.flickr.com/services/api/auth.howto.web.html.
func (c *Client) GetToken(frob string) (string, os.Error) {
  r := struct {
    Auth struct {
      Token struct { Content string "_content" }
    }
  }{}
  if err := flickrGet(c, getTokenURL(c, frob), &r); err != nil {
    return "", err
  }
  return r.Auth.Token.Content, nil
}

// Returns URL for Flickr photo search.
func searchURL(c *Client, args map[string]string) string {
  return url(c, "flickr.photos.search", args)
}

// Searches for photos.  args contains search parameters as described in
// http://www.flickr.com/services/api/flickr.photos.search.html.
func (c *Client) Search(args map[string]string) (*SearchResponse, os.Error) {
  r := struct {
    Photos SearchResponse
  }{}
  if err := flickrGet(c, searchURL(c, args), &r); err != nil {
    return nil, err
  }
  return &r.Photos, nil
}

// Initiates an asynchronous photo upload and returns the ticket ID.  See
// http://www.flickr.com/services/api/upload.async.html for details.
func (c *Client) Upload(name string, photo []byte,
                        args map[string]string) (ticketId string, err os.Error) {
  req, uErr := uploadRequest(c, name, photo, args)
  if (uErr != nil) {
    return "", wrapErr("request creation failed", uErr)
  }

  resp := struct {
    TicketID string
  }{}
  if err := flickrPost(c, req, &resp); err != nil {
    return "", wrapErr("uploading failed", err)
  }
  return resp.TicketID, nil
}
