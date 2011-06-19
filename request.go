// Copyright 2011 Muthukannan T <manki@manki.in>. All Rights Reserved.

package flickgo

import (
  "crypto/md5"
  "fmt"
  "http"
  "json"
  "log"
  "os"
  "regexp"
  "sort"
  "strings"
)

const (
  service = "http://www.flickr.com/services"
)

func keys(m map[string]string) sort.StringArray {
  ks := make([]string, len(m))
  i := 0
  for k, _ := range m {
    ks[i] = k
    i++
  }
  return ks
}

func sign(secret string, apiKey string, path string, args map[string]string) string {
  args["api_key"] = apiKey
  ks := keys(args)
  ks.Sort()
  parts := make([]string, len(ks) + 1)
  m := md5.New()
  m.Write([]byte(secret))
  for i, k := range ks {
    value := http.URLEscape(args[k])
    parts[i] = fmt.Sprintf("%s=%s", k, value)
    m.Write([]byte(k + value))
  }
  parts[len(ks)] = fmt.Sprintf("api_sig=%x", m.Sum())
  return fmt.Sprintf("%s/%s/?", service, path) + strings.Join(parts, "&")
}

func url(c *Client, method string, args map[string]string) string {
  args["method"] = method
  args["format"] = "json"
  if len(c.AuthToken) > 0 {
    args["auth_token"] = c.AuthToken
  }
  return sign(c.secret, c.apiKey, "rest", args)
}

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

// Sends a GET request to u and returns the response bytes.
func (c *Client) fetch(u string) ([]byte, os.Error) {
  r, _, getErr := c.httpClient.Get(u)
  if getErr != nil {
    return nil, getErr
  }
  defer r.Body.Close()

  // TODO: handle error response codes like 401 and 500.

  buf, readErr := c.readFn(r.Body)
  if readErr != nil {
    return nil, readErr
  }
  return extractJSON(buf), nil
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
  data, err := c.fetch(url)
  if err != nil {
    return err
  }
  log.Printf("JSON received: %s", string(data))

  // Try to parse the response as error.  Both success and failure responses
  // have a 'stat' field; if stat != "ok" the request was successful.
  r := flickrError{}
  if err = json.Unmarshal(data, &r); err != nil {
    return os.NewError(err.String() + "; JSON=" + string(data))
  }
  if r.Stat != "ok" {
    return os.NewError(fmt.Sprintf("Flickr error code %d: %s",
                                   r.Code, r.Message))
  }

  if err = json.Unmarshal(data, resp); err != nil {
    return os.NewError(err.String() + "; JSON=" + string(data))
  }
  return nil
}
