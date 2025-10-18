package config

func McpServerPort() int {
	return GetInt("SERVER_PORT", 8080)
}
