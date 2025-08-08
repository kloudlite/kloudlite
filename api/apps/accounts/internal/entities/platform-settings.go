package entities

import (
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/repos"
)

type OAuthProvider struct {
	Enabled      bool   `json:"enabled"`
	ClientId     string `json:"clientId,omitempty"`
	ClientSecret string `json:"clientSecret,omitempty"`
	// Additional provider-specific settings
	TenantId     string `json:"tenantId,omitempty"` // For Azure AD
}

type PlatformSettings struct {
	repos.BaseEntity `json:",inline" graphql:"noinput"`
	common.ResourceMetadata `json:",inline"`
	
	// Platform identity
	PlatformName        string `json:"platformName"`
	PlatformOwnerEmail  string `json:"platformOwnerEmail"`
	SupportEmail        string `json:"supportEmail"`
	
	// Authentication settings
	AllowSignup         bool   `json:"allowSignup"`
	RequireEmailVerification bool `json:"requireEmailVerification"`
	
	// OAuth providers
	OAuthProviders struct {
		Google    OAuthProvider `json:"google"`
		GitHub    OAuthProvider `json:"github"`
		Microsoft OAuthProvider `json:"microsoft"`
	} `json:"oauthProviders"`
	
	// Email settings
	EmailSettings struct {
		Enabled       bool   `json:"enabled"`
		SMTPHost      string `json:"smtpHost,omitempty"`
		SMTPPort      int    `json:"smtpPort,omitempty"`
		SMTPUsername  string `json:"smtpUsername,omitempty"`
		SMTPPassword  string `json:"smtpPassword,omitempty"`
		FromEmail     string `json:"fromEmail,omitempty"`
		FromName      string `json:"fromName,omitempty"`
		// For services like Mailtrap, SendGrid, etc
		APIKey        string `json:"apiKey,omitempty"`
		ServiceType   string `json:"serviceType,omitempty"` // "smtp", "mailtrap", "sendgrid"
	} `json:"emailSettings"`
	
	// Team settings
	TeamSettings struct {
		RequireApproval      bool `json:"requireApproval"`
		AutoApproveFirstTeam bool `json:"autoApproveFirstTeam"`
		MaxTeamsPerUser      int  `json:"maxTeamsPerUser"`
	} `json:"teamSettings"`
	
	// Platform features
	Features struct {
		EnableDeviceFlow bool `json:"enableDeviceFlow"`
		EnableCLI        bool `json:"enableCLI"`
		EnableAPI        bool `json:"enableAPI"`
	} `json:"features"`
	
	// Cloud provider (only one can be active)
	CloudProvider struct {
		Provider string `json:"provider,omitempty"` // "aws", "gcp", "azure", "digitalocean"
		AWS struct {
			AccessKeyId     string `json:"accessKeyId,omitempty"`
			SecretAccessKey string `json:"secretAccessKey,omitempty"`
			Region          string `json:"region,omitempty"`
		} `json:"aws,omitempty"`
		GCP struct {
			ProjectId         string `json:"projectId,omitempty"`
			ServiceAccountKey string `json:"serviceAccountKey,omitempty"`
		} `json:"gcp,omitempty"`
		Azure struct {
			SubscriptionId string `json:"subscriptionId,omitempty"`
			TenantId       string `json:"tenantId,omitempty"`
			ClientId       string `json:"clientId,omitempty"`
			ClientSecret   string `json:"clientSecret,omitempty"`
		} `json:"azure,omitempty"`
		DigitalOcean struct {
			Token string `json:"token,omitempty"`
		} `json:"digitalocean,omitempty"`
	} `json:"cloudProvider"`
	
	// Platform status
	IsInitialized bool `json:"isInitialized"`
}

var PlatformSettingsIndices = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}

// Default platform settings
func DefaultPlatformSettings() *PlatformSettings {
	settings := &PlatformSettings{
		PlatformName: "Kloudlite",
		AllowSignup: false, // Disabled by default - only admin can sign up initially
		RequireEmailVerification: true,
		OAuthProviders: struct {
			Google    OAuthProvider `json:"google"`
			GitHub    OAuthProvider `json:"github"`
			Microsoft OAuthProvider `json:"microsoft"`
		}{
			Google:    OAuthProvider{Enabled: false},
			GitHub:    OAuthProvider{Enabled: false},
			Microsoft: OAuthProvider{Enabled: false},
		},
		TeamSettings: struct {
			RequireApproval      bool `json:"requireApproval"`
			AutoApproveFirstTeam bool `json:"autoApproveFirstTeam"`
			MaxTeamsPerUser      int  `json:"maxTeamsPerUser"`
		}{
			RequireApproval:      true,
			AutoApproveFirstTeam: true,
			MaxTeamsPerUser:      10,
		},
		Features: struct {
			EnableDeviceFlow bool `json:"enableDeviceFlow"`
			EnableCLI        bool `json:"enableCLI"`
			EnableAPI        bool `json:"enableAPI"`
		}{
			EnableDeviceFlow: true,
			EnableCLI:        true,
			EnableAPI:        true,
		},
		CloudProvider: struct {
			Provider string `json:"provider,omitempty"`
			AWS struct {
				AccessKeyId     string `json:"accessKeyId,omitempty"`
				SecretAccessKey string `json:"secretAccessKey,omitempty"`
				Region          string `json:"region,omitempty"`
			} `json:"aws,omitempty"`
			GCP struct {
				ProjectId         string `json:"projectId,omitempty"`
				ServiceAccountKey string `json:"serviceAccountKey,omitempty"`
			} `json:"gcp,omitempty"`
			Azure struct {
				SubscriptionId string `json:"subscriptionId,omitempty"`
				TenantId       string `json:"tenantId,omitempty"`
				ClientId       string `json:"clientId,omitempty"`
				ClientSecret   string `json:"clientSecret,omitempty"`
			} `json:"azure,omitempty"`
			DigitalOcean struct {
				Token string `json:"token,omitempty"`
			} `json:"digitalocean,omitempty"`
		}{},
		IsInitialized: false,
	}
	settings.Id = "platform-settings"
	return settings
}