package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go-lang-jwt/helpers"
	"go-lang-jwt/models"

	"github.com/PuerkitoBio/goquery"
)

// ArticleData represents extracted article information
type ArticleData struct {
	Title       string
	URL         string
	Summary     string
	PublishedAt *time.Time
	Author      string
	ContentHash string
}

// ExtractArticles fetches URL and extracts articles
func ExtractArticles(ctx context.Context, source models.Source) ([]ArticleData, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", source.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; FeedAggregator/1.0)")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %v", err)
	}

	var articles []ArticleData

	// Strategy 1: Hacker News specific
	if strings.Contains(source.URL, "news.ycombinator.com") {
		articles = extractHackerNews(doc, source)
	}

	// Strategy 2: Lobsters specific
	if len(articles) == 0 && strings.Contains(source.URL, "lobste.rs") {
		articles = extractLobsters(doc, source)
	}

	// Strategy 3: Generic <article> tags
	if len(articles) == 0 {
		doc.Find("article").Each(func(i int, s *goquery.Selection) {
			if article := extractFromSelection(s, source); article != nil {
				articles = append(articles, *article)
			}
		})
	}

	// Strategy 4: Common class names
	if len(articles) == 0 {
		doc.Find("div.post, div.entry, div.article-item, div.story, li.story").Each(func(i int, s *goquery.Selection) {
			if article := extractFromSelection(s, source); article != nil {
				articles = append(articles, *article)
			}
		})
	}

	// Strategy 5: Any link with heading
	if len(articles) == 0 {
		doc.Find("h1 a, h2 a, h3 a").Each(func(i int, s *goquery.Selection) {
			title := strings.TrimSpace(s.Text())
			url, exists := s.Attr("href")

			if !exists || title == "" || len(title) < 10 {
				return
			}

			if strings.HasPrefix(url, "/") {
				url = source.URL + url
			}

			articles = append(articles, ArticleData{
				Title:       title,
				URL:         url,
				Summary:     "",
				ContentHash: helpers.GenerateContentHash(title, ""),
			})
		})
	}

	if len(articles) == 0 {
		return nil, errors.New("no articles found on page")
	}

	// Limit to 50 articles per crawl
	if len(articles) > 50 {
		articles = articles[:50]
	}

	return articles, nil
}

func extractHackerNews(doc *goquery.Document, source models.Source) []ArticleData {
	var articles []ArticleData

	doc.Find("tr.athing").Each(func(i int, s *goquery.Selection) {
		titleLink := s.Find("span.titleline > a").First()
		title := strings.TrimSpace(titleLink.Text())
		url, exists := titleLink.Attr("href")

		if !exists || title == "" {
			return
		}

		// Make URL absolute
		if strings.HasPrefix(url, "item?id=") {
			url = "https://news.ycombinator.com/" + url
		}

		articles = append(articles, ArticleData{
			Title:       title,
			URL:         url,
			Summary:     "",
			ContentHash: helpers.GenerateContentHash(title, ""),
		})
	})

	return articles
}

func extractLobsters(doc *goquery.Document, source models.Source) []ArticleData {
	var articles []ArticleData

	doc.Find("li.story").Each(func(i int, s *goquery.Selection) {
		titleLink := s.Find("a.u-url").First()
		title := strings.TrimSpace(titleLink.Text())
		url, exists := titleLink.Attr("href")

		if !exists || title == "" {
			return
		}

		if strings.HasPrefix(url, "/") {
			url = "https://lobste.rs" + url
		}

		articles = append(articles, ArticleData{
			Title:       title,
			URL:         url,
			Summary:     "",
			ContentHash: helpers.GenerateContentHash(title, ""),
		})
	})

	return articles
}

// Helper function
func extractFromSelection(s *goquery.Selection, source models.Source) *ArticleData {
	title := s.Find("h1, h2, h3, a").First().Text()
	title = strings.TrimSpace(title)

	url, exists := s.Find("a").First().Attr("href")
	if !exists || url == "" || title == "" || len(title) < 10 {
		return nil
	}

	if strings.HasPrefix(url, "/") {
		url = source.URL + url
	}

	summary := s.Find("p").First().Text()
	summary = strings.TrimSpace(summary)
	if len(summary) > 500 {
		summary = summary[:500] + "..."
	}

	return &ArticleData{
		Title:       title,
		URL:         url,
		Summary:     summary,
		ContentHash: helpers.GenerateContentHash(title, summary),
	}
}
