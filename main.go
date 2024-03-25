package main

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/anthdm/hollywood/actor"
	"golang.org/x/net/html"
)

type VisitRequest struct {
	links []string
}

type Manager struct {
}

func NewManager() actor.Producer {
	return func() actor.Receiver {
		return &Manager{}
	}
}

func (m *Manager) Receive(c *actor.Context) {
	switch msg := c.Message().(type) {
	case VisitRequest:
		m.handleVisitRequest(msg)
	case actor.Started:
		slog.Info("Manager started")
	case actor.Stopped:
	}
}

func (m *Manager) handleVisitRequest(msg VisitRequest) error {
	for _, link := range msg.links {
		slog.Info("visiting url", "url", link)
	}
	return nil
}

func main() {
	baseURL, err := url.Parse("https://nation.africa/kenya")
	if err != nil {
		log.Fatal(err)
	}
	resp, err := http.Get(baseURL.String())
	if err != nil {
		log.Fatal(err)
	}

	links := extractLinks(resp.Body)
	for _, link := range links {
		lurl, err := url.Parse(link)
		if err != nil {
			log.Fatal(err)
		}
		actualLink := baseURL.ResolveReference(lurl)
		fmt.Println(actualLink)
	}

	fmt.Println(extractLinks(resp.Body))
	return
	e, err := actor.NewEngine(actor.NewEngineConfig())
	if err != nil {
		log.Fatal(err)
	}
	pid := e.Spawn(NewManager(), "manager")

	time.Sleep(time.Millisecond * 500)

	e.Send(pid, VisitRequest{links: []string{"https://nation.africa/kenya"}})

	time.Sleep(time.Second * 10)
}

func extractLinks(body io.Reader) []string {
	links := make([]string, 0)
	tokenizer := html.NewTokenizer(body)

	for {
		tokenType := tokenizer.Next()
		if tokenType == html.ErrorToken {
			return links
		}

		if tokenType == html.StartTagToken {
			token := tokenizer.Token()
			if token.Data == "a" {
				for _, attr := range token.Attr {
					if attr.Key == "href" {
						links = append(links, attr.Val)
					}
				}
			}
		}
	}
	return links
}
