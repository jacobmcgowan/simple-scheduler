package options

import authProviderTypes "github.com/jacobmcgowan/simple-scheduler/shared/auth/auth-provider-types"

type LoginOptions struct {
	ProviderType authProviderTypes.AuthProviderType
	ClientId     string
	ClientSecret string
}
