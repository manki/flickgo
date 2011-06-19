// Copyright 2011 Muthukannan T <manki@manki.in>. All Rights Reserved.

package flickgo

import (
  "crypto/md5"
  "fmt"
  "http"
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
