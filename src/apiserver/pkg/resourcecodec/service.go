package resourcecodec

func serviceCodecConfig() resourceCodecConfig {
	return resourceCodecConfig{
		nameKey:           "name",
		associationFields: []string{"upstream_id"},
		stripFields:       []string{"id", "name", "upstream_id"},
	}
}
