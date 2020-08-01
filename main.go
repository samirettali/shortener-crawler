package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

type Shortener struct {
	BaseURL   string
	Charset   string
	MinLength int
	MaxLength int
}

func main() {
	inventShortener := &Shortener{
		BaseURL:   "https://invent.ge/",
		Charset:   "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
		MinLength: 7,
		MaxLength: 7,
	}

	// bitdoShortener := &Shortener{
	// 	BaseURL:   "https://bit.do/",
	// 	Charset:   "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
	// 	MinLength: 3,
	// 	MaxLength: 8,
	// }

	bitlyShortener := &Shortener{
		BaseURL:   "https://bit.ly/",
		Charset:   "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
		MinLength: 7,
		MaxLength: 7,
	}

	crawler := NewCrawler(40)

	crawler.AddShortener(inventShortener)
	// crawler.AddShortener(bitdoShortener)
	crawler.AddShortener(bitlyShortener)

	crawler.Start()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop
	fmt.Println("Stopping crawler...")
	crawler.Stop()
}
