package config

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"
	grid "github.com/thegreenwebfoundation/grid-intensity-go"
	ci "github.com/thegreenwebfoundation/grid-intensity-go/carbonintensity"
)

// EnergyScoreProvider is the strategy that returns energy scores for the node.
type EnergyScoreProvider interface {
	GetCarbonIntensity(ctx context.Context) (float64, error)
}

const (
	AWS = "aws"
	GCP = "gcp"
	AZ  = "azure"
	EM  = "electricity-map"
	CI  = "carbon-intensity"
)

// EnergyConfig represents the possible configurations for energy scoring
// based off parsed client config.
type EnergyConfig struct {
	Region                string `hcl:"region"`
	ProviderKey           string `hcl:"provider"`
	ScoreProvider         *EnergyScoreProvider
	AWSConfig             *AWSConfig             `hcl:"aws"`
	GCPConfig             *GCPConfig             `hcl:"gcp"`
	AzureConfig           *AzureConfig           `hcl:"azure"`
	CarbonIntensityConfig *CarbonIntensityConfig `hcl:"carbon_intensity"`
	ElectricityMapConfig  *ElectricityMapConfig  `hcl:"electricity_map"`
}

func (ec *EnergyConfig) Copy() *EnergyConfig {
	if ec == nil {
		return nil
	}

	nec := &EnergyConfig{
		Region:      ec.Region,
		ProviderKey: ec.ProviderKey,
	}

	if ec.AWSConfig != nil {
		nec.AWSConfig = ec.AWSConfig.Copy()
	}

	if ec.GCPConfig != nil {
		nec.GCPConfig = ec.GCPConfig.Copy()
	}

	if ec.AzureConfig != nil {
		nec.AzureConfig = ec.AzureConfig.Copy()
	}

	if ec.CarbonIntensityConfig != nil {
		nec.CarbonIntensityConfig = ec.CarbonIntensityConfig.Copy()
	}

	if ec.ElectricityMapConfig != nil {
		nec.ElectricityMapConfig = ec.ElectricityMapConfig.Copy()
	}

	// set the ScoreProvider instance by calling finalize
	_ = nec.Finalize()

	return nec
}

func (ec *EnergyConfig) Validate() (err error) {
	if ec == nil {
		return fmt.Errorf("invalid energy config: config is nil")
	}

	switch ec.ProviderKey {
	case AWS:
		err = ec.AWSConfig.Validate()
	case GCP:
		err = ec.GCPConfig.Validate()
	case AZ:
		err = ec.AzureConfig.Validate()
	case CI:
		err = ec.CarbonIntensityConfig.Validate()
	case EM:
		err = ec.ElectricityMapConfig.Validate()
	default:
		err = fmt.Errorf("invalid energy config: provider %s not recognized", ec.ProviderKey)
	}

	return
}

// Finalize sets the provider instances based on the user specified configuration.
func (ec *EnergyConfig) Finalize() (err error) {
	if ec == nil {
		return
	}

	if err = ec.Validate(); err != nil {
		return
	}

	var factoryFn func(*EnergyConfig) (EnergyScoreProvider, error)
	switch ec.ProviderKey {
	case AWS:
		factoryFn = newAWSProvider
	case GCP:
		factoryFn = newGCPProvider
	case AZ:
		factoryFn = newAzureProvider
	case CI:
		factoryFn = newCIProvider
	case EM:
		factoryFn = newEMProvider
	}

	if factoryFn == nil {
		return
	}

	var provider EnergyScoreProvider
	provider, err = factoryFn(ec)
	if err != nil {
		return
	}

	ec.ScoreProvider = &provider
	return
}

type AWSConfig struct {
	AccessKeyID     string `hcl:"access_key_id"`
	SecretAccessKey string `hcl:"secret_access_key"`
	SessionToken    string `hcl:"session_token"`
}

func (aws *AWSConfig) Copy() *AWSConfig {
	if aws == nil {
		return nil
	}

	n := new(AWSConfig)
	*n = *aws

	return n
}

func (aws *AWSConfig) Validate() error {
	if aws == nil {
		return fmt.Errorf("invalid energy config: AWS specified but not configured")
	}

	var mErr multierror.Error

	if aws.AccessKeyID == "" {
		mErr.Errors = append(mErr.Errors, fmt.Errorf("invalid energy config: access_key_id required"))
	}

	if aws.SecretAccessKey == "" {
		mErr.Errors = append(mErr.Errors, fmt.Errorf("invalid energy config: secret_acces_key required"))
	}

	if aws.SessionToken == "" {
		mErr.Errors = append(mErr.Errors, fmt.Errorf("invalid energy config: session_token required"))
	}

	return mErr.ErrorOrNil()
}

func newAWSProvider(config *EnergyConfig) (EnergyScoreProvider, error) {
	return &awsProvider{
		config,
	}, nil
}

type awsProvider struct {
	config *EnergyConfig
}

func (aws *awsProvider) GetCarbonIntensity(ctx context.Context) (float64, error) {
	return 0, nil
}

type gcpProvider struct {
	config *EnergyConfig
}

type GCPConfig struct {
	ServiceAccountKey string `hcl:"service_account_key"`
}

func (gcp *GCPConfig) Copy() *GCPConfig {
	if gcp == nil {
		return nil
	}

	n := new(GCPConfig)
	*n = *gcp

	return n
}

func (gcp *GCPConfig) Validate() error {
	if gcp == nil {
		return fmt.Errorf("invalid energy config: GCP specified but not configured")
	}

	if gcp.ServiceAccountKey == "" {
		return fmt.Errorf("invalid energy config: service_account_key required")
	}

	return nil
}

func newGCPProvider(config *EnergyConfig) (EnergyScoreProvider, error) {
	return &gcpProvider{
		config,
	}, nil
}

func (gcp *gcpProvider) GetCarbonIntensity(ctx context.Context) (float64, error) {
	return 0, nil
}

type azureProvider struct {
	config *EnergyConfig
}

type AzureConfig struct {
	ClientID     string `hcl:"client_id"`
	ClientSecret string `hcl:"client_secret"`
	TenantID     string `hcl:"tenant_id"`
}

func (az *AzureConfig) Copy() *AzureConfig {
	if az == nil {
		return nil
	}

	n := new(AzureConfig)
	*n = *az

	return n
}

func (az *AzureConfig) Validate() error {
	if az == nil {
		return fmt.Errorf("invalid energy config: Azure specified but not configured")
	}

	var mErr multierror.Error

	if az.ClientID == "" {
		mErr.Errors = append(mErr.Errors, fmt.Errorf("invalid energy config: client_id required"))
	}

	if az.ClientSecret == "" {
		mErr.Errors = append(mErr.Errors, fmt.Errorf("invalid energy config: client_secret required"))
	}

	if az.TenantID == "" {
		mErr.Errors = append(mErr.Errors, fmt.Errorf("invalid energy config: tenant_id required"))
	}

	return mErr.ErrorOrNil()
}

func newAzureProvider(config *EnergyConfig) (EnergyScoreProvider, error) {
	return &azureProvider{
		config,
	}, nil
}

func (az *azureProvider) GetCarbonIntensity(ctx context.Context) (float64, error) {
	return 0, nil
}

type ciProvider struct {
	config   *EnergyConfig
	provider grid.Provider
}

type CarbonIntensityConfig struct {
	APIUrl string `hcl:"api_url"`
}

func (ci *CarbonIntensityConfig) Copy() *CarbonIntensityConfig {
	if ci == nil {
		return nil
	}

	n := new(CarbonIntensityConfig)
	*n = *ci

	return n
}

func (ci *CarbonIntensityConfig) Validate() error {
	if ci == nil {
		return fmt.Errorf("invalid energy config: Carbon Intensity specified but not configured")
	}

	if ci.APIUrl == "" {
		return fmt.Errorf("invalid energy config: api_url required")
	}

	return nil
}

func newCIProvider(config *EnergyConfig) (EnergyScoreProvider, error) {
	provider, err := ci.New(ci.WithAPIURL(config.CarbonIntensityConfig.APIUrl))
	if err != nil {
		return nil, err
	}

	return &ciProvider{
		config,
		provider,
	}, nil
}

func (ci *ciProvider) GetCarbonIntensity(ctx context.Context) (float64, error) {
	return ci.provider.GetCarbonIntensity(ctx, ci.config.Region)
}

type emProvider struct {
	config *ElectricityMapConfig
}

type ElectricityMapConfig struct {
	APIKey string `hcl:"api_key"`
	APIUrl string `hcl:"api_url"`
}

func (em *ElectricityMapConfig) Copy() *ElectricityMapConfig {
	if em == nil {
		return nil
	}

	n := new(ElectricityMapConfig)
	*n = *em

	return n
}

func (em *ElectricityMapConfig) Validate() error {
	if em == nil {
		return fmt.Errorf("invalid energy config: Electricity Maps specified but not configured")
	}

	var mErr multierror.Error

	if em.APIKey == "" {
		mErr.Errors = append(mErr.Errors, fmt.Errorf("invalid energy config: api_key required"))
	}

	if em.APIUrl == "" {
		mErr.Errors = append(mErr.Errors, fmt.Errorf("invalid energy config: api_url required"))
	}

	return mErr.ErrorOrNil()
}

func newEMProvider(config *EnergyConfig) (EnergyScoreProvider, error) {
	return &emProvider{
		config.ElectricityMapConfig,
	}, nil
}

func (em *emProvider) GetCarbonIntensity(ctx context.Context) (float64, error) {
	return 0, nil
}
