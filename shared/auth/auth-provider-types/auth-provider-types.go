package authProviderTypes

type AuthProviderType string

const (
	GitHub AuthProviderType = "github"
)

func (e *AuthProviderType) String() string {
	return string(*e)
}

func (e *AuthProviderType) Set(value string) error {
	*e = AuthProviderType(value)
	return nil
}

func (e *AuthProviderType) Type() string {
	return "AuthProviderType"
}
