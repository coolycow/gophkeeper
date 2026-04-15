package service

// GenerateSalt генерирует криптографически стойкую соль (32 байта)
func generateSalt(length int) ([]byte, error) {
	return generateRandom(length)
}
