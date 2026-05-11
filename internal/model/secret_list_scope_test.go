package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Константы scope должны быть различимы друг от друга.
func TestSecretListScopeDistinct(t *testing.T) {
	vals := map[SecretListScope]bool{
		SecretListScopeUnspecified: false,
		SecretListScopeActiveOnly:  false,
		SecretListScopeDeletedOnly: false,
		SecretListScopeAll:         false,
	}
	assert.Len(t, vals, 4)
}
