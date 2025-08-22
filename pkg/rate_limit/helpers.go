package rate_limit

// mergeConfig merges custom config with instance config
func (r *rateLimit) mergeConfig(customConfig *RateLimitConfig) *RateLimitConfig {
	merged := &RateLimitConfig{}

	// Copy from instance config first
	if r.config != nil {
		merged.Limit = r.config.Limit
		merged.Window = r.config.Window
		merged.KeyGenerator = r.config.KeyGenerator
		merged.Skip = r.config.Skip
		merged.Message = r.config.Message
		merged.OnRateLimited = r.config.OnRateLimited
	}

	// Override with custom config if provided
	if customConfig != nil {
		if customConfig.Limit > 0 {
			merged.Limit = customConfig.Limit
		}
		if customConfig.Window > 0 {
			merged.Window = customConfig.Window
		}
		if customConfig.KeyGenerator != nil {
			merged.KeyGenerator = customConfig.KeyGenerator
		}
		if customConfig.Skip != nil {
			merged.Skip = customConfig.Skip
		}
		if customConfig.Message != "" {
			merged.Message = customConfig.Message
		}
		if customConfig.OnRateLimited != nil {
			merged.OnRateLimited = customConfig.OnRateLimited
		}
	}

	return merged
}
