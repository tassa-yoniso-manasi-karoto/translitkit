package tha

import (
	"context"
	"fmt"
	"time"

	"github.com/tassa-yoniso-manasi-karoto/go-pythainlp"
	"github.com/tassa-yoniso-manasi-karoto/translitkit/common"
)

// =============================================================================
// DOCKER CONTAINER LIFECYCLE - IMPORTANT FOR FUTURE DEVELOPERS/LLMs
// =============================================================================
//
// PyThaiNLPProvider is the OWNER of the pythainlp Docker container lifecycle.
// When this provider initializes, it starts the Docker container. When it closes,
// the container is stopped.
//
// OTHER PROVIDERS (like PaiboonizerProvider) that depend on pythainlp MUST NOT
// create their own pythainlp.PyThaiNLPManager. Instead, they should:
//   1. Use go-pythainlp's package-level functions (e.g., pythainlp.SyllableTokenize())
//      which use a default manager that reuses any existing container
//   2. Rely on this provider being initialized first in hybrid schemes
//
// This design prevents:
//   - Multiple managers fighting over the same Docker container
//   - Race conditions during container startup/shutdown
//   - Resource leaks from orphaned containers
//
// In hybrid schemes like "paiboon-hybrid" (pythainlp → paiboonizer):
//   - pythainlp provider starts the container (for word tokenization)
//   - paiboonizer reuses the same container (for syllable tokenization via package-level funcs)
//   - When pythainlp provider closes, the container shuts down
//
// =============================================================================

// PyThaiNLPProvider implements the Provider interface using go-pythainlp
// It can operate in two modes:
// - TokenizerMode: Only tokenization
// - CombinedMode: Tokenization + romanization
type PyThaiNLPProvider struct {
	manager                  *pythainlp.PyThaiNLPManager
	config                   map[string]interface{}
	romanEngine              string
	progressCallback         common.ProgressCallback
	downloadProgressCallback common.DownloadProgressCallback
}

// NewPyThaiNLPProvider creates a new provider
func NewPyThaiNLPProvider() *PyThaiNLPProvider {
	return &PyThaiNLPProvider{
		romanEngine: pythainlp.EngineRoyin, // default
		config:      make(map[string]interface{}),
	}
}

// SaveConfig stores configuration for later application during initialization
func (p *PyThaiNLPProvider) SaveConfig(cfg map[string]interface{}) error {
	p.config = cfg
	
	// Extract romanization engine if specified
	if engine, ok := cfg["roman_engine"].(string); ok {
		// Validate engine is supported in lightweight mode
		switch engine {
		case pythainlp.EngineRoyin, pythainlp.EngineTLTKRom, pythainlp.EngineLookup:
			p.romanEngine = engine
		default:
			return fmt.Errorf("romanization engine '%s' not supported in lightweight mode", engine)
		}
	}
	
	// Handle scheme configuration from translitkit
	if scheme, ok := cfg["scheme"].(string); ok {
		// Map scheme names to engines
		switch scheme {
		case "royin":
			p.romanEngine = pythainlp.EngineRoyin
		case "tltk":
			p.romanEngine = pythainlp.EngineTLTKRom
		case "lookup":
			p.romanEngine = pythainlp.EngineLookup
		default:
			return fmt.Errorf("romanization scheme '%s' not supported", scheme)
		}
	}
	
	return nil
}

// InitWithContext initializes the provider with context
func (p *PyThaiNLPProvider) InitWithContext(ctx context.Context) error {
	// Build manager options
	opts := []pythainlp.ManagerOption{
		pythainlp.WithQueryTimeout(30 * time.Second),
		pythainlp.WithLightweightMode(true),
	}

	// Add download progress callback if set, wrapping to inject provider name
	if p.downloadProgressCallback != nil {
		opts = append(opts, pythainlp.WithDownloadProgressCallback(func(current, total int64, status string) {
			p.downloadProgressCallback(p.Name(), current, total, status)
		}))
	}

	// Create PyThaiNLP manager - always use lightweight mode for translitkit
	manager, err := pythainlp.NewManager(ctx, opts...)
	if err != nil {
		return fmt.Errorf("failed to create PyThaiNLP manager: %w", err)
	}

	// Use InitRecreate instead of Init to handle port mismatches
	// Each NewManager allocates a new port, but an existing stopped container
	// has the old port mapping. InitRecreate removes and recreates the container
	// with the correct port binding.
	if err := manager.InitRecreate(ctx, false); err != nil {
		return fmt.Errorf("failed to initialize PyThaiNLP: %w", err)
	}

	p.manager = manager

	// Set as the default manager so package-level functions work.
	// This is critical for PaiboonizerProvider which uses pythainlp.SyllableTokenize()
	// (a package-level function) to reuse this container instead of creating a new one.
	pythainlp.SetDefaultManager(manager)

	return nil
}

// Init initializes the provider with background context
func (p *PyThaiNLPProvider) Init() error {
	return p.InitWithContext(context.Background())
}

// InitRecreateWithContext reinitializes the provider
func (p *PyThaiNLPProvider) InitRecreateWithContext(ctx context.Context, noCache bool) error {
	if p.manager != nil {
		pythainlp.ClearDefaultManager()
		p.manager.Close()
	}

	// Build manager options
	opts := []pythainlp.ManagerOption{
		pythainlp.WithQueryTimeout(30 * time.Second),
		pythainlp.WithLightweightMode(true),
	}

	// Add download progress callback if set, wrapping to inject provider name
	if p.downloadProgressCallback != nil {
		opts = append(opts, pythainlp.WithDownloadProgressCallback(func(current, total int64, status string) {
			p.downloadProgressCallback(p.Name(), current, total, status)
		}))
	}

	manager, err := pythainlp.NewManager(ctx, opts...)
	if err != nil {
		return fmt.Errorf("failed to create PyThaiNLP manager: %w", err)
	}

	if err := manager.InitRecreate(ctx, noCache); err != nil {
		return fmt.Errorf("failed to recreate PyThaiNLP: %w", err)
	}

	p.manager = manager
	pythainlp.SetDefaultManager(manager)
	return nil
}

// InitRecreate reinitializes with background context
func (p *PyThaiNLPProvider) InitRecreate(noCache bool) error {
	return p.InitRecreateWithContext(context.Background(), noCache)
}

// CloseWithContext releases resources
func (p *PyThaiNLPProvider) CloseWithContext(ctx context.Context) error {
	if p.manager != nil {
		// Clear default manager reference before closing to prevent stale references
		pythainlp.ClearDefaultManager()
		return p.manager.Close()
	}
	return nil
}

// Close releases resources with background context
func (p *PyThaiNLPProvider) Close() error {
	return p.CloseWithContext(context.Background())
}

// ProcessFlowController processes input based on operating mode
func (p *PyThaiNLPProvider) ProcessFlowController(ctx context.Context, mode common.OperatingMode, input common.AnyTokenSliceWrapper) (common.AnyTokenSliceWrapper, error) {
	raw := input.GetRaw()
	if input.Len() == 0 && len(raw) == 0 {
		return nil, fmt.Errorf("empty input")
	}
	
	// We expect raw input chunks
	if len(raw) == 0 {
		return nil, fmt.Errorf("PyThaiNLP provider requires raw text input")
	}
	
	tsw := &TknSliceWrapper{}
	totalChunks := len(raw)
	
	for idx, chunk := range raw {
		if p.progressCallback != nil {
			p.progressCallback(idx, totalChunks)
		}
		
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		
		var tokens []*Tkn
		var err error
		
		// Process based on the specified mode
		if mode == common.TokenizerMode {
			tokens, err = p.tokenizeOnly(ctx, chunk)
		} else { // CombinedMode
			tokens, err = p.analyzeText(ctx, chunk)
		}
		
		if err != nil {
			return nil, fmt.Errorf("processing chunk %d failed: %w", idx, err)
		}
		
		// Convert to TknSliceWrapper
		for _, token := range tokens {
			tsw.Append(token)
		}
	}
	
	return tsw, nil
}

// tokenizeOnly performs tokenization without romanization
func (p *PyThaiNLPProvider) tokenizeOnly(ctx context.Context, text string) ([]*Tkn, error) {
	result, err := p.manager.Tokenize(ctx, text)
	if err != nil {
		return nil, fmt.Errorf("tokenization failed: %w", err)
	}
	
	// Convert to Tkn using token integration
	tokens, err := common.IntegrateProviderTokensV2(text, result.Raw)
	if err != nil {
		common.Log.Debug().
			Err(err).
			Msg("Token integration had issues, continuing with partial results")
	}
	
	// Convert common.Tkn to tha.Tkn
	thaiTokens := make([]*Tkn, len(tokens))
	for i, token := range tokens {
		thaiTokens[i] = convertToThaiToken(token)
	}
	
	return thaiTokens, nil
}

// analyzeText performs both tokenization and romanization
func (p *PyThaiNLPProvider) analyzeText(ctx context.Context, text string) ([]*Tkn, error) {
	// Use the analyze API for combined operation with specified romanization engine
	opts := pythainlp.AnalyzeOptions{
		Features:       []string{"tokenize", "romanize"},
		RomanizeEngine: p.romanEngine,
	}
	
	result, err := p.manager.AnalyzeWithOptions(ctx, text, opts)
	if err != nil {
		return nil, fmt.Errorf("analysis failed: %w", err)
	}
	
	// Convert to Tkn using token integration
	tokens, err := common.IntegrateProviderTokensV2(text, result.RawTokens)
	if err != nil {
		common.Log.Debug().
			Err(err).
			Msg("Token integration had issues, continuing with partial results")
	}
	
	// Convert to Thai tokens with romanization
	thaiTokens := make([]*Tkn, len(tokens))
	for i, token := range tokens {
		thaiToken := convertToThaiToken(token)
		
		// Add romanization if available
		if i < len(result.RomanizedParts) && token.IsLexical {
			thaiToken.Romanization = result.RomanizedParts[i]
		}
		
		thaiTokens[i] = thaiToken
	}
	
	return thaiTokens, nil
}

// WithProgressCallback sets the progress callback
func (p *PyThaiNLPProvider) WithProgressCallback(callback common.ProgressCallback) {
	p.progressCallback = callback
}

// WithDownloadProgressCallback sets a callback for download progress.
// This callback is used during Docker image pull to report progress.
func (p *PyThaiNLPProvider) WithDownloadProgressCallback(callback common.DownloadProgressCallback) {
	p.downloadProgressCallback = callback
}

// Name returns the provider name
func (p *PyThaiNLPProvider) Name() string {
	return "pythainlp"
}

// SupportedModes returns the operating modes this provider supports
func (p *PyThaiNLPProvider) SupportedModes() []common.OperatingMode {
	return []common.OperatingMode{common.TokenizerMode, common.CombinedMode}
}

// GetMaxQueryLen returns the maximum query length
func (p *PyThaiNLPProvider) GetMaxQueryLen() int {
	// PyThaiNLP can handle large texts, but we'll chunk for progress reporting
	return 5000
}