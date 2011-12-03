// Copyright 2011 Muthukannan T <manki@manki.in>. All Rights Reserved.

package flickgo

import (
	"fmt"
)

// Image sizes supported by Flickr.  See
// http://www.flickr.com/services/api/misc.urls.html for more information.
const (
	SizeSmallSquare = "s"
	SizeThumbnail   = "t"
	SizeSmall       = "m"
	SizeMedium500   = "-"
	SizeMedium640   = "z"
	SizeLarge       = "b"
	SizeOriginal    = "o"
)

// Response for photo search requests.
type SearchResponse struct {
	Page    string  `xml:"attr"`
	Pages   string  `xml:"attr"`
	PerPage string  `xml:"attr"`
	Total   string  `xml:"attr"`
	Photos  []Photo `xml:"photo>"`
}

// A Flickr user.
type User struct {
	UserName string `xml:"attr"`
	NSID     string `xml:"attr"`
}

// Represents a Flickr photo.
type Photo struct {
	ID       string `xml:"attr"`
	Owner    string `xml:"attr"`
	Secret   string `xml:"attr"`
	Server   string `xml:"attr"`
	Farm     string `xml:"attr"`
	Title    string `xml:"attr"`
	IsPublic string `xml:"attr"`
	Width_T  string `xml:"attr"`
	Height_T string `xml:"attr"`
	// Photo's aspect ratio: width divided by height.
	Ratio float64
}

// Returns the URL to this photo in the specified size.
func (p *Photo) URL(size string) string {
	if size == "-" {
		return fmt.Sprintf("http://farm%s.static.flickr.com/%s/%s_%s.jpg",
			p.Farm, p.Server, p.ID, p.Secret)
	}
	return fmt.Sprintf("http://farm%s.static.flickr.com/%s/%s_%s_%s.jpg",
		p.Farm, p.Server, p.ID, p.Secret, size)
}

type PhotoSet struct {
	ID          string `xml:"attr"`
	Title       string
	Description string
}
