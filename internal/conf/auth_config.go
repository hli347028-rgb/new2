package conf

import "time"

// AuthConfig holds Web3 auth settings loaded from config.yaml.
type AuthConfig struct {
	JwtSecret          string   `json:"jwt_secret" yaml:"jwt_secret"`
	BootstrapAddresses []string `json:"bootstrap_addresses" yaml:"bootstrap_addresses"`
	AdminAddresses     []string `json:"admin_addresses" yaml:"admin_addresses"`
	AdminAccount       string   `json:"admin_account" yaml:"admin_account"`
	AdminPassword      string   `json:"admin_password" yaml:"admin_password"`
	ChallengeTTL       string   `json:"challenge_ttl" yaml:"challenge_ttl"`
}

func (a *AuthConfig) GetAdminAddresses() []string {
	if a == nil {
		return nil
	}
	return a.AdminAddresses
}

func (a *AuthConfig) GetBootstrapAddresses() []string {
	if a == nil {
		return nil
	}
	return a.BootstrapAddresses
}

func (a *AuthConfig) ChallengeDuration() time.Duration {
	if a == nil || a.ChallengeTTL == "" {
		return 5 * time.Minute
	}
	d, err := time.ParseDuration(a.ChallengeTTL)
	if err != nil {
		return 5 * time.Minute
	}
	return d
}

func (a *AuthConfig) GetJwtSecret() string {
	if a == nil || a.JwtSecret == "" {
		return "change-me-in-production"
	}
	return a.JwtSecret
}

func (a *AuthConfig) GetAdminAccount() string {
	if a == nil || a.AdminAccount == "" {
		return "admin"
	}
	return a.AdminAccount
}

func (a *AuthConfig) GetAdminPassword() string {
	if a == nil || a.AdminPassword == "" {
		return "admin123"
	}
	return a.AdminPassword
}
