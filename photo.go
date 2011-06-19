// Copyright 2011 Muthukannan T <manki@manki.in>. All Rights Reserved.

package flickgo

import (
  "fmt"
)

// Image sizes supported by Flickr.  See
// http://www.flickr.com/services/api/misc.urls.html for more information.
const (
  SizeSmallSquare = "s"
  SizeThumbnail = "t"
  SizeSmall = "m"
  SizeMedium500 = "-"
  SizeMedium640 = "z"
  SizeLarge = "b"
  SizeOriginal = "o"
)

// Response for photo search requests.
type SearchResponse struct {
  Page int
  Pages int
  PerPage int
  Total string
  Photos []Photo "photo"
}

// Represents a Flickr photo.
type Photo struct {
  Id string
  Owner string
  Secret string
  Server string
  Farm int
  Title string
}

// Returns the URL to this photo in the specified size.
func (p *Photo) URL(size string) string {
  return fmt.Sprintf("http://farm%d.static.flickr.com/%s/%s_%s_%s.jpg",
                     p.Farm, p.Server, p.Id, p.Secret, size)
}
