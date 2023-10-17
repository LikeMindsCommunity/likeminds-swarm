package externalHelpers

type CommunityConfiguration struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Value       map[string]interface{} `json:"value"`
}

type ExternalEntities struct {
	CommunityConfigurations []CommunityConfiguration
}
