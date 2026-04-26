package resourcecodec

func globalRuleCodecConfig() resourceCodecConfig {
	return resourceCodecConfig{
		nameKey:     "name",
		stripFields: []string{"id", "name"},
	}
}
