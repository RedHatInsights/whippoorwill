package database

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"strings"

	crd "cloud.redhat.com/clowder/v2/apis/cloud.redhat.com/v1alpha1"
	"cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/config"
	"cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/errors"
	"cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/providers"
	core "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var rdsCa string

const caURL string = "https://s3.amazonaws.com/rds-downloads/rds-combined-ca-bundle.pem"

type appInterface struct {
	providers.Provider
	Config config.DatabaseConfig
}

func fetchCa() (string, error) {
	resp, err := http.Get(caURL)

	if err != nil {
		return "", errors.Wrap("Error fetching CA bundle", err)
	}

	if resp.StatusCode != 200 {
		msg := fmt.Sprintf("Bad status code: %d", resp.StatusCode)
		return "", errors.New(msg)
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		errors.Wrap("Error reading response body", err)
	}

	caBundle := string(body)

	if !strings.HasPrefix(caBundle, "-----BEGIN CERTIFICATE") {
		return "", errors.New("Invalid RDS CA bundle")
	}

	return caBundle, nil
}

// NewAppInterfaceDBProvider creates a new app-interface DB provider obejct.
func NewAppInterfaceDBProvider(p *providers.Provider) (providers.ClowderProvider, error) {
	provider := appInterface{Provider: *p}

	if rdsCa == "" {
		_rdsCa, err := fetchCa()

		if err != nil {
			return nil, errors.Wrap("Failed to fetch RDS CA bundle", err)
		}

		rdsCa = _rdsCa
	}
	return &provider, nil
}

func (a *appInterface) Provide(app *crd.ClowdApp, c *config.AppConfig) error {
	if app.Spec.Database.Name == "" && app.Spec.Database.SharedDBAppName == "" {
		return nil
	}

	if app.Spec.Database.Name != "" && app.Spec.Database.SharedDBAppName != "" {
		return errors.New("Cannot set dbName & shared db app name")
	}

	var dbSpec crd.DatabaseSpec
	var namespace string

	if app.Spec.Database.Name != "" {
		dbSpec = app.Spec.Database
		namespace = app.Namespace
	} else if app.Spec.Database.SharedDBAppName != "" {
		err := checkDependency(app)
		if err != nil {
			return err
		}

		refApp, err := crd.GetAppForDBInSameEnv(a.Ctx, a.Client, app)

		if err != nil {
			return err
		}

		dbSpec = refApp.Spec.Database
		namespace = refApp.Namespace

	}

	secrets := core.SecretList{}
	err := a.Client.List(a.Ctx, &secrets, client.InNamespace(namespace))

	if err != nil {
		msg := fmt.Sprintf("Failed to list secrets in %s", namespace)
		return errors.Wrap(msg, err)
	}

	sort.Slice(secrets.Items, func(i, j int) bool {
		return secrets.Items[i].Name < secrets.Items[j].Name
	})

	var matched config.DatabaseConfig

	matches, err := searchAnnotationSecret(app.Name, secrets.Items)

	if err != nil {
		return errors.Wrap("failed to extract annotated secret", err)
	}

	if len(matches) == 0 {

		dbConfigs, err := genDbConfigs(secrets.Items)

		if err != nil {
			return err
		}

		matched = resolveDb(dbSpec, dbConfigs)

		if matched == (config.DatabaseConfig{}) {
			return &errors.MissingDependencies{
				MissingDeps: map[string][]string{
					"database": {app.Name},
				},
			}
		}
	} else {
		matched = matches[0]
	}

	// The creds given by app-interface have elevated privileges
	matched.AdminPassword = matched.Password
	matched.AdminUsername = matched.Username
	matched.RdsCa = &rdsCa

	a.Config = matched

	c.Database = &a.Config

	return nil
}

func resolveDb(spec crd.DatabaseSpec, c []config.DatabaseConfig) config.DatabaseConfig {
	for _, config := range c {
		hostname := strings.Split(config.Hostname, ".")[0]
		nameSegments := strings.Split(hostname, "-")
		segLen := len(nameSegments)
		lastSegment := nameSegments[segLen-1]

		if lastSegment == "readonly" {
			continue
		}

		dbName := strings.Join(nameSegments[:segLen-1], "-")

		if dbName == spec.Name {
			return config
		}
	}

	return config.DatabaseConfig{}
}

func genDbConfigs(secrets []core.Secret) ([]config.DatabaseConfig, error) {
	configs := []config.DatabaseConfig{}

	var err error

	extractFn := func(m map[string][]byte) {
		port, erro := strconv.Atoi(string(m["db.port"]))

		if erro != nil {
			err = errors.Wrap("Failed to parse DB port", err)
			return
		}

		dbConfig := config.DatabaseConfig{
			Hostname: string(m["db.host"]),
			Port:     port,
			Username: string(m["db.user"]),
			Password: string(m["db.password"]),
			Name:     string(m["db.name"]),
			SslMode:  "verify-full",
		}

		configs = append(configs, dbConfig)
	}

	keys := []string{"db.host", "db.port", "db.user", "db.password", "db.name"}
	providers.ExtractSecretData(secrets, extractFn, keys...)

	if err != nil {
		return nil, err
	}

	return configs, nil
}

func searchAnnotationSecret(appName string, secrets []core.Secret) ([]config.DatabaseConfig, error) {
	for _, secret := range secrets {
		anno := secret.GetAnnotations()
		if v, ok := anno["database"]; ok && v == appName {
			configs, err := genDbConfigs([]core.Secret{secret})
			return configs, err
		}
	}
	return []config.DatabaseConfig{}, nil
}
