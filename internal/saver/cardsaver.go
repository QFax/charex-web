package saver

import (
	"charex/internal/core"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"bytes"

	"github.com/murkland/pngchunks"
)

var (
	// A regular expression to catch characters that are not safe for filenames.
	unsafeChars = regexp.MustCompile(`[\\?%*:|"<>]`)
)

// sanitizeFilename replaces unsafe characters with underscores and cleans up the name.
func sanitizeFilename(name string) string {
	// Replace unsafe characters with an underscore.
	sanitized := unsafeChars.ReplaceAllString(name, "_")
	// Replace spaces with underscores.
	sanitized = strings.ReplaceAll(sanitized, " ", "_")
	// Reduce multiple underscores to a single one.
	sanitized = regexp.MustCompile(`_+`).ReplaceAllString(sanitized, "_")
	// Trim leading/trailing underscores.
	sanitized = strings.Trim(sanitized, "_")

	if sanitized == "" {
		return "unnamed_character"
	}
	return sanitized
}

// SaveCard performs the complete save operation for a character card.
func SaveCard(card *core.TavernCardV2, rawData []byte, cardImage []byte, source string) error {
	// Create the source-specific directory.
	outputDir := "output"
	sourceDir := filepath.Join(outputDir, source)
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		return fmt.Errorf("failed to create source directory: %w", err)
	}

	// Determine the base filename from the display name or the card name.
	baseFilename := card.DisplayName
	if baseFilename == "" {
		baseFilename = card.Data.Name
	}
	baseFilename = sanitizeFilename(baseFilename)
	log.Printf("Saving card with base filename: %s", baseFilename)

	// 1. Save the raw data.
	rawPath := filepath.Join(sourceDir, baseFilename+".raw.json")
	if err := ioutil.WriteFile(rawPath, rawData, 0644); err != nil {
		return fmt.Errorf("failed to save raw data: %w", err)
	}

	// 2. Save the V2 JSON data.
	v2Json, err := json.MarshalIndent(card, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal v2 json: %w", err)
	}
	v2Path := filepath.Join(sourceDir, baseFilename+".v2.json")
	if err := ioutil.WriteFile(v2Path, v2Json, 0644); err != nil {
		return fmt.Errorf("failed to save v2 json: %w", err)
	}

	// 3. Save the PNG with embedded data, if an image is provided.
	if cardImage != nil {
		pngPath := filepath.Join(sourceDir, baseFilename+".png")
		if err := embedDataInPng(cardImage, v2Json, pngPath); err != nil {
			return fmt.Errorf("failed to save png with embedded data: %w", err)
		}
	}

	return nil
}

// embedDataInPng injects the character data into a new tEXt chunk and saves the new PNG.
func embedDataInPng(imageData, jsonData []byte, outputPath string) error {
	// Base64 encode the JSON data.
	encodedJson := base64.StdEncoding.EncodeToString(jsonData)
	textChunkData := []byte("chara\x00" + encodedJson)

	// Create a reader for the original image.
	reader, err := pngchunks.NewReader(bytes.NewReader(imageData))
	if err != nil {
		return fmt.Errorf("failed to create png reader: %w", err)
	}

	// Create the output file.
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output png file: %w", err)
	}
	defer f.Close()

	// Create a writer for the new image.
	writer, err := pngchunks.NewWriter(f)
	if err != nil {
		return fmt.Errorf("failed to create png writer: %w", err)
	}

	// Read the first chunk (IHDR) and write it to the new file.
	ihdr, err := reader.NextChunk()
	if err != nil {
		return fmt.Errorf("failed to read IHDR chunk: %w", err)
	}
	if err := writer.WriteChunk(ihdr.Length(), ihdr.Type(), ihdr); err != nil {
		return fmt.Errorf("failed to write IHDR chunk: %w", err)
	}
	ihdr.Close()

	// Write our new tEXt chunk.
	if err := writer.WriteChunk(int32(len(textChunkData)), "tEXt", bytes.NewReader(textChunkData)); err != nil {
		return fmt.Errorf("failed to write tEXt chunk: %w", err)
	}

	// Copy the rest of the chunks from the original image to the new one.
	for {
		chunk, err := reader.NextChunk()
		if err != nil {
			break // End of chunks
		}
		if err := writer.WriteChunk(chunk.Length(), chunk.Type(), chunk); err != nil {
			return fmt.Errorf("failed to write chunk: %w", err)
		}
		chunk.Close()
	}

	return nil
}