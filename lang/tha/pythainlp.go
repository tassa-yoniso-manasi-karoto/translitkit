package tha

import (
	"context"
	"fmt"
	"time"

	"github.com/tassa-yoniso-manasi-karoto/go-pythainlp"
	"github.com/tassa-yoniso-manasi-karoto/translitkit/common"
)

// PyThaiNLPProvider implements the Provider interface using go-pythainlp
// It can operate in two modes:
// - TokenizerType: Only tokenization
// - CombinedType: Tokenization + romanization
type PyThaiNLPProvider struct {
	manager          *pythainlp.PyThaiNLPManager
	config           map[string]interface{}
	operatingMode    common.ProviderType
	romanEngine      string
	progressCallback common.ProgressCallback
}

// NewPyThaiNLPProvider creates a new provider with specified operating mode
func NewPyThaiNLPProvider(mode common.ProviderType) *PyThaiNLPProvider {
	return &PyThaiNLPProvider{
		operatingMode: mode,
		romanEngine:   pythainlp.EngineRoyin, // default
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
	// Create PyThaiNLP manager - always use lightweight mode for translitkit
	manager, err := pythainlp.NewManager(ctx,
		pythainlp.WithQueryTimeout(30*time.Second),
		pythainlp.WithLightweightMode(true))
	if err != nil {
		return fmt.Errorf("failed to create PyThaiNLP manager: %w", err)
	}
	
	// Initialize the manager
	if err := manager.Init(ctx); err != nil {
		return fmt.Errorf("failed to initialize PyThaiNLP: %w", err)
	}
	
	p.manager = manager
	return nil
}

// Init initializes the provider with background context
func (p *PyThaiNLPProvider) Init() error {
	return p.InitWithContext(context.Background())
}

// InitRecreateWithContext reinitializes the provider
func (p *PyThaiNLPProvider) InitRecreateWithContext(ctx context.Context, noCache bool) error {
	if p.manager != nil {
		p.manager.Close()
	}
	
	// Always use lightweight mode for translitkit
	manager, err := pythainlp.NewManager(ctx,
		pythainlp.WithQueryTimeout(30*time.Second),
		pythainlp.WithLightweightMode(true))
	if err != nil {
		return fmt.Errorf("failed to create PyThaiNLP manager: %w", err)
	}
	
	if err := manager.InitRecreate(ctx, noCache); err != nil {
		return fmt.Errorf("failed to recreate PyThaiNLP: %w", err)
	}
	
	p.manager = manager
	return nil
}

// InitRecreate reinitializes with background context
func (p *PyThaiNLPProvider) InitRecreate(noCache bool) error {
	return p.InitRecreateWithContext(context.Background(), noCache)
}

// CloseWithContext releases resources
func (p *PyThaiNLPProvider) CloseWithContext(ctx context.Context) error {
	if p.manager != nil {
		return p.manager.Close()
	}
	return nil
}

// Close releases resources with background context
func (p *PyThaiNLPProvider) Close() error {
	return p.CloseWithContext(context.Background())
}

// ProcessFlowController processes input based on operating mode
func (p *PyThaiNLPProvider) ProcessFlowController(ctx context.Context, input common.AnyTokenSliceWrapper) (common.AnyTokenSliceWrapper, error) {
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
		
		if p.operatingMode == common.TokenizerType {
			tokens, err = p.tokenizeOnly(ctx, chunk)
		} else { // CombinedType
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

// Name returns the provider name based on operating mode
func (p *PyThaiNLPProvider) Name() string {
	if p.operatingMode == common.TokenizerType {
		return "pythainlp-tokenizer"
	}
	return "pythainlp"
}

// GetType returns the provider type
func (p *PyThaiNLPProvider) GetType() common.ProviderType {
	return p.operatingMode
}

// GetMaxQueryLen returns the maximum query length
func (p *PyThaiNLPProvider) GetMaxQueryLen() int {
	// PyThaiNLP can handle large texts, but we'll chunk for progress reporting
	return 5000
}