package campaignbus

type CampaignBusiness struct {
	campaignStorer CampaignStorer
	targetStorer   TargetStorer
}

type TargetBusiness struct {
	campaignStorer CampaignStorer
	targetStorer   TargetStorer
}

func NewCampaignBusiness(campaignStorer CampaignStorer, targetStorer TargetStorer) *CampaignBusiness {
	return &CampaignBusiness{campaignStorer: campaignStorer, targetStorer: targetStorer}
}

func NewTargetBusiness(campaignStorer CampaignStorer, targetStorer TargetStorer) *TargetBusiness {
	return &TargetBusiness{campaignStorer: campaignStorer, targetStorer: targetStorer}
}
