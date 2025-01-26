package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_envReader_readEnv(t *testing.T) {

	{
		os.Setenv("APP_TEST1", "test1-value")

		test1 := "qwe"

		r := NewEnvReader()
		r.String(&test1, "test1", nil)
		assert.Equal(t, test1, "test1-value")
	}
	{
		file := os.TempDir() + "/test2"

		os.WriteFile(file, []byte("test2-value"), 0600)

		os.Setenv("APP_TEST2_FILE", file)

		test2 := "qwe"

		r := NewEnvReader()
		r.String(&test2, "test2", nil)
		assert.Equal(t, test2, "test2-value")
		assert.Equal(t, r.envError, nil)
	}
	{
		os.Setenv("APP_TEST1", "123")

		test1 := 0

		r := NewEnvReader()
		r.Int(&test1, "test1", nil)
		assert.Equal(t, test1, 123)
	}

	{
		os.Setenv("APP_TEST1", "true")

		test1 := false

		r := NewEnvReader()
		r.Bool(&test1, "test1", nil)
		assert.Equal(t, test1, true)
	}
}
