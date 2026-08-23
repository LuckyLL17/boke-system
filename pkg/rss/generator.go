package rss

import (
	"encoding/xml"
	"fmt"
	"time"

	"podcast-platform/internal/domain"
	"podcast-platform/pkg/utils"
)

type ItunesCategory struct {
	XMLName     xml.Name        `xml:"itunes:category"`
	Text        string          `xml:"text,attr"`
	SubCategory *ItunesCategory `xml:"itunes:category,omitempty"`
}

type ItunesOwner struct {
	XMLName xml.Name `xml:"itunes:owner"`
	Name    string   `xml:"itunes:name"`
	Email   string   `xml:"itunes:email"`
}

type ItunesImage struct {
	XMLName xml.Name `xml:"itunes:image"`
	Href    string   `xml:"href,attr"`
}

type RSSEpisodeEnclosure struct {
	XMLName xml.Name `xml:"enclosure"`
	URL     string   `xml:"url,attr"`
	Length  int64    `xml:"length,attr"`
	Type    string   `xml:"type,attr"`
}

type RSSEpisode struct {
	XMLName     xml.Name            `xml:"item"`
	Title       string              `xml:"title"`
	Link        string              `xml:"link,omitempty"`
	PubDate     string              `xml:"pubDate"`
	GUID        string              `xml:"guid"`
	Description string              `xml:"description"`
	Author      string              `xml:"itunes:author,omitempty"`
	Subtitle    string              `xml:"itunes:subtitle,omitempty"`
	Summary     string              `xml:"itunes:summary,omitempty"`
	Duration    string              `xml:"itunes:duration,omitempty"`
	Explicit    string              `xml:"itunes:explicit"`
	Image       ItunesImage         `xml:"itunes:image,omitempty"`
	Enclosure   RSSEpisodeEnclosure `xml:"enclosure"`
	Episode     string              `xml:"itunes:episode,omitempty"`
	Season      string              `xml:"itunes:season,omitempty"`
	EpisodeType string              `xml:"itunes:episodeType,omitempty"`
	Chapters    []RSSChapter        `xml:"psc:chapters>psc:chapter,omitempty"`
}

type RSSChapter struct {
	XMLName xml.Name `xml:"psc:chapter"`
	Start   string   `xml:"start,attr"`
	Title   string   `xml:"title,attr,omitempty"`
	Href    string   `xml:"href,attr,omitempty"`
	Image   string   `xml:"image,attr,omitempty"`
}

type RSSChannel struct {
	XMLName        xml.Name         `xml:"channel"`
	Title          string           `xml:"title"`
	Link           string           `xml:"link"`
	Description    string           `xml:"description"`
	Language       string           `xml:"language"`
	Copyright      string           `xml:"copyright,omitempty"`
	LastBuildDate  string           `xml:"lastBuildDate"`
	PubDate        string           `xml:"pubDate"`
	Generator      string           `xml:"generator"`
	WebMaster      string           `xml:"webMaster,omitempty"`
	ManagingEditor string           `xml:"managingEditor,omitempty"`
	Image          RSSImage         `xml:"image,omitempty"`
	ItunesAuthor   string           `xml:"itunes:author,omitempty"`
	ItunesSubtitle string           `xml:"itunes:subtitle,omitempty"`
	ItunesSummary  string           `xml:"itunes:summary,omitempty"`
	ItunesOwner    ItunesOwner      `xml:"itunes:owner,omitempty"`
	ItunesImage    ItunesImage      `xml:"itunes:image,omitempty"`
	ItunesExplicit string           `xml:"itunes:explicit"`
	ItunesCategory []ItunesCategory `xml:"itunes:category,omitempty"`
	ItunesType     string           `xml:"itunes:type,omitempty"`
	ItunesNewFeed  string           `xml:"itunes:new-feed-url,omitempty"`
	Items          []RSSEpisode     `xml:"item"`
}

type RSSImage struct {
	XMLName xml.Name `xml:"image"`
	URL     string   `xml:"url"`
	Title   string   `xml:"title"`
	Link    string   `xml:"link"`
}

type RSSFeed struct {
	XMLName   xml.Name   `xml:"rss"`
	Version   string     `xml:"version,attr"`
	ItunesNS  string     `xml:"xmlns:itunes,attr"`
	ContentNS string     `xml:"xmlns:content,attr"`
	AtomNS    string     `xml:"xmlns:atom,attr"`
	PscNS     string     `xml:"xmlns:psc,attr"`
	Channel   RSSChannel `xml:"channel"`
}

type Generator struct {
	BaseURL string
}

func NewGenerator(baseURL string) *Generator {
	return &Generator{BaseURL: baseURL}
}

func (g *Generator) GenerateFeed(channel *domain.Channel, episodes []domain.Episode, baseSiteURL string) string {
	feed := RSSFeed{
		Version:   "2.0",
		ItunesNS:  "http://www.itunes.com/dtds/podcast-1.0.dtd",
		ContentNS: "http://purl.org/rss/1.0/modules/content/",
		AtomNS:    "http://www.w3.org/2005/Atom",
		PscNS:     "http://podlove.org/simple-chapters",
	}

	rssChannel := RSSChannel{
		Title:         utils.EscapeXML(channel.Title),
		Link:          baseSiteURL,
		Description:   utils.EscapeXML(channel.Description),
		Language:      channel.Language,
		Copyright:     utils.EscapeXML(channel.Copyright),
		LastBuildDate: utils.FormatRFC822(time.Now()),
		PubDate:       utils.FormatRFC822(time.Now()),
		Generator:     "Podcast Platform v1.0",
		ItunesAuthor:  utils.EscapeXML(channel.Author),
		ItunesSummary: utils.EscapeXML(channel.Description),
		ItunesOwner: ItunesOwner{
			Name:  utils.EscapeXML(channel.Author),
			Email: utils.EscapeXML(channel.Email),
		},
		ItunesExplicit: boolToYesNo(channel.Explicit),
		ItunesType:     "episodic",
	}

	if channel.CoverImageURL != "" {
		coverURL := resolveURL(g.BaseURL, channel.CoverImageURL)
		rssChannel.Image = RSSImage{
			URL:   coverURL,
			Title: rssChannel.Title,
			Link:  rssChannel.Link,
		}
		rssChannel.ItunesImage = ItunesImage{Href: coverURL}
	}

	if channel.ITunesCategory != "" {
		rssChannel.ItunesCategory = []ItunesCategory{
			{Text: channel.ITunesCategory},
		}
	}

	if channel.Email != "" {
		rssChannel.ManagingEditor = fmt.Sprintf("%s (%s)", utils.EscapeXML(channel.Email), utils.EscapeXML(channel.Author))
		rssChannel.WebMaster = rssChannel.ManagingEditor
	}

	rssChannel.Items = make([]RSSEpisode, 0, len(episodes))
	for i := range episodes {
		ep := &episodes[i]
		item := g.generateEpisode(ep, channel, g.BaseURL)
		rssChannel.Items = append(rssChannel.Items, item)
	}

	feed.Channel = rssChannel

	result, _ := xml.MarshalIndent(feed, "", "  ")
	return utils.XMLHeader() + string(result)
}

func (g *Generator) generateEpisode(ep *domain.Episode, channel *domain.Channel, siteURL string) RSSEpisode {
	item := RSSEpisode{
		Title:       utils.EscapeXML(ep.Title),
		Link:        fmt.Sprintf("%s/episode/%d", siteURL, ep.ID),
		GUID:        fmt.Sprintf("%s/episode/%d", siteURL, ep.ID),
		Description: utils.EscapeXML(ep.Description),
		Explicit:    boolToYesNo(ep.Explicit),
		EpisodeType: "full",
	}

	pubDate := ep.PublishedAt
	if pubDate == nil {
		now := time.Now()
		pubDate = &now
	}
	item.PubDate = utils.FormatRFC822(*pubDate)

	if ep.Author != "" {
		item.Author = utils.EscapeXML(ep.Author)
	} else {
		item.Author = utils.EscapeXML(channel.Author)
	}

	if ep.Description != "" {
		item.Subtitle = utils.EscapeXML(truncate(ep.Description, 255))
		item.Summary = utils.EscapeXML(ep.Description)
	}

	if ep.Duration > 0 {
		item.Duration = formatDuration(ep.Duration)
	}

	if ep.CoverImageURL != "" {
		item.Image = ItunesImage{Href: resolveURL(g.BaseURL, ep.CoverImageURL)}
	} else if channel.CoverImageURL != "" {
		item.Image = ItunesImage{Href: resolveURL(g.BaseURL, channel.CoverImageURL)}
	}

	item.Enclosure = RSSEpisodeEnclosure{
		URL:    resolveURL(g.BaseURL, ep.AudioFileURL),
		Length: ep.AudioFileSize,
		Type:   ep.MimeType,
	}

	if ep.EpisodeNumber > 0 {
		item.Episode = fmt.Sprintf("%d", ep.EpisodeNumber)
	}
	if ep.SeasonNumber > 0 {
		item.Season = fmt.Sprintf("%d", ep.SeasonNumber)
	}

	if len(ep.Chapters) > 0 {
		item.Chapters = make([]RSSChapter, 0, len(ep.Chapters))
		for j := range ep.Chapters {
			ch := &ep.Chapters[j]
			c := RSSChapter{
				Start: formatChapterTime(ch.StartTime),
				Title: utils.EscapeXML(ch.Title),
			}
			if ch.URL != "" {
				c.Href = ch.URL
			}
			if ch.ImageURL != "" {
				c.Image = ch.ImageURL
			}
			item.Chapters = append(item.Chapters, c)
		}
	}

	return item
}

func resolveURL(base, path string) string {
	if len(path) == 0 {
		return base
	}
	if len(path) >= 4 && (path[:4] == "http" || path[:4] == "HTTP") {
		return path
	}
	if base != "" && base[len(base)-1] == '/' && path[0] == '/' {
		return base[:len(base)-1] + path
	}
	return base + path
}

func boolToYesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}

func formatDuration(seconds int) string {
	if seconds <= 0 {
		return "0:00"
	}
	h := seconds / 3600
	m := (seconds % 3600) / 60
	s := seconds % 60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%d:%02d", m, s)
}

func formatChapterTime(seconds int) string {
	h := seconds / 3600
	m := (seconds % 3600) / 60
	s := seconds % 60
	ms := 0
	return fmt.Sprintf("%02d:%02d:%02d.%03d", h, m, s, ms)
}

func ValidateFeed(channel *domain.Channel, episodes []domain.Episode) []string {
	var issues []string
	if channel.Title == "" {
		issues = append(issues, "channel title is required")
	}
	if channel.Description == "" {
		issues = append(issues, "channel description is required")
	}
	if len(episodes) == 0 {
		issues = append(issues, "no episodes published")
	}
	for i := range episodes {
		ep := &episodes[i]
		if ep.Title == "" {
			issues = append(issues, fmt.Sprintf("episode %d: title required", i+1))
		}
		if ep.AudioFileURL == "" {
			issues = append(issues, fmt.Sprintf("episode %s: audio file required", ep.Title))
		}
		if ep.MimeType == "" {
			issues = append(issues, fmt.Sprintf("episode %s: mime type missing", ep.Title))
		}
	}
	return issues
}
