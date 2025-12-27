package config

func NewGlobal() *Global {
	return &Global{}
}

type Global struct{}

func (c *Global) ServerPort() int {
	return GetInt("SERVER_PORT", 8080)
}

func (c *Global) BitbucketUrl() string {
	return GetString("BITBUCKET_URL", "")
}

func (c *Global) BitbucketEmail() string {
	return GetString("BITBUCKET_EMAIL", "")
}

func (c *Global) BitbucketApiToken() string {
	return GetString("BITBUCKET_API_TOKEN", "")
}

func (c *Global) BitbucketTimeout() int {
	return GetInt("BITBUCKET_TIMEOUT", 5)
}
