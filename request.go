// Copyright 2011 Muthukannan T <manki@manki.in>. All Rights Reserved.

package flickgo

import (
  "bytes"
  "crypto/md5"
  "fmt"
  "http"
  "io"
  "json"
  "log"
  "multipart_writer"
  "os"
  "path/filepath"
  "regexp"
  "sort"
  "strings"
)

const (
  service = "http://www.flickr.com/services"
  uploadURL = "http://api.flickr.com/services/upload"
)

// Returns all keys of map m.
func keys(m map[string]string) sort.StringArray {
  ks := make([]string, len(m))
  i := 0
  for k, _ := range m {
    ks[i] = k
    i++
  }
  return ks
}

// Converts a map[string]string to a map[string][]string by boxing each value
// into a single-element array.
func multimap(m map[string]string) map[string][]string {
  r := make(map[string][]string)
  for k, v := range m {
    r[k] = []string{v}
  }
  return r
}

// Clones a string -> string map.
func clone(m map[string]string) map[string]string {
  r := make(map[string]string)
  for k, v := range m {
    r[k] = v
  }
  return r
}

func wrapErr(msg string, err os.Error) os.Error {
  return os.NewError(msg + ": " + err.String())
}

// Returns an API signature for the given arguments.
func sign(secret string, args map[string]string) string {
  ks := keys(args)
  ks.Sort()
  m := md5.New()
  m.Write([]byte(secret))
  for _, k := range ks {
    m.Write([]byte(k + http.URLEscape(args[k])))
  }
  return fmt.Sprintf("%x", m.Sum())
}

// Returns a signed URL.  path should be "auth" for auth requests and "rest"
// for all other requests.  args specifies the query arguments.  Signing of the
// URL is done by adding "api_sig" argument to the query string, whose value is
// derived by signing the query values with secret.
func signedURL(secret string, apiKey string, path string, args map[string]string) string {
  a := clone(args)
  a["api_key"] = apiKey
  a["api_sig"] = sign(secret, a)
  qry := http.EncodeQuery(multimap(a))
  return fmt.Sprintf("%s/%s/?%s", service, path, qry)
}

// Returns a URL for invoking a Flickr method with the specified arguments.  If
// c has its AuthToken field set, the auth token is added to the URL.  Returned
// URL is always signed with c.secret.
func url(c *Client, method string, args map[string]string) string {
  a := clone(args)
  a["method"] = method
  a["format"] = "json"
  if len(c.AuthToken) > 0 {
    a["auth_token"] = c.AuthToken
  }
  return signedURL(c.secret, c.apiKey, "rest", a)
}

// Regular expressions for identifying non-JSON part of the JSONP response
// returned by Flickr.
var (
  begin = regexp.MustCompile(`^[ \t\n]*jsonFlickrApi\(`)
  end = regexp.MustCompile(`\)[ \t\n]*$`)
)

// Extracts JSON data from the JSONP returned by Flickr.
func extractJSON(jsonp []byte) []byte {
  empty := []byte("")
  t := begin.ReplaceAll(jsonp, empty)
  return end.ReplaceAll(t, empty)
}

// Processes a response and returns JSON content from it.
func processReponse(c *Client, r *http.Response) ([]byte, os.Error) {
  // TODO: handle error response codes like 401 and 500.

  defer r.Body.Close()
  buf, readErr := c.readFn(r.Body)
  if readErr != nil {
    return nil, readErr
  }
  return extractJSON(buf), nil
}

func parseJSON(data []byte, resp interface{}) os.Error {
  // Try to parse the response as error.  Both success and failure responses
  // have a 'stat' field; if stat != "ok" the request was successful.
  r := flickrError{}
  if err := json.Unmarshal(data, &r); err != nil {
    return os.NewError(err.String() + "; JSON=" + string(data))
  }
  if r.Stat != "ok" {
    return os.NewError(fmt.Sprintf("Flickr error code %d: %s",
                                   r.Code, r.Message))
  }

  if err := json.Unmarshal(data, resp); err != nil {
    return os.NewError(err.String() + "; JSON=" + string(data))
  }
  return nil
}

// Sends a GET request to u and returns the response JSON.
func fetch(c *Client, u string) ([]byte, os.Error) {
  r, _, getErr := c.httpClient.Get(u)
  if getErr != nil {
    return nil, getErr
  }
  return processReponse(c, r)
}

// Flickr's JSON error objects have the same structure.
type flickrError struct {
  Stat string
  Code int
  Message string
}

// Sends a Flickr request, parses the response JSON and populates values in
// resp.  url represents the complete Flickr request with the arguments signed
// with the API secret.
func flickrGet(c *Client, url string, resp interface{}) os.Error {
  data, err := fetch(c, url)
  if err != nil {
    return err
  }
  log.Printf("JSON received: %s", string(data))
  return parseJSON(data, resp)
}

func flickrPost(c *Client, req *http.Request, resp interface{}) os.Error {
  r, rErr := c.httpClient.Do(req)
  if rErr != nil {
    return rErr
  }
  data, pErr := processReponse(c, r)
  if pErr != nil {
    return pErr
  }
  log.Printf("JSON received: %s", string(data))
  return parseJSON(data, resp)
}

var contentType = map[string]string{
  ".jpg": "image/jpeg",
  ".jpeg": "image/jpeg",
  ".jpe": "image/jpeg",
  ".gif": "image/gif",
  ".png": "image/png",
}

func multipartWriter(w io.Writer, filename string, photo []byte,
                     args map[string]string) (*multipart_writer.Writer, os.Error) {
  mpw := multipart_writer.NewWriter(w)
  for k, v := range args {
    if err := mpw.WriteField(k, v); err != nil {
      return nil, wrapErr(fmt.Sprintf("field write failed [%v=%v]", k, v), err)
    }
  }
  w, cErr := mpw.CreateFormFile("photo", filename,
                                contentType[strings.ToLower(filepath.Ext(filename))])
  if cErr != nil {
    return nil, wrapErr("form file creation failed [" + filename + "]", cErr)
  }
  if _, err := w.Write(photo); err != nil {
    return nil, wrapErr("adding photo data failed", err)
  }
  if err := mpw.Close(); err != nil {
    return nil, wrapErr("multipart close failed", err)
  }
  return mpw, nil
}

func uploadRequest(c *Client, filename string, photo []byte,
                   args map[string]string) (*http.Request, os.Error) {
  a := clone(args)
  a["api_key"] = c.apiKey
  a["auth_token"] = c.AuthToken
  a["format"] = "json"
  a["async"] = "1"
  a["api_sig"] = sign(c.secret, a)

  buf := bytes.NewBuffer(make([]byte, len(photo) * 2))
  mpw, wErr := multipartWriter(buf, filename, photo, a)
  if wErr != nil {
    return nil, wrapErr("writer creation failed", wErr)
  }

  req, rErr := http.NewRequest("POST", uploadURL, buf)
  if rErr != nil {
    return nil, wrapErr("request creation failed", rErr)
  }
  req.Header.Set("Content-Type", mpw.FormDataContentType())
  return req, nil
}
