package dependencies

import (
	"fmt"

	crd "cloud.redhat.com/clowder/v2/apis/cloud.redhat.com/v1alpha1"
	"cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/config"
	"cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/errors"
)

func GetDeploymentName(app *crd.ClowdApp, deployment *crd.Deployment) string {
	return fmt.Sprintf("%s-%s", app.Name, deployment.Name)
}

func (dep *dependenciesProvider) makeDependencies(app *crd.ClowdApp, c *config.AppConfig) error {

	if dep.Provider.Env.Spec.Providers.Web.PrivatePort == 0 {
		dep.Provider.Env.Spec.Providers.Web.PrivatePort = 10000
	}

	depConfig := []config.DependencyEndpoint{}
	privDepConfig := []config.PrivateDependencyEndpoint{}

	processAppEndpoints(
		map[string]crd.ClowdApp{app.Name: *app},
		[]string{app.Name},
		&depConfig,
		&privDepConfig,
		dep.Provider.Env.Spec.Providers.Web.Port,
		dep.Provider.Env.Spec.Providers.Web.PrivatePort,
	)

	// Return if no deps

	deps := app.Spec.Dependencies
	odeps := app.Spec.OptionalDependencies
	if len(deps) == 0 && len(odeps) == 0 {
		c.Endpoints = depConfig
		c.PrivateEndpoints = privDepConfig
		return nil
	}

	// Get all ClowdApps

	apps := crd.ClowdAppList{}
	err := dep.Client.List(dep.Ctx, &apps)

	if err != nil {
		return errors.Wrap("Failed to list apps", err)
	}

	// Iterate over all deps
	missingDeps := makeDepConfig(
		&depConfig,
		&privDepConfig,
		dep.Provider.Env.Spec.Providers.Web.Port,
		dep.Provider.Env.Spec.Providers.Web.PrivatePort,
		app,
		&apps,
	)

	if len(missingDeps) > 0 {
		depVal := map[string][]string{"services": missingDeps}
		return &errors.MissingDependencies{MissingDeps: depVal}
	}

	c.Endpoints = depConfig
	c.PrivateEndpoints = privDepConfig
	return nil
}

func makeDepConfig(
	depConfig *[]config.DependencyEndpoint,
	privDepConfig *[]config.PrivateDependencyEndpoint,
	webPort int32,
	privatePort int32,
	app *crd.ClowdApp,
	apps *crd.ClowdAppList,
) (missingDeps []string) {

	appMap := map[string]crd.ClowdApp{}

	for _, iapp := range apps.Items {
		if iapp.Spec.EnvName == app.Spec.EnvName {
			appMap[iapp.Name] = iapp
		}
	}

	missingDeps = processAppEndpoints(appMap, app.Spec.Dependencies, depConfig, privDepConfig, webPort, privatePort)
	_ = processAppEndpoints(appMap, app.Spec.OptionalDependencies, depConfig, privDepConfig, webPort, privatePort)

	return missingDeps
}

func processAppEndpoints(
	appMap map[string]crd.ClowdApp,
	depList []string,
	depConfig *[]config.DependencyEndpoint,
	privDepConfig *[]config.PrivateDependencyEndpoint,
	webPort int32,
	privatePort int32,
) (missingDeps []string) {

	missingDeps = []string{}

	for _, dep := range depList {
		depApp, exists := appMap[dep]
		if !exists {
			missingDeps = append(missingDeps, dep)
			continue
		}

		// If app has public endpoint, add it to app config

		for _, deployment := range depApp.Spec.Deployments {
			if deployment.WebServices.Public.Enabled {
				name := fmt.Sprintf("%s-%s", depApp.Name, deployment.Name)
				*depConfig = append(*depConfig, config.DependencyEndpoint{
					Hostname: fmt.Sprintf("%s.%s.svc", name, depApp.Namespace),
					Port:     int(webPort),
					Name:     deployment.Name,
					App:      depApp.Name,
				})
			}
			if deployment.WebServices.Private.Enabled {
				name := fmt.Sprintf("%s-%s", depApp.Name, deployment.Name)
				*privDepConfig = append(*privDepConfig, config.PrivateDependencyEndpoint{
					Hostname: fmt.Sprintf("%s.%s.svc", name, depApp.Namespace),
					Port:     int(privatePort),
					Name:     deployment.Name,
					App:      depApp.Name,
				})
			}
		}
	}

	return missingDeps
}
