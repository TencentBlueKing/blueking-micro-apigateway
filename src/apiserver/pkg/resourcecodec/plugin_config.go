package resourcecodec

func pluginConfigCodecConfig() resourceCodecConfig {
	return resourceCodecConfig{
		nameKey:     "name",
		stripFields: []string{"id", "name"},
	}
}
