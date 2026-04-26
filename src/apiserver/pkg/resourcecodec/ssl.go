package resourcecodec

func sslCodecConfig() resourceCodecConfig {
	return resourceCodecConfig{
		nameKey:     "name",
		stripFields: []string{"id", "name"},
	}
}
