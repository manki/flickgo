// Copyright 2011 Muthukannan T <manki@manki.in>. All Rights Reserved.

package flickgo

import (
  "crypto/md5"
  "fmt"
  "http"
  "os"
  "regexp"
  "sort"
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

func sign(secret string, apiKey string, path string, args map[string]string) *http.URL {
  args["api_key"] = apiKey
  ks := keys(args)
  ks.Sort()
  s := fmt.Sprintf("%s/%s/?", service, path)
  m := md5.New()
  m.Write([]byte(secret))
  for _, k := range ks {
    value := http.URLEscape(args[k])
    s += fmt.Sprintf("%s=%s&", k, value)
    m.Write([]byte(k + value))
  }
  s += fmt.Sprintf("api_sig=%x", m.Sum())
  u, err := http.ParseURL(s)
  if err != nil {
    panic("URL parsing failed")
  }
  return u
}

func (c *Client) url(method string, args map[string]string) *http.URL {
  a := make(map[string]string)
  for k, v := range args {
    a[k] = v
  }
  a["method"] = method
  a["format"] = "json"
  return sign(c.secret, c.apiKey, "rest", a)
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
func (c *Client) fetch(u *http.URL) ([]byte, os.Error) {
  r, _, getErr := c.httpClient.Get(u.String())
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
