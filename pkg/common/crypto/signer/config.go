package signer

// Config represents signer configuration
type Config struct {
	// KeystorePath is the path to the keystore file
	KeystorePath string
	// Password is the password to decrypt the keystore
	Password string
}

// IsValid checks if the config is valid
func (c *Config) IsValid() bool {
	return c.KeystorePath != "" && c.Password != ""
}
