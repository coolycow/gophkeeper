package clientdata

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// Kind идентифицирует, какие поля имеют смысл в Payload.
type Kind string

const (
	// KindLoginPair это запись логина/пароля (и опционально URL).
	KindLoginPair Kind = "login_pair"
	// KindText это произвольный UTF-8 текст.
	KindText Kind = "text"
	// KindBinary хранит маленькие бинарные блоки как base64 в JSON (не идеально для больших файлов; вложения отдельные).
	KindBinary Kind = "binary"
	// KindBankCard хранит cardholder, number, expiry, CVC как простые поля перед шифрованием.
	KindBankCard Kind = "bank_card"
)

// Payload это JSON объект, зашифрованный как одна версия секрета.
type Payload struct {
	Kind         Kind   `json:"kind"`                    // Kind идентифицирует, какие поля имеют смысл в Payload.
	Meta         string `json:"meta,omitempty"`          // Meta это метаданные секрета.
	Title        string `json:"title,omitempty"`         // Title это заголовок секрета.
	URL          string `json:"url,omitempty"`           // URL это URL секрета.
	Login        string `json:"login,omitempty"`         // Login это логин секрета.
	Password     string `json:"password,omitempty"`      // Password это пароль секрета.
	Text         string `json:"text,omitempty"`          // Text это текст секрета.
	BinaryBase64 string `json:"binary_base64,omitempty"` // BinaryBase64 это бинарные данные секрета.
	CardHolder   string `json:"card_holder,omitempty"`   // CardHolder это держатель карты секрета.
	CardNumber   string `json:"card_number,omitempty"`   // CardNumber это номер карты секрета.
	Expiry       string `json:"expiry,omitempty"`        // Expiry это срок секрета.
	CVC          string `json:"cvc,omitempty"`           // CVC это CVC секрета.
}

// ErrValidation означает отсутствие или несогласованность полей для Kind.
var ErrValidation = errors.New("clientdata: validation failed")

// MarshalJSON возвращает канонические байты JSON для шифрования.
func (p *Payload) MarshalJSONBytes() ([]byte, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return json.Marshal(p)
}

// UnmarshalJSONBytes парсит расшифрованный JSON.
func UnmarshalJSONBytes(b []byte) (*Payload, error) {
	var p Payload
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, err
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return &p, nil
}

// Validate проверяет необходимые поля для Kind.
func (p *Payload) Validate() error {
	switch p.Kind {
	case KindLoginPair:
		if p.Login == "" || p.Password == "" {
			return fmt.Errorf("%w: login_pair needs login and password", ErrValidation)
		}
	case KindText:
		if p.Text == "" {
			return fmt.Errorf("%w: text needs non-empty text", ErrValidation)
		}
	case KindBinary:
		if p.BinaryBase64 == "" {
			return fmt.Errorf("%w: binary needs binary_base64", ErrValidation)
		}
		if _, err := base64.StdEncoding.DecodeString(p.BinaryBase64); err != nil {
			return fmt.Errorf("%w: binary_base64 must be valid base64: %v", ErrValidation, err)
		}
	case KindBankCard:
		if p.CardNumber == "" || p.Expiry == "" {
			return fmt.Errorf("%w: bank_card needs card_number and expiry", ErrValidation)
		}
	default:
		return fmt.Errorf("%w: unknown kind %q", ErrValidation, p.Kind)
	}
	return nil
}

// BinaryBytes декодирует BinaryBase64 после проверки.
func (p *Payload) BinaryBytes() ([]byte, error) {
	if p.Kind != KindBinary {
		return nil, fmt.Errorf("not a binary payload")
	}
	return base64.StdEncoding.DecodeString(p.BinaryBase64)
}

// SetBinary хранит raw байты как base64.
func (p *Payload) SetBinary(b []byte) {
	p.Kind = KindBinary
	p.BinaryBase64 = base64.StdEncoding.EncodeToString(b)
}

// VersionListTitle — строка, которую клиент шифрует в title_encrypted версии (одинаковая логика для всех Kind).
// Всегда непустая (после trim), чтобы согласоваться с NOT NULL title_encrypted в БД и не шифровать пустой UTF-8.
func VersionListTitle(p *Payload) string {
	var s string

	// если payload пустой, возвращаем "секрет"
	if p == nil {
		return "секрет"
	}

	// получаем заголовок (поля Title и Meta есть у любого типа)
	if t := strings.TrimSpace(p.Title); t != "" {
		s = t
	} else if t := strings.TrimSpace(p.Meta); t != "" {
		s = t
	}

	// если заголовок пустой, возвращаем тип секрета в человеко-читаемом виде
	if s == "" {
		switch p.Kind {
		case KindLoginPair:
			s = "логин/пароль"
		case KindText:
			s = "текст"
		case KindBinary:
			s = "бинарные данные"
		case KindBankCard:
			s = "банковская карта"
		}
	}

	// проверяем, что строка не пустая
	if strings.TrimSpace(s) == "" {
		return "секрет"
	}

	return s
}
