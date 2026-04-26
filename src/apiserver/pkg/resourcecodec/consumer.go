package resourcecodec

func consumerCodecConfig() resourceCodecConfig {
	return resourceCodecConfig{
		nameKey:           "username",
		associationFields: []string{"group_id"},
		stripFields:       []string{"id", "name", "username", "group_id"},
	}
}
