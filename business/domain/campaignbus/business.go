package campaignbus

type CampaignBusiness struct {
	campaignStorer  CampaignStorer
	targetStorer    TargetStorer
	messageProvider MessageQuerier
	landingProvider LandingQuerier
	varsValidator   VarsValidator
}

type TargetBusiness struct {
	campaignStorer CampaignStorer
	targetStorer   TargetStorer
}

func NewCampaignBusiness(campaignStorer CampaignStorer, targetStorer TargetStorer, messageProvider MessageQuerier, landingProvider LandingQuerier, varsValidator VarsValidator) *CampaignBusiness {
	return &CampaignBusiness{campaignStorer: campaignStorer, targetStorer: targetStorer, messageProvider: messageProvider, landingProvider: landingProvider, varsValidator: varsValidator}
}

func NewTargetBusiness(campaignStorer CampaignStorer, targetStorer TargetStorer) *TargetBusiness {
	return &TargetBusiness{campaignStorer: campaignStorer, targetStorer: targetStorer}
}
