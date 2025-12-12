package zho

import (
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/tassa-yoniso-manasi-karoto/translitkit/common"
	"github.com/yanyiwu/gojieba"
)

// Dictionary files required by gojieba with their expected sizes for progress tracking
var dictFiles = []struct {
	name string
	size int64
}{
	{"jieba.dict.utf8", 5079385},
	{"hmm_model.utf8", 519568},
	{"user.dict.utf8", 49},
	{"idf.utf8", 6083765},
	{"stop_words.utf8", 8987},
}

// dictBaseURL is the base URL for downloading dictionary files from gojieba's GitHub repo
const dictBaseURL = "https://raw.githubusercontent.com/yanyiwu/gojieba/v1.4.6/deps/cppjieba/dict/"

// GoJiebaProvider implements the Provider interface for Chinese text segmentation.
// It uses the gojieba library to tokenize Chinese text with word boundaries and
// part-of-speech tagging, while preserving non-lexical tokens like punctuation.
type GoJiebaProvider struct {
	config                   map[string]interface{}
	progressCallback         common.ProgressCallback
	downloadProgressCallback common.DownloadProgressCallback
	jieba                    *gojieba.Jieba
}

// WithProgressCallback sets a callback function for reporting progress during processing.
// The callback will be invoked with the current chunk index and total number of chunks.
func (p *GoJiebaProvider) WithProgressCallback(callback common.ProgressCallback) {
	p.progressCallback = callback
}

// WithDownloadProgressCallback sets a callback for download progress during dictionary downloads.
func (p *GoJiebaProvider) WithDownloadProgressCallback(callback common.DownloadProgressCallback) {
	p.downloadProgressCallback = callback
}

// SaveConfig stores the configuration for later application during initialization.
// This allows the provider to be configured before being initialized.
//
// Returns an error if the configuration is invalid.
func (p *GoJiebaProvider) SaveConfig(cfg map[string]interface{}) error {
	p.config = cfg
	return nil
}

// InitWithContext initializes the gojieba engine with the given context.
// This is called automatically before processing if the engine is not already initialized.
// On first run, it downloads the required dictionary files (~14MB) to the user's data directory.
// The context can be used for cancellation during initialization or download.
//
// Returns an error if initialization fails, download fails, or the context is canceled.
func (p *GoJiebaProvider) InitWithContext(ctx context.Context) error {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("gojieba: context canceled during initialization: %w", err)
	}

	if p.jieba != nil {
		return nil
	}

	// Get/create dictionary directory
	dictDir, err := ensureDictDir()
	if err != nil {
		return fmt.Errorf("gojieba: failed to create dictionary directory: %w", err)
	}

	// Download dictionaries if needed
	if err := p.ensureDictionaries(ctx, dictDir); err != nil {
		return fmt.Errorf("gojieba: failed to download dictionaries: %w", err)
	}

	// Pass explicit paths to NewJieba to avoid runtime.Caller path issues
	p.jieba = gojieba.NewJieba(
		filepath.Join(dictDir, "jieba.dict.utf8"),
		filepath.Join(dictDir, "hmm_model.utf8"),
		filepath.Join(dictDir, "user.dict.utf8"),
		filepath.Join(dictDir, "idf.utf8"),
		filepath.Join(dictDir, "stop_words.utf8"),
	)
	return nil
}

// Init initializes the provider with a background context.
// This is a convenience method for operations that don't need cancellation control.
//
// Returns an error if initialization fails.
func (p *GoJiebaProvider) Init() error {
	return p.InitWithContext(context.Background())
}

// InitRecreateWithContext frees existing gojieba resources and re-initializes from scratch.
// This is useful when dictionary paths or other configuration changes.
// The context can be used for cancellation during reinitialization.
//
// Returns an error if reinitialization fails or the context is canceled.
func (p *GoJiebaProvider) InitRecreateWithContext(ctx context.Context, noCache bool) error {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("gojieba: context canceled during reinitialization: %w", err)
	}
	
	if p.jieba != nil {
		p.jieba.Free()
		p.jieba = nil
	}
	return p.InitWithContext(ctx)
}

// InitRecreate reinitializes the provider with a background context.
// This is a convenience method for operations that don't need cancellation control.
//
// Returns an error if reinitialization fails.
func (p *GoJiebaProvider) InitRecreate(noCache bool) error {
	return p.InitRecreateWithContext(context.Background(), noCache)
}

// ProcessFlowController processes input tokens using the specified context.
// This handles raw input chunks and performs Chinese word segmentation with POS tagging.
// The context is used for cancellation during processing.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - input: The token slice wrapper containing raw input chunks
//
// Returns:
//   - AnyTokenSliceWrapper: A wrapper containing the processed tokens
//   - error: An error if processing fails, the context is canceled, or initialization fails
func (p *GoJiebaProvider) ProcessFlowController(ctx context.Context, mode common.OperatingMode, input common.AnyTokenSliceWrapper) (common.AnyTokenSliceWrapper, error) {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("gojieba: context canceled during processing: %w", err)
	}
	
	// Ensure gojieba is initialized
	if p.jieba == nil {
		if err := p.InitWithContext(ctx); err != nil {
			return nil, fmt.Errorf("failed to init gojieba: %w", err)
		}
	}

	rawChunks := input.GetRaw()
	if len(rawChunks) == 0 {
		return input, nil
	}

	outWrapper := &TknSliceWrapper{}
	totalChunks := len(rawChunks)

	for idx, chunk := range rawChunks {
		// Check for context cancellation
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("gojieba: context canceled while processing chunk %d: %w", idx, err)
		}
		
		// Report progress if callback is set
		if p.progressCallback != nil {
			p.progressCallback(idx, totalChunks)
		}
		
		if chunk == "" {
			continue
		}

		// 1) Use gojieba for lexical segmentation + POS tags
		words := p.jieba.Cut(chunk, true) // "precise" mode with HMM
		tags := p.jieba.Tag(chunk)
		if len(words) != len(tags) {
			return nil, fmt.Errorf("gojieba mismatch: len(words)=%d, len(tags)=%d", len(words), len(tags))
		}

		// 2) Integrate lexical tokens with filler
		integrated := common.IntegrateProviderTokens(chunk, words)

		// We'll attach each recognized lexical token's POS from 'tags' in order
		lexCount := 0
		for _, fillerOrLex := range integrated {
			// Build a new zho.Tkn from the integrated token
			zhoTkn := &Tkn{
				Tkn: *fillerOrLex,

				// For Chinese tokens, we can at least guess that 'Surface' is both
				// the simplified and traditional form if we have no external DB:
				Simplified:  fillerOrLex.Surface,
				Traditional: fillerOrLex.Surface,

				// We won't fill `NumStrokes`, `Radical`, etc. because gojieba
				// doesn't supply stroke or radical data.
				// We'll also leave morphological + idiomatic fields at defaults.
			}

			if fillerOrLex.IsLexical {
				// The next POS tag in 'tags' corresponds to this lexical word
				pos := tags[lexCount]
				lexCount++

				// Store generic POS in Tkn.PartOfSpeech
				zhoTkn.PartOfSpeech = pos

				// If the user wants, store that same POS in some custom field, or
				// interpret it further:
				// For instance, if POS == "q", we might guess it's a classifier.
				if pos == "q" {
					// Mark it as a classifier
					zhoTkn.ClassifierType = "indiv" // naive assumption
				}

				// If you want to guess if it's an idiom or something from the POS,
				// you could do so here, though gojieba's default dict rarely marks that.

				// If we see 'a' (形容词), we might guess it's a stative verb in Chinese:
				if pos == "a" {
					zhoTkn.IsStative = true
				}
			}

			// Append the new token
			outWrapper.Append(zhoTkn)
		}
	}
	
	// Report completion if callback is set
	if p.progressCallback != nil {
		p.progressCallback(totalChunks, totalChunks)
	}

	// Clear raw chunks to mark they've been processed
	input.ClearRaw()

	return outWrapper, nil
}

// Name returns the unique name of this provider.
func (p *GoJiebaProvider) Name() string {
	return "gojieba"
}

// SupportedModes returns the operating modes this provider supports.
func (p *GoJiebaProvider) SupportedModes() []common.OperatingMode {
	return []common.OperatingMode{common.TokenizerMode}
}

// GetMaxQueryLen returns a large number so the module can handle big input.
func (p *GoJiebaProvider) GetMaxQueryLen() int {
	return math.MaxInt32
}

// CloseWithContext releases resources used by the provider with the given context.
// This frees the gojieba instance to release memory.
// The context can be used for cancellation during resource release.
//
// Returns an error if closing fails or the context is canceled.
func (p *GoJiebaProvider) CloseWithContext(ctx context.Context) error {
	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("gojieba: context canceled during close: %w", err)
	}
	
	if p.jieba != nil {
		p.jieba.Free()
		p.jieba = nil
	}
	return nil
}

// Close releases resources used by the provider with a background context.
// This is a convenience method for operations that don't need cancellation control.
//
// Returns an error if closing fails.
func (p *GoJiebaProvider) Close() error {
	return p.CloseWithContext(context.Background())
}

// ensureDictDir creates and returns the dictionary directory path.
// Uses XDG base directory specification for cross-platform support:
// - Linux: ~/.local/share/langkit/gojieba/dict/
// - macOS: ~/Library/Application Support/langkit/gojieba/dict/
// - Windows: %APPDATA%\langkit\gojieba\dict\
func ensureDictDir() (string, error) {
	dictDir := filepath.Join(xdg.DataHome, "langkit", "gojieba", "dict")
	return dictDir, os.MkdirAll(dictDir, 0755)
}

// ensureDictionaries checks if all dictionary files exist, and downloads any missing ones.
func (p *GoJiebaProvider) ensureDictionaries(ctx context.Context, dictDir string) error {
	// Check if all files already exist
	allExist := true
	for _, df := range dictFiles {
		if _, err := os.Stat(filepath.Join(dictDir, df.name)); os.IsNotExist(err) {
			allExist = false
			break
		}
	}
	if allExist {
		return nil
	}

	// Calculate total size for progress tracking
	var totalSize int64
	for _, df := range dictFiles {
		totalSize += df.size
	}

	// Download each file with progress
	var downloaded int64
	for _, df := range dictFiles {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("context canceled: %w", err)
		}

		destPath := filepath.Join(dictDir, df.name)
		if _, err := os.Stat(destPath); err == nil {
			// File already exists, count it as downloaded for progress
			downloaded += df.size
			continue
		}

		if err := p.downloadFile(ctx, dictBaseURL+df.name, destPath, &downloaded, totalSize); err != nil {
			return fmt.Errorf("failed to download %s: %w", df.name, err)
		}
	}
	return nil
}

// downloadFile downloads a single file from url to destPath, updating progress.
func (p *GoJiebaProvider) downloadFile(ctx context.Context, url, destPath string, downloaded *int64, totalSize int64) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// Create temp file first, then rename for atomicity
	tmpPath := destPath + ".tmp"
	out, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() {
		out.Close()
		os.Remove(tmpPath) // Clean up temp file on error
	}()

	// Copy with progress tracking
	buf := make([]byte, 32*1024)
	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := out.Write(buf[:n]); writeErr != nil {
				return fmt.Errorf("failed to write: %w", writeErr)
			}
			*downloaded += int64(n)
			if p.downloadProgressCallback != nil {
				p.downloadProgressCallback(*downloaded, totalSize, "Downloading GoJieba dictionaries...")
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return fmt.Errorf("failed to read: %w", readErr)
		}
	}

	// Close before rename
	if err := out.Close(); err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpPath, destPath); err != nil {
		return fmt.Errorf("failed to rename: %w", err)
	}

	return nil
}
