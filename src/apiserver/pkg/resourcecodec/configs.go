package resourcecodec

func routeCodecConfig() resourceCodecConfig {
	return resourceCodecConfig{
		nameKey:           "name",
		associationFields: []string{"service_id", "upstream_id", "plugin_config_id"},
		stripFields:       []string{"id", "name", "service_id", "upstream_id", "plugin_config_id"},
	}
}

func serviceCodecConfig() resourceCodecConfig {
	return resourceCodecConfig{
		nameKey:           "name",
		associationFields: []string{"upstream_id"},
		stripFields:       []string{"id", "name", "upstream_id"},
	}
}

func upstreamCodecConfig() resourceCodecConfig {
	return resourceCodecConfig{
		nameKey:           "name",
		associationFields: []string{"tls.client_cert_id"},
		stripFields:       []string{"id", "name", "tls.client_cert_id"},
	}
}

func consumerCodecConfig() resourceCodecConfig {
	return resourceCodecConfig{
		nameKey:           "username",
		associationFields: []string{"group_id"},
		stripFields:       []string{"id", "name", "username", "group_id"},
	}
}

func consumerGroupCodecConfig() resourceCodecConfig {
	return resourceCodecConfig{
		nameKey:     "name",
		stripFields: []string{"id", "name"},
	}
}

func pluginConfigCodecConfig() resourceCodecConfig {
	return resourceCodecConfig{
		nameKey:     "name",
		stripFields: []string{"id", "name"},
	}
}

func globalRuleCodecConfig() resourceCodecConfig {
	return resourceCodecConfig{
		nameKey:     "name",
		stripFields: []string{"id", "name"},
	}
}

func pluginMetadataCodecConfig() resourceCodecConfig {
	return resourceCodecConfig{
		nameKey:     "name",
		stripFields: []string{"id", "name"},
	}
}

func protoCodecConfig() resourceCodecConfig {
	return resourceCodecConfig{
		nameKey:     "name",
		stripFields: []string{"id", "name"},
	}
}

func sslCodecConfig() resourceCodecConfig {
	return resourceCodecConfig{
		nameKey:     "name",
		stripFields: []string{"id", "name"},
	}
}

func streamRouteCodecConfig() resourceCodecConfig {
	return resourceCodecConfig{
		nameKey:           "name",
		associationFields: []string{"service_id", "upstream_id"},
		stripFields:       []string{"id", "name", "service_id", "upstream_id"},
	}
}
