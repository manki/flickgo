// Copyright 2011 Muthukannan T <manki@manki.in>. All Rights Reserved.

package flickgo

import (
  "fmt"
  "json"
  "log"
  "os"
)

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
