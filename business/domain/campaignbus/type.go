package campaignbus

import "fmt"

type CampaignType struct {
	value string
}

func (t CampaignType) String() string {
	return t.value
}

var campaignTypes = make(map[string]CampaignType)

var (
	EmailCampaign = newCampaignType("EMAIL")
	MaxCampaign   = newCampaignType("MAX")
)

func newCampaignType(value string) CampaignType {
	t := CampaignType{value: value}
	campaignTypes[value] = t
	return t
}

func ParseCampaignType(value string) (CampaignType, error) {
	t, ok := campaignTypes[value]
	if !ok {
		return CampaignType{}, fmt.Errorf("invalid campaign type: '%s'", value)
	}
	return t, nil
}
