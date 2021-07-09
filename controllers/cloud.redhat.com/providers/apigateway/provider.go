package apigateway

import (
	p "cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/providers"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
)

// ProvName sets the provider name identifier
var ProvName = "apigateway"

// CoreApiGateway is the cronapigateway for the apps apigateways.
var ApiGatewayConfig = p.NewSingleResourceIdent(ProvName, "apigateway_config", &core.ConfigMap{})

var ApiGatewayDeployment = p.NewSingleResourceIdent(ProvName, "apigateway_deployment", &apps.Deployment{})

var ApiGatewayService = p.NewSingleResourceIdent(ProvName, "apigateway_service", &core.Service{})

// GetEnd returns the correct end provider.
func GetApiGateway(c *p.Provider) (p.ClowderProvider, error) {
	return NewApiGatewayProvider(c)
}

func init() {
	p.ProvidersRegistration.Register(GetApiGateway, 4, ProvName)
}
