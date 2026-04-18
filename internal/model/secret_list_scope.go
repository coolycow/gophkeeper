package model

// SecretListScope задаёт, какие секреты возвращать в списке (активные, только корзина или все).
type SecretListScope int

const (
	// SecretListScopeUnspecified — по умолчанию как активные (основной список).
	SecretListScopeUnspecified SecretListScope = 0
	SecretListScopeActiveOnly  SecretListScope = 1
	SecretListScopeDeletedOnly SecretListScope = 2
	SecretListScopeAll         SecretListScope = 3
)
