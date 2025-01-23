package executor

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSubmit(t *testing.T) {
	expect := "mock"
	future := Submit[string](func() (string, error) {
		return expect, nil
	})

	ret, _ := future.Join()
	assert.Equal(t, expect, ret)
}
