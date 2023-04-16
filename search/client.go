package search

import (
	"bufio"
	"context"
	"fmt"
	"github.com/Av1shay/adstext-search/model"
	"log"
	"net/http"
	"strings"
	"sync"
)

const (
	lineSep = ":_:"
)

var linesPool = sync.Pool{
	New: func() any {
		return new(model.AdstxtLine)
	},
}

type Config struct {
	UserAgent string
	Workers   int
}

type Client struct {
	config     Config
	httpClient *http.Client
}

func NewClient(config Config, httpClient *http.Client) *Client {
	return &Client{config, httpClient}
}

func (c *Client) Do(ctx context.Context, lines []model.AdstxtLine, domains []string) (map[string][]model.AdstxtLine, error) {
	workers := c.config.Workers
	wg := sync.WaitGroup{}
	wg.Add(workers)
	domainsChan := make(chan string, workers)

	linesMapLookup := make(map[string]model.AdstxtLine, len(lines))
	for _, line := range lines {
		k := lineToKey(line)
		linesMapLookup[k] = line
	}

	res := make(map[string][]model.AdstxtLine, len(domains))
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()

			for d := range domainsChan {
				missingLines, err := c.findMissingAdstxtLines(ctx, d, linesMapLookup)
				if err != nil {
					log.Printf("findMissingAdstxtLines() error fod domain %s: %s\n", d, err)
					continue
				}
				res[d] = missingLines
			}

		}()
	}

	for _, d := range domains {
		domainsChan <- d
	}

	close(domainsChan)
	wg.Wait()

	return res, nil
}

func (c *Client) findMissingAdstxtLines(ctx context.Context, domain string, linesMapLookup map[string]model.AdstxtLine) ([]model.AdstxtLine, error) {
	u := fmt.Sprintf("https://%s/ads.txt", domain)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequest: %s", err)
	}
	req.Header.Set("User-Agent", c.config.UserAgent)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to GET %s: %s", u, err)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("GET %s status code: %d", u, resp.StatusCode)
	}
	defer resp.Body.Close()

	foundMap := make(map[string]struct{})
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), ",")
		if len(parts) < 3 {
			continue
		}
		line := linesPool.Get().(*model.AdstxtLine)
		line.Reset()
		line.Host = strings.TrimSpace(parts[0])
		line.SellerID = strings.TrimSpace(parts[1])
		line.PublisherType = model.PublisherType(strings.TrimSpace(parts[2]))
		if len(parts) > 3 {
			line.PublisherID = strings.TrimSpace(parts[3])
		}
		k := lineToKey(*line)
		linesPool.Put(line)
		if _, found := linesMapLookup[k]; found {
			foundMap[k] = struct{}{}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	missing := make([]model.AdstxtLine, 0, len(linesMapLookup))
	foundMissing := make(map[string]struct{})
	for k, line := range linesMapLookup {
		if _, found := foundMap[k]; !found {
			if _, found := foundMissing[k]; found { // skip if we already saw this line
				continue
			}
			missing = append(missing, line)
		}
	}

	return missing, nil
}

func lineToKey(line model.AdstxtLine) string {
	var sb strings.Builder
	sb.WriteString(line.Host)
	sb.WriteString(lineSep)
	sb.WriteString(line.SellerID)
	sb.WriteString(lineSep)
	sb.WriteString(string(line.PublisherType))
	sb.WriteString(lineSep)
	sb.WriteString(line.PublisherID)
	return sb.String()
}
