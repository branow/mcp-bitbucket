package config

func McpServerPort() int {
	return GetInt("SERVER_PORT", 8080)
}

func BitBucketEmail() string {
	return GetString("BITBUCKET_EMAIL", "")
}

func BitBucketApiToken() string {
	return GetString("BITBUCKET_API_TOKEN", "")
}
