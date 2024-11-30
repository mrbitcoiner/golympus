package util

import (
	"os"
	"strconv"
)

// MustEnv returns the environment variable, or panic if it is not set
func MustEnv(envKey string) string {
	v, ok := os.LookupEnv(envKey)
	if !ok {
		panic("undefined env: " + envKey)
	}
	return v
}

// MustInt64Env returns the environment variable as int64, or panic if it is
// not set or results in parse error
func MustInt64Env(envKey string) int64 {
	v := MustEnv(envKey)
	intV, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		panic(ErrWrap("failed to parse env "+envKey, err))
	}
	return intV
}

// EnvOrDefault return the environment variable that matches the key or the
// default value when env is undefined
func EnvOrDefault(envKey string, defaultValue string) string {
	val, ok := os.LookupEnv(envKey)
	if !ok {
		return defaultValue
	}
	return val
}
