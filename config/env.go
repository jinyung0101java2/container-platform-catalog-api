package config

import (
	"github.com/gofiber/fiber/v2/log"
	"github.com/spf13/viper"
	"helm.sh/helm/v3/pkg/repo"
	"os"
)

var Env *envConfigs

func InitEnvConfigs() {
	Env = loadEnvVariables()
	makeRepoConfig()
}

type envConfigs struct {
	ServerPort                string `mapstructure:"SERVER_PORT"`
	JwtSecret                 string `mapstructure:"JWT_SECRET"`
	HelmRepoConfig            string `mapstructure:"HELM_REPO_CONFIG"`
	HelmRepoCache             string `mapstructure:"HELM_REPO_CACHE"`
	HelmRepoCA                string `mapstructure:"HELM_REPO_CA"`
	ArtifactHubUrl            string `mapstructure:"ARTIFACT_HUB_API_URL"`
	ArtifactHubRepoSearch     string `mapstructure:"ARTIFACT_HUB_REPO_SEARCH"`
	ArtifactHubPackageSearch  string `mapstructure:"ARTIFACT_HUB_PACKAGE_SEARCH"`
	ArtifactHubPackageDetail  string `mapstructure:"ARTIFACT_HUB_PACKAGE_DETAIL"`
	ArtifactHubPackageValues  string `mapstructure:"ARTIFACT_HUB_PACKAGE_VALUES"`
	ArtifactHubPackageLogoUrl string `mapstructure:"ARTIFACT_HUB_PACKAGE_LOGO_URL"`
	VaultUrl                  string `mapstructure:"VAULT_URL"`
	VaultRoleName             string `mapstructure:"VAULT_ROLE_NAME"`
	VaultRoleId               string `mapstructure:"VAULT_ROLE_ID"`
	VaultSecretId             string `mapstructure:"VAULT_SECRET_ID"`
	VaultClusterPath          string `mapstructure:"VAULT_CLUSTER_PATH"`
	VaultUserPath             string `mapstructure:"VAULT_USER_PATH"`
}

func loadEnvVariables() (config *envConfigs) {
	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	viper.SetConfigType("env")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal("Error reading env file", err)
	}
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatal(err)
	}
	return
}

func makeRepoConfig() {
	// Check repositories.yaml exists
	if _, err := os.Stat(Env.HelmRepoConfig); os.IsNotExist(err) {
		log.Infof("[FILE NOT FOUND] PATH:: %v...", Env.HelmRepoConfig)
		repositories := repo.NewFile()
		log.Infof("[CREATING NEW FILE] PATH:: %v...", Env.HelmRepoConfig)
		if err = repositories.WriteFile(Env.HelmRepoConfig, 0600); err != nil {
			log.Errorf("[FAILED TO CREATE FILE] PATH:: %v, ERR:: %v", Env.HelmRepoConfig, err)
		}
	}

	// Check repository cache path exists
	if err := os.MkdirAll(Env.HelmRepoCache, os.ModePerm); err != nil {
		log.Errorf("[FAILED TO CREATE CACHE DIR] PATH:: %v, ERR:: %v", Env.HelmRepoCache, err)
	}

	// Check repository ca file path exists
	if err := os.MkdirAll(Env.HelmRepoCA, os.ModePerm); err != nil {
		log.Errorf("[FAILED TO CREATE CERT DIR] PATH:: %v, ERR:: %v", Env.HelmRepoCA, err)
	}
}
