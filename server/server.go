package server

import (
	"context"
	_ "embed"
	"errors"
	"github.com/Av1shay/adstext-search/model"
	"github.com/Av1shay/adstext-search/search"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"
)

//go:embed templates/index.html
var indexContent string

type DomainResult struct {
	Name       string
	Missing    []model.AdstxtLine
	MissingLen int
	MissingCSS template.CSS
}

type SearchMacro struct {
	DomainResults []DomainResult
	Total         int
	Lines         string
	Domains       string
	Err           string
}

type Server struct {
	indexTmpl    *template.Template
	searchClient *search.Client
}

func New(searchClient *search.Client) (*Server, error) {
	indexTmpl, err := template.New("index").Parse(indexContent)
	if err != nil {
		return nil, err
	}
	return &Server{
		indexTmpl,
		searchClient,
	}, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		if err := s.indexTmpl.Execute(w, SearchMacro{}); err != nil {
			log.Println("failed to execute empty template:", err)
			w.Write([]byte("something went wrong :("))
		}
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	if err := r.ParseForm(); err != nil {
		log.Println("r.ParseForm() failed:", err)
		return
	}
	linesRaw := r.FormValue("adstext-lines")
	domainsRaw := r.FormValue("domains")

	lines, domains, err := parseFormValues(linesRaw, domainsRaw)
	if err != nil {
		if err := s.indexTmpl.Execute(w, SearchMacro{Err: err.Error()}); err != nil {
			log.Println("failed to execute template:", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("something went wrong :("))
		}
		return
	}
	searchRes, err := s.searchClient.Do(ctx, lines, domains)
	if err != nil {
		if err := s.indexTmpl.Execute(w, SearchMacro{Err: err.Error()}); err != nil {
			log.Println("failed to execute template:", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("something went wrong :("))
		}
		return
	}

	domainResults := make([]DomainResult, 0, len(searchRes))
	for domain, missingLines := range searchRes {
		missingLen := len(missingLines)
		missingCSS := `color:orange`
		if missingLen == len(lines) {
			missingCSS = `color:red`
		} else if missingLen == 0 {
			missingCSS = `color:green`
		}
		domainResults = append(domainResults, DomainResult{
			Name:       domain,
			Missing:    missingLines,
			MissingLen: missingLen,
			MissingCSS: template.CSS(missingCSS),
		})
	}

	if err := s.indexTmpl.Execute(w, SearchMacro{
		DomainResults: domainResults,
		Total:         len(lines),
		Lines:         linesRaw,
		Domains:       domainsRaw,
	}); err != nil {
		log.Println("failed to execute template:", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("something went wrong :("))
	}
}

func parseFormValues(linesRaw, domainsRaw string) ([]model.AdstxtLine, []string, error) {
	if linesRaw == "" {
		return nil, nil, errors.New("lines can't be empty")
	}
	if domainsRaw == "" {
		return nil, nil, errors.New("domains can't be empty")
	}
	lines := strings.Split(linesRaw, "\n")
	adstxtLines := make([]model.AdstxtLine, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSuffix(line, "\r")
		parts := strings.Split(line, ",")
		if len(parts) < 3 {
			continue
		}
		adstxtLine := model.AdstxtLine{
			Host:          strings.TrimSpace(parts[0]),
			SellerID:      strings.TrimSpace(parts[1]),
			PublisherType: model.PublisherType(strings.TrimSpace(parts[2])),
		}
		if len(parts) > 3 {
			adstxtLine.PublisherID = strings.TrimSpace(parts[3])
		}
		adstxtLines = append(adstxtLines, adstxtLine)
	}

	domainLines := strings.Split(domainsRaw, "\n")
	domains := make([]string, 0, len(domainLines))
	for _, u := range domainLines {
		domains = append(domains, strings.TrimSuffix(strings.TrimSuffix(strings.TrimSpace(u), "\t"), "/"))
	}

	return adstxtLines, domains, nil
}
