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
	Page    string  `xml:"page,attr"`
	Pages   string  `xml:"pages,attr"`
	PerPage string  `xml:"perpage,attr"`
	Total   string  `xml:"total,attr"`
	Photos  []Photo `xml:"photo"`
}

// A Flickr user.
type User struct {
	UserName string `xml:"username,attr"`
	NSID     string `xml:"nsid,attr"`
}

// Represents a Flickr photo.
type Photo struct {
	ID       string `xml:"id,attr"`
	Owner    string `xml:"owner,attr"`
	Secret   string `xml:"secret,attr"`
	Server   string `xml:"server,attr"`
	Farm     string `xml:"farm,attr"`
	Title    string `xml:"title,attr"`
	IsPublic string `xml:"ispublic,attr"`
	Width_T  string `xml:"width_t,attr"`
	Height_T string `xml:"height_t,attr"`
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
	ID          string `xml:"id,attr"`
	Title       string `xml:"title"`
	Description string `xml:"description"`
}

type GetLocationResponse struct {
	Photo    string   `xml:"id,attr"`
	Location Location `xml:"location"`
}

type Location struct {
	Latitude  string `xml:"latitude,attr"`
	Longitude string `xml:"longitude,attr"`
	Accuracy  string `xml:"accuracy,attr"`
	Context   string `xml:"context,attr"`
	PlaceID   string `xml:"place_id,attr"`
	Woeid     string `xml:"woeid,attr"`
}

type GetPersonResponse struct {
	ID         string `xml:"id,attr"`
	Nsid       string `xml:"nsid,attr"`
	Ispro      string `xml:"ispro,attr"`
	Iconserver string `xml:"iconserver,attr"`
	Iconfarm   string `xml:"iconfarm,attr"`
	Path_alias string `xml:"path_alias,attr"`
	Gender     string `xml:"gender,attr"`
	Ignored    string `xml:"ignored,attr"`
	Contact    string `xml:"contact,attr"`
	Friend     string `xml:"friend,attr"`
	Family     string `xml:"family,attr"`
	Revcontact string `xml:"revcontact,attr"`
	Revfriend  string `xml:"revfriend,attr"`
	Revfamily  string `xml:"revfamily,attr"`
	Username   string `xml:"username"`
}
