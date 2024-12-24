package config

type Config interface {
	GetString(key, defaultValue string) string
	GetInt(key string, defaultValue int) (int, error)
	GetFloat64(key string, defaultValue float64) (float64, error)
	GetBool(key string, defaultValue bool) (bool, error)
}
