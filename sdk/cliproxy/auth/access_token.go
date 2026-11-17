package auth

func (a *Auth) CurrentAccessToken() string {
	return authAccessToken(a)
}
