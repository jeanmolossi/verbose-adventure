package config

type Session struct {
	EncryptionKey string `envconfig:"ENCRYPTION_KEY" required:"true"`
	JWTSecret     string `envconfig:"JWT_SECRET"     required:"true"`
}
