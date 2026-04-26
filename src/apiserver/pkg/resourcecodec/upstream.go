package resourcecodec

func upstreamCodecConfig() resourceCodecConfig {
	return resourceCodecConfig{
		nameKey:           "name",
		associationFields: []string{"tls.client_cert_id"},
		stripFields:       []string{"id", "name", "tls.client_cert_id"},
	}
}
