package resourcecodec

func streamRouteCodecConfig() resourceCodecConfig {
	return resourceCodecConfig{
		nameKey:           "name",
		associationFields: []string{"service_id", "upstream_id"},
		stripFields:       []string{"id", "name", "service_id", "upstream_id"},
	}
}
