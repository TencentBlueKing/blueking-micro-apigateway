package resourcecodec

func routeCodecConfig() resourceCodecConfig {
	return resourceCodecConfig{
		nameKey:           "name",
		associationFields: []string{"service_id", "upstream_id", "plugin_config_id"},
		stripFields:       []string{"id", "name", "service_id", "upstream_id", "plugin_config_id"},
	}
}
