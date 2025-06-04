package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPtr(t *testing.T) {
	t.Parallel()

	t.Run("Integer pointer", func(t *testing.T) {
		t.Parallel()
		val := 42
		ptr := Ptr(val)
		assert.NotNil(t, ptr)
		assert.Equal(t, 42, *ptr)
	})

	t.Run("String pointer", func(t *testing.T) {
		t.Parallel()
		val := "hello"
		ptr := Ptr(val)
		assert.NotNil(t, ptr)
		assert.Equal(t, "hello", *ptr)
	})

	t.Run("Boolean pointer", func(t *testing.T) {
		t.Parallel()
		val := true
		ptr := Ptr(val)
		assert.NotNil(t, ptr)
		assert.True(t, *ptr)
	})

	t.Run("Struct pointer", func(t *testing.T) {
		t.Parallel()
		type sample struct {
			A int
			B string
		}
		val := sample{A: 5, B: "test"}
		ptr := Ptr(val)
		assert.NotNil(t, ptr)
		assert.Equal(t, 5, ptr.A)
		assert.Equal(t, "test", ptr.B)
	})
}
