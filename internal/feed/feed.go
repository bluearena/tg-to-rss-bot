package feed

import (
	"encoding/xml"
	"io"
	"os"
	"sync"
	"time"

	"github.com/ShoshinNikita/log"
	"github.com/gorilla/feeds"
	"github.com/pkg/errors"

	"github.com/ShoshinNikita/tg-to-rss-bot/internal/params"
)

// init create params.RssFolder
func init() {
	err := os.MkdirAll(params.RssFolder, 0666)
	if err != nil {
		log.Fatal("can't create folder %s: %s", params.RssFolder, err)
	}
}

type Feed struct {
	feed  *feeds.RssFeed
	mutex *sync.RWMutex
}

func NewFeed() *Feed {
	return &Feed{
		feed:  new(feeds.RssFeed),
		mutex: new(sync.RWMutex),
	}
}

func (feed *Feed) Init() error {
	f, err := os.Open(params.RssFile)
	if err != nil {
		// Need to create a file
		f, err = os.Create(params.RssFile)
		if err != nil {
			return errors.Wrapf(err, "can't create a new file %s", params.RssFile)
		}
		defer f.Close()

		feed.feed = &feeds.RssFeed{
			Title:       "tg-to-rss-bot",
			Link:        "https://github.com/ShoshinNikita/tg-to-rss-bot",
			Description: "This feed is generated by https://github.com/ShoshinNikita/tg-to-rss-bot",
		}

		err = xml.NewEncoder(f).Encode(feed.feed)
		if err != nil {
			return errors.Wrap(err, "can't write RSS feed into file")
		}

		return nil
	}

	defer f.Close()
	err = xml.NewDecoder(f).Decode(feed.feed)
	if err != nil {
		return errors.Wrap(err, "can't decode RSS feed from a file")
	}

	return nil
}

func (feed *Feed) Add(author, title, description, filepath string, created time.Time) error {
	feed.mutex.Lock()
	defer feed.mutex.Unlock()

	item := &feeds.RssItem{
		Title:       title,
		Link:        params.Host + "/" + filepath,
		Description: description,
		Author:      author,
		PubDate:     created.Format(time.RFC822),
	}

	feed.feed.Items = append(feed.feed.Items, item)

	// Write into disk
	f, err := os.OpenFile(params.RssFile, os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		return errors.Wrapf(err, "can't open file %s", params.RssFile)
	}
	defer f.Close()

	err = xml.NewEncoder(f).Encode(feed.feed)
	if err != nil {
		return errors.Wrap(err, "can't write RSS feed into file")
	}

	return nil
}

func (feed *Feed) Write(w io.Writer) error {
	feed.mutex.RLock()
	defer feed.mutex.RUnlock()

	err := feeds.WriteXML(feed.feed, w)
	if err != nil {
		errors.Wrap(err, "can't write RSS feed into io.Writer")
	}

	return nil
}
