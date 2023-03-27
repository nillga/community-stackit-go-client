package serviceaccounts

import (
	"github.com/SchwarzIT/community-stackit-go-client/pkg/contracts"
	"github.com/SchwarzIT/community-stackit-go-client/pkg/env"
)

var BaseURLs = env.URLs(
	"service_accounts",
	"https://api.stackit.cloud/service-account/",
	"https://api-qa.stackit.cloud/service-account/",
	"https://api-dev.stackit.cloud/service-account/",
)

func NewService(c contracts.BaseClientInterface) *ClientWithResponses {
	return NewClient(BaseURLs.GetURL(c.GetEnvironment()), c)
}
