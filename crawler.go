package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Crawler struct {
	shorteners []*Shortener
	stop       chan struct{}
	wg         sync.WaitGroup
	urls       chan string
}

func NewCrawler(concurrency int) *Crawler {
	return &Crawler{
		stop: make(chan struct{}),
		urls: make(chan string, concurrency),
	}
}

func (c *Crawler) Start() {
	c.wg.Add(len(c.shorteners))
	c.wg.Add(len(c.urls))

	for _, s := range c.shorteners {
		go c.crawl(s)
	}

	for i := 0; i < cap(c.urls); i++ {
		go c.worker()
	}
}

func (c *Crawler) Stop() {
	close(c.stop)
	c.wg.Wait()
}

func (c *Crawler) crawl(s *Shortener) {
	// fmt.Printf("Started %s crawler\n", s.BaseURL)
	urlsChan := c.getUrlsChan(s)

	for url := range urlsChan {
		c.urls <- url
	}

	c.wg.Done()
}

func (c *Crawler) worker() {
	// fmt.Println("Started worker")
	proxyUrl, _ := url.Parse("socks5://tor:5566")
	httpTransport := &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
		Proxy:               http.ProxyURL(proxyUrl),
	}

	httpClient := &http.Client{
		Timeout:   time.Second * 10,
		Transport: httpTransport,
	}

	for {
		select {
		case <-c.stop:
			c.wg.Done()
			return
		case url := <-c.urls:
			resp, err := httpClient.Get(url)

			if err == nil && resp.StatusCode == 200 {
				defer resp.Body.Close()
				doc, err := goquery.NewDocumentFromReader(resp.Body)

				if err != nil {
					log.Println(err)
					continue
				}

				title := doc.Find("title").Text()
				strings.ReplaceAll(title, "\n", "")

				fmt.Printf("%-30s %s\n", url, title)
			}
		}
	}
}

func (c *Crawler) getUrlsChan(s *Shortener) <-chan string {
	urlsChan := make(chan string)
	go c.startGenerator(s, urlsChan)
	return urlsChan
}

func (c *Crawler) startGenerator(s *Shortener, urlsChan chan string) {
	charsetSize := len(s.Charset)
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	var length int
	if s.MinLength != s.MaxLength {
		length = rnd.Intn(s.MaxLength-s.MinLength) + s.MinLength
	} else {
		length = s.MinLength
	}

	for {
		select {
		case <-c.stop:
			close(urlsChan)
			return
		default:
			var sb strings.Builder
			sb.WriteString(s.BaseURL)

			for i := 0; i < length; i++ {
				sb.WriteString(string(s.Charset[rnd.Intn(charsetSize)]))
			}

			urlsChan <- sb.String()
		}
	}
}

func (c *Crawler) AddShortener(s *Shortener) {
	c.shorteners = append(c.shorteners, s)
}
