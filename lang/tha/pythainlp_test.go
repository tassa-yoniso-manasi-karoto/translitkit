package tha

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tassa-yoniso-manasi-karoto/translitkit/common"
)

func TestPyThaiNLPProvider(t *testing.T) {
	// Skip if not explicitly enabled
	if os.Getenv("PYTHAINLP_TEST") != "1" {
		t.Skip("PyThaiNLP integration tests disabled. Set PYTHAINLP_TEST=1 to run")
	}

	ctx := context.Background()

	t.Run("TokenizerMode", func(t *testing.T) {
		provider := &PyThaiNLPProvider{operatingMode: common.TokenizerType}
		
		// Initialize
		err := provider.InitWithContext(ctx)
		assert.NoError(t, err, "Failed to initialize PyThaiNLP tokenizer")
		defer provider.Close()
		
		// Test tokenization
		input := &TknSliceWrapper{
			TknSliceWrapper: common.TknSliceWrapper{
				Raw: []string{"สวัสดีครับ ผมชื่อโกโก้"},
			},
		}
		
		result, err := provider.ProcessFlowController(ctx, input)
		assert.NoError(t, err, "Failed to process text")
		
		// Check results
		assert.Greater(t, result.Len(), 0, "Expected tokens")
		
		// Verify tokenization
		tokens := result.(*TknSliceWrapper).Slice
		surfaces := make([]string, 0)
		for _, token := range tokens {
			if token.IsLexicalContent() {
				surfaces = append(surfaces, token.GetSurface())
			}
		}
		
		t.Logf("Tokenized: %v", surfaces)
		assert.Contains(t, surfaces, "สวัสดี")
		assert.Contains(t, surfaces, "ครับ")
		assert.Contains(t, surfaces, "ผม")
		assert.Contains(t, surfaces, "ชื่อ")
		assert.Contains(t, surfaces, "โกโก้")
	})

	t.Run("CombinedMode", func(t *testing.T) {
		provider := &PyThaiNLPProvider{operatingMode: common.CombinedType}
		
		// Initialize
		err := provider.InitWithContext(ctx)
		assert.NoError(t, err, "Failed to initialize PyThaiNLP combined")
		defer provider.Close()
		
		// Test tokenization + romanization
		input := &TknSliceWrapper{
			TknSliceWrapper: common.TknSliceWrapper{
				Raw: []string{"ภาษาไทยเป็นภาษาที่สวยงาม"},
			},
		}
		
		result, err := provider.ProcessFlowController(ctx, input)
		assert.NoError(t, err, "Failed to process text")
		
		// Check results
		assert.Greater(t, result.Len(), 0, "Expected tokens")
		
		// Verify romanization
		wrapper := result.(*TknSliceWrapper)
		romanized := wrapper.Roman()
		t.Logf("Romanized: %s", romanized)
		assert.NotEmpty(t, romanized, "Expected romanized text")
		
		// Check individual tokens
		hasRomanization := false
		for i := 0; i < wrapper.Len(); i++ {
			token := wrapper.GetIdx(i)
			if token.IsLexicalContent() && token.Roman() != "" {
				hasRomanization = true
				t.Logf("Token: %s → %s", token.GetSurface(), token.Roman())
			}
		}
		assert.True(t, hasRomanization, "Expected at least one token with romanization")
	})

	t.Run("ProgressCallback", func(t *testing.T) {
		provider := &PyThaiNLPProvider{operatingMode: common.TokenizerType}
		
		// Initialize
		err := provider.InitWithContext(ctx)
		assert.NoError(t, err)
		defer provider.Close()
		
		// Set up progress tracking
		progressCalled := false
		provider.WithProgressCallback(func(current, total int) {
			progressCalled = true
			t.Logf("Progress: %d/%d", current, total)
		})
		
		// Process multiple chunks
		input := &TknSliceWrapper{
			TknSliceWrapper: common.TknSliceWrapper{
				Raw: []string{"ทดสอบ", "การแบ่งข้อความ", "หลายส่วน"},
			},
		}
		
		_, err = provider.ProcessFlowController(ctx, input)
		assert.NoError(t, err)
		assert.True(t, progressCalled, "Progress callback should have been called")
	})

	t.Run("EmptyInput", func(t *testing.T) {
		provider := &PyThaiNLPProvider{operatingMode: common.TokenizerType}
		
		err := provider.InitWithContext(ctx)
		assert.NoError(t, err)
		defer provider.Close()
		
		// Test empty input
		input := &TknSliceWrapper{}
		_, err = provider.ProcessFlowController(ctx, input)
		assert.Error(t, err, "Expected error for empty input")
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		provider := &PyThaiNLPProvider{operatingMode: common.TokenizerType}
		
		err := provider.InitWithContext(ctx)
		assert.NoError(t, err)
		defer provider.Close()
		
		// Create cancellable context
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel() // Cancel immediately
		
		input := &TknSliceWrapper{
			TknSliceWrapper: common.TknSliceWrapper{
				Raw: []string{"ทดสอบ"},
			},
		}
		
		_, err = provider.ProcessFlowController(cancelCtx, input)
		assert.Error(t, err, "Expected error due to cancelled context")
	})
}

func TestProviderRegistration(t *testing.T) {
	// This test verifies that providers are properly registered
	// It doesn't require Docker/PyThaiNLP to be running
	
	// Test that we can create modules with different configurations
	testCases := []struct {
		name          string
		providerNames []string
		expectError   bool
	}{
		{
			name:          "PyThaiNLP Combined",
			providerNames: []string{"pythainlp"},
			expectError:   false,
		},
		{
			name:          "PyThaiNLP Tokenizer Only",
			providerNames: []string{"pythainlp-tokenizer"},
			expectError:   true, // Will error because it needs a transliterator
		},
		{
			name:          "Thai2English Combined",
			providerNames: []string{"thai2english.com"},
			expectError:   false,
		},
		{
			name:          "Hybrid Mode",
			providerNames: []string{"pythainlp-tokenizer", "thai2english.com"},
			expectError:   true, // Will error because thai2english is not registered as transliterator
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := common.NewModule("tha", tc.providerNames...)
			if tc.expectError {
				assert.Error(t, err, "Expected error for configuration: %v", tc.providerNames)
			} else {
				assert.NoError(t, err, "Expected success for configuration: %v", tc.providerNames)
			}
		})
	}
}

// BenchmarkPyThaiNLP provides performance comparison between modes
func BenchmarkPyThaiNLP(b *testing.B) {
	if os.Getenv("PYTHAINLP_TEST") != "1" {
		b.Skip("PyThaiNLP benchmarks disabled")
	}
	
	ctx := context.Background()
	testText := "ภาษาไทยเป็นภาษาที่มีเอกลักษณ์เฉพาะตัว มีระบบการเขียนและการออกเสียงที่แตกต่างจากภาษาอื่นๆ"
	
	b.Run("TokenizerMode", func(b *testing.B) {
		provider := &PyThaiNLPProvider{operatingMode: common.TokenizerType}
		provider.InitWithContext(ctx)
		defer provider.Close()
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			input := &TknSliceWrapper{
				TknSliceWrapper: common.TknSliceWrapper{
					Raw: []string{testText},
				},
			}
			provider.ProcessFlowController(ctx, input)
		}
	})
	
	b.Run("CombinedMode", func(b *testing.B) {
		provider := &PyThaiNLPProvider{operatingMode: common.CombinedType}
		provider.InitWithContext(ctx)
		defer provider.Close()
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			input := &TknSliceWrapper{
				TknSliceWrapper: common.TknSliceWrapper{
					Raw: []string{testText},
				},
			}
			provider.ProcessFlowController(ctx, input)
		}
	})
}