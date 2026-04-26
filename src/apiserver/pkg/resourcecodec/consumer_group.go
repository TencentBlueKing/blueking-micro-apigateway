package resourcecodec

func consumerGroupCodecConfig() resourceCodecConfig {
	return resourceCodecConfig{
		nameKey:     "name",
		stripFields: []string{"id", "name"},
	}
}
