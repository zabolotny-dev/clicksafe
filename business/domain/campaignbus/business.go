package campaignbus

type CampaignBusiness struct {
	campaignStorer     CampaignStorer
	targetStorer       TargetStorer
	messageProvider    MessageQuerier
	landingProvider    LandingQuerier
	employeeProvider   EmployeeQuerier
	varsValidator      VarsValidator
	attachmentProvider AttachmentProvider
}

type TargetBusiness struct {
	campaignStorer CampaignStorer
	targetStorer   TargetStorer
}

func NewCampaignBusiness(campaignStorer CampaignStorer, targetStorer TargetStorer, messageProvider MessageQuerier, landingProvider LandingQuerier, employeeProvider EmployeeQuerier, varsValidator VarsValidator, attachmentProvider AttachmentProvider) *CampaignBusiness {
	return &CampaignBusiness{campaignStorer: campaignStorer, targetStorer: targetStorer, messageProvider: messageProvider, landingProvider: landingProvider, employeeProvider: employeeProvider, varsValidator: varsValidator, attachmentProvider: attachmentProvider}
}

func NewTargetBusiness(campaignStorer CampaignStorer, targetStorer TargetStorer) *TargetBusiness {
	return &TargetBusiness{campaignStorer: campaignStorer, targetStorer: targetStorer}
}
