package feeds

// rss support
// validation done according to spec here:
//    http://cyber.law.harvard.edu/rss/rss.html

import (
	"encoding/xml"
	"fmt"
	"strings"
	"time"
)

// private wrapper around the RssFeed which gives us the <rss>..</rss> xml
type RssFeedXml struct {
	XMLName          xml.Name `xml:"rss"`
	Version          string   `xml:"version,attr"`
	ContentNamespace string   `xml:"xmlns:content,attr"`
	Channel          *RssFeed
}

type RssContent struct {
	XMLName xml.Name `xml:"content:encoded"`
	Content string   `xml:",cdata"`
}

type RssTitle interface {
	SetTitle(string) RssTitle
}

type RssTitleString struct {
	XMLName xml.Name `xml:"title"`
	Title   string   `xml:",cdata"`
}

func (rst RssTitleString) SetTitle(title string) RssTitle {
	return RssTitleString{Title: title}
}

type RssTitleStringEncoded struct {
	XMLName xml.Name `xml:"title:encoded"`
	Title   string   `xml:",cdata"`
}

func (rst RssTitleStringEncoded) SetTitle(title string) RssTitle {
	return RssTitleStringEncoded{Title: title}
}

type RssDescription interface {
	SetDesc(string) RssDescription
}

type RssDescriptionString struct {
	XMLName     xml.Name `xml:"description"`
	Description string   `xml:"description"`
}

func (rsd *RssDescriptionString) SetDesc(desk string) RssDescription {
	return &RssDescriptionString{Description: desk}
}

type RssDescriptionStringEncoded struct {
	XMLName     xml.Name `xml:"description:encoded"`
	Description string   `xml:",cdata"`
}

func (rsd RssDescriptionStringEncoded) SetDesc(desk string) RssDescription {
	return &RssDescriptionStringEncoded{Description: desk}
}

type RssImage struct {
	XMLName xml.Name `xml:"image"`
	Url     string   `xml:"url"`
	Title   string   `xml:"title"`
	Link    string   `xml:"link"`
	Width   int      `xml:"width,omitempty"`
	Height  int      `xml:"height,omitempty"`
}

type RssTextInput struct {
	XMLName     xml.Name `xml:"textInput"`
	Title       string   `xml:"title"`
	Description string   `xml:"description"`
	Name        string   `xml:"name"`
	Link        string   `xml:"link"`
}

type Id struct {
	XMLName     xml.Name `xml:"guid,omitempty"`
	IsPermaLink bool     `xml:"isPermaLink,attr,omitempty"`
	Guid        string   `xml:",chardata"`
}

type RssMediaContent struct {
	XMLName xml.Name `xml:"media:content"`
	Url     string   `xml:"url,attr"`
	Length  string   `xml:"length,attr,omitempty"`
	Type    string   `xml:"type,attr"`
	Medium  string   `xml:"medium,attr"`
}

type RssMediaTitle struct {
	XMLName xml.Name `xml:"media:title,omitempty"`
	Type    string   `xml:"type,attr,omitempty"`
	Title   string   `xml:",chardata"`
}
type RssFeed struct {
	XMLName        xml.Name `xml:"channel"`
	Title          string   `xml:"title"`       // required
	Link           string   `xml:"link"`        // required
	Description    string   `xml:"description"` // required
	Language       string   `xml:"language,omitempty"`
	Copyright      string   `xml:"copyright,omitempty"`
	ManagingEditor string   `xml:"managingEditor,omitempty"` // Author used
	WebMaster      string   `xml:"webMaster,omitempty"`
	PubDate        string   `xml:"pubDate,omitempty"`       // created or updated
	LastBuildDate  string   `xml:"lastBuildDate,omitempty"` // updated used
	Category       string   `xml:"category,omitempty"`
	Generator      string   `xml:"generator,omitempty"`
	Docs           string   `xml:"docs,omitempty"`
	Cloud          string   `xml:"cloud,omitempty"`
	Ttl            int      `xml:"ttl,omitempty"`
	Rating         string   `xml:"rating,omitempty"`
	SkipHours      string   `xml:"skipHours,omitempty"`
	SkipDays       string   `xml:"skipDays,omitempty"`
	Image          *RssImage
	TextInput      *RssTextInput
	Items          []*RssItem `xml:"item"`
}

type RssItem struct {
	XMLName       xml.Name       `xml:"item"`
	Title         RssTitle       // required
	Link          string         `xml:"link"` // required
	Description   RssDescription // required
	Content       *RssContent
	Author        string `xml:"author,omitempty"`
	Category      string `xml:"category,omitempty"`
	Comments      string `xml:"comments,omitempty"`
	MediaContent  *RssMediaContent
	MediaKeywords string `xml:"media:keywords,omitempty"`
	MediaTitle    *RssMediaTitle
	Guid          *Id
	PubDate       string `xml:"pubDate,omitempty"` // created or updated
	Source        string `xml:"source,omitempty"`
}

type RssGuid struct {
	//RSS 2.0 <guid isPermaLink="true">http://inessential.com/2002/09/01.php#a2</guid>
	XMLName     xml.Name `xml:"guid"`
	Id          string   `xml:",chardata"`
	IsPermaLink string   `xml:"isPermaLink,attr,omitempty"` // "true", "false", or an empty string
}

type Rss struct {
	*Feed
}

// create a new RssItem with a generic Item struct's data
func newRssItem(i *Item) *RssItem {
	item := &RssItem{
		PubDate:  anyTimeFormat(time.RFC1123Z, i.Created, i.Updated),
		Category: i.Category,
	}
	if i.Id != "" {
		item.Guid = &Id{Guid: i.Id, IsPermaLink: true}
	}

	item.Title.SetTitle(i.Title)
	item.Description.SetDesc(i.Description)
	if i.Link != nil {
		item.Link = i.Link.Href
	}
	if len(i.Content) > 0 {
		item.Content = &RssContent{Content: i.Content}
	}
	if i.Source != nil {
		item.Source = i.Source.Href
	}

	// Define a closure
	if i.MediaContent != nil && i.MediaContent.Type != "" && i.MediaContent.Length != "" {
		item.MediaContent = &RssMediaContent{Url: i.MediaContent.Url, Type: i.MediaContent.Type, Medium: strings.Split(i.MediaContent.Type, "/")[0]}
		if i.MediaTitle != "" {
			item.MediaTitle = &RssMediaTitle{Title: i.MediaTitle, Type: "plain"}
		}
		if i.MediaKeywords != "" {
			item.MediaKeywords = i.MediaKeywords
		}
	}

	if i.Author != nil {
		item.Author = i.Author.Name
	}
	return item
}

// create a new RssFeed with a generic Feed struct's data
func (r *Rss) RssFeed() *RssFeed {
	pub := anyTimeFormat(time.RFC1123Z, r.Created, r.Updated)
	build := anyTimeFormat(time.RFC1123Z, r.Updated)
	author := ""
	if r.Author != nil {
		author = r.Author.Email
		if len(r.Author.Name) > 0 {
			author = fmt.Sprintf("%s (%s)", r.Author.Email, r.Author.Name)
		}
	}

	var image *RssImage
	if r.Image != nil {
		image = &RssImage{Url: r.Image.Url, Title: r.Image.Title, Link: r.Image.Link, Width: r.Image.Width, Height: r.Image.Height}
	}

	var href string
	if r.Link != nil {
		href = r.Link.Href
	}
	channel := &RssFeed{
		Title:          r.Title,
		Link:           href,
		Description:    r.Description,
		ManagingEditor: author,
		PubDate:        pub,
		LastBuildDate:  build,
		Copyright:      r.Copyright,
		Image:          image,
	}
	for _, i := range r.Items {
		channel.Items = append(channel.Items, newRssItem(i))
	}
	return channel
}

// FeedXml returns an XML-Ready object for an Rss object
func (r *Rss) FeedXml() interface{} {
	// only generate version 2.0 feeds for now
	return r.RssFeed().FeedXml()

}

// FeedXml returns an XML-ready object for an RssFeed object
func (r *RssFeed) FeedXml() interface{} {
	return &RssFeedXml{
		Version:          "2.0",
		Channel:          r,
		ContentNamespace: "http://purl.org/rss/1.0/modules/content/",
	}
}
