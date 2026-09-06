package httpsec

type VerifyConfig struct {
	RequireHeader bool
}

type VerifyOption = func(opts *VerifyConfig)

func RequireHeader() VerifyOption {
	return func(cfg *VerifyConfig) {
		cfg.RequireHeader = true
	}
}
