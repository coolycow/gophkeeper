package grpcserver

import (
	"testing"

	"github.com/coolycow/gophkeeper/internal/proto/gophkeeperpb"
	"github.com/stretchr/testify/assert"
)

// Публичные RPC не требуют JWT; остальные — защищены (проверка в других тестах / интеграции).
func TestIsPublicGRPCMethod(t *testing.T) {
	assert.True(t, isPublicGRPCMethod(gophkeeperpb.GophKeeperService_Ping_FullMethodName))
	assert.True(t, isPublicGRPCMethod(gophkeeperpb.GophKeeperService_Register_FullMethodName))
	assert.True(t, isPublicGRPCMethod(gophkeeperpb.GophKeeperService_Login_FullMethodName))
	assert.True(t, isPublicGRPCMethod(gophkeeperpb.GophKeeperService_RefreshToken_FullMethodName))
	assert.False(t, isPublicGRPCMethod("/gophkeeper.v1.GophKeeperService/GetSecret"))
}
