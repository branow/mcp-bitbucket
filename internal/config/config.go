package config

func McpServerPort() int {
	return GetInt("SERVER_PORT", 8080)
}

func BitBucketUrl() string {
	return GetString("BITBUCKET_URL", "")
}

func BitBucketNamespace() string {
	return GetString("BITBUCKET_NAMESPACE", "")
}

func BitBucketEmail() string {
	return GetString("BITBUCKET_EMAIL", "")
}

func BitBucketApiToken() string {
	return GetString("BITBUCKET_API_TOKEN", "")
}

func BitBucketTimeout() int {
	return GetInt("BITBUCKET_TIMEOUT", 5)
}
