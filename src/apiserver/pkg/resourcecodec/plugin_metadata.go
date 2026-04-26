package resourcecodec

func pluginMetadataCodecConfig() resourceCodecConfig {
	return resourceCodecConfig{
		nameKey:     "name",
		stripFields: []string{"id", "name"},
	}
}
