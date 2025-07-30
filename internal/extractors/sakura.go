package extractors

import (
	"bytes"
	"charex/internal/core"
	"fmt"
	"image/jpeg" // Add jpeg support
	"image/png"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/image/webp"
)

// SakuraFMExtractor specializes in extracting character data from Sakura.fm URLs.
type SakuraFMExtractor struct{}

// NewSakuraFMExtractor creates a new instance of the SakuraFMExtractor.
func NewSakuraFMExtractor() *SakuraFMExtractor {
	return &SakuraFMExtractor{}
}

// Extract fetches the content from a Sakura.fm URL and parses it to create a character card.
func (e *SakuraFMExtractor) Extract(input []byte) (*core.TavernCardV2, []byte, []byte, error) {
	url := string(input)
	if !strings.Contains(url, "sakura.fm") {
		return nil, nil, nil, fmt.Errorf("invalid url: not a sakura.fm url")
	}

	// Fetch the HTML page.
	res, err := http.Get(url)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to fetch url: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, nil, nil, fmt.Errorf("failed to fetch url: status code %d", res.StatusCode)
	}

	// Read the body of the response.
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to read response body: %w", err)
	}
	rawData := body

	// Load the HTML document.
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to parse html: %w", err)
	}

	// Extract data using selectors from the TypeScript file.
	container := doc.Find("div.flex.flex-col.space-y-6.pt-6")
	name := strings.TrimSpace(container.Find(".text-muted-foreground.line-clamp-2").First().Text())
	description := strings.TrimSpace(container.Find(".text-muted-foreground.line-clamp-3").First().Text())
	scenario := strings.TrimSpace(container.Find(".text-muted-foreground.line-clamp-5").First().Text())
	firstMes := strings.TrimSpace(doc.Find(".bg-message-assistant").First().Text())
	creator := "Anonymous" // Default value
	doc.Find("div.font-bold").Each(func(i int, s *goquery.Selection) {
		if strings.TrimSpace(s.Text()) == "Creator" {
			creatorName := s.Next().Find("span.flex-1.truncate.tracking-tight").Text()
			if creatorName != "" {
				creator = strings.TrimSpace(creatorName)
			}
		}
	})
	log.Printf("Extracted data: name='%s', description='%s', scenario='%s', firstMes='%s', creator='%s'", name, description, scenario, firstMes, creator)
	cardData := core.TavernCardData{
		Name:                   name,
		Description:            scenario,
		Scenario:               "",
		FirstMes:               firstMes,
		Creator:                creator,
		Personality:            "",
		MesExample:             "",
		CreatorNotes:           description,
		SystemPrompt:           "",
		PostHistoryInstructions: "",
		AlternateGreetings:     []string{},
		Tags:                   []string{"SakuraFM"},
		CharacterVersion:       "1.0",
		Extensions:             make(map[string]interface{}),
	}

	// Create the full V2 card.
	card := &core.TavernCardV2{
		Spec:        "chara_card_v2",
		SpecVersion: "2.0",
		Data:        cardData,
		DisplayName: cardData.Name,
	}

	// Extract the character image.
	var cardImage []byte
	imgSrc, exists := doc.Find("img.mx-auto.h-\\[200px\\].w-\\[200px\\].rounded-md.object-cover").Attr("src")
	if exists {
		cardImage, err = downloadImage(imgSrc)
		if err != nil {
			// We can consider this a non-fatal error and continue without an image.
			fmt.Printf("Warning: failed to download character image: %v\n", err)
		}
	}

	return card, rawData, cardImage, nil
}

// downloadImage fetches an image from a URL and returns it as a byte slice.
func downloadImage(url string) ([]byte, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("failed to download image: status code %d", res.StatusCode)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	contentType := http.DetectContentType(body)
	fmt.Printf("Detected image Content-Type: %s\n", contentType) // Add logging
	if strings.Contains(contentType, "webp") {
		img, err := webp.Decode(bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("failed to decode webp: %w", err)
		}

		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
			return nil, fmt.Errorf("failed to encode png: %w", err)
		}
		return buf.Bytes(), nil
	} else if strings.Contains(contentType, "jpeg") {
		img, err := jpeg.Decode(bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("failed to decode jpeg: %w", err)
		}
		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
			return nil, fmt.Errorf("failed to encode png: %w", err)
		}
		return buf.Bytes(), nil
	}

	// Assume it's a PNG if not WEBP, as per original logic.
	// The saver will validate if it's a valid PNG later.
	return body, nil
}