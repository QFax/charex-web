package web

import (
	"charex/internal/core"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// CardSource represents a source of character cards (e.g., 'sakura').
type CardSource struct {
	Name  string                `json:"name"`
	Cards []core.TavernCardV2   `json:"cards"`
}

// CardsResponse is the structure for the GET /api/cards response.
type CardsResponse struct {
	Sources []CardSource `json:"sources"`
}

func (s *Server) GetCards(w http.ResponseWriter, r *http.Request) {
	sources, err := s.scanForCardSources("output")
	if err != nil {
		http.Error(w, "Failed to scan for card sources", http.StatusInternalServerError)
		return
	}

	response := CardsResponse{Sources: sources}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (s *Server) scanForCardSources(outputDir string) ([]CardSource, error) {
	var sources []CardSource

	sourceDirs, err := ioutil.ReadDir(outputDir)
	if err != nil {
		return nil, err
	}

	for _, sourceDir := range sourceDirs {
		if !sourceDir.IsDir() {
			continue
		}

		sourceName := sourceDir.Name()
		sourcePath := filepath.Join(outputDir, sourceName)
		cards, err := s.loadCardsFromSource(sourcePath)
		if err != nil {
			log.Printf("Error loading cards from source %s: %v", sourceName, err)
			continue
		}

		sources = append(sources, CardSource{
			Name:  sourceName,
			Cards: cards,
		})
	}

	return sources, nil
}

func (s *Server) loadCardsFromSource(sourcePath string) ([]core.TavernCardV2, error) {
	var cards []core.TavernCardV2
	var files []os.FileInfo

	err := filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".v2.json") {
			files = append(files, info)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().After(files[j].ModTime())
	})

	for _, file := range files {
		card, err := s.loadCard(filepath.Join(sourcePath, file.Name()))
		if err != nil {
			log.Printf("Error loading card %s: %v", file.Name(), err)
			continue
		}
		cards = append(cards, *card)
	}

	return cards, nil
}

func (s *Server) loadCard(cardPath string) (*core.TavernCardV2, error) {
	data, err := ioutil.ReadFile(cardPath)
	if err != nil {
		return nil, err
	}

	var card core.TavernCardV2
	if err := json.Unmarshal(data, &card); err != nil {
		return nil, err
	}

	return &card, nil
}