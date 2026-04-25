// Package tui implements the interactive terminal UI for the GophKeeper client.
package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/coolycow/gophkeeper/internal/buildinfo"
	"github.com/coolycow/gophkeeper/internal/clientdata"
	"github.com/coolycow/gophkeeper/internal/clientgrpc"
	"github.com/coolycow/gophkeeper/internal/clientsession"
	"github.com/coolycow/gophkeeper/internal/config"
	"github.com/coolycow/gophkeeper/internal/proto/gophkeeperpb"
	"github.com/coolycow/gophkeeper/internal/secretcrypto"
)

// view — текущий экран приложения.
type view int

// доступные экраны приложения.
const (
	viewMenu     view = iota // меню
	viewLogin                // вход
	viewRegister             // регистрация
	viewList                 // список секретов
	viewDetail               // детали секрета
	viewCreate               // создание / правка секрета
)

// Model — состояние Bubble Tea.
type Model struct {
	cfg *config.ConfigClient // конфигурация клиента

	width  int  // ширина окна
	height int  // высота окна
	v      view // текущий экран

	api *clientgrpc.Client // клиент gRPC

	emailTI         textinput.Model   // поле ввода email
	passTI          textinput.Model   // поле ввода пароля
	passAgainTI     textinput.Model   // поле ввода пароля ещё раз
	createInputs    []textinput.Model // поля формы «новый / правка секрета»
	createFieldKeys []string          // ключи полей (порядок = порядок createInputs)
	createFormFocus int               // фокус в форме создания
	formFocus       int               // фокус на поле ввода (логин / регистрация)

	passwordSession string // пароль сессии

	secrets          []*gophkeeperpb.SecretSummary // список секретов
	secretListTitles []string                      // расшифрованные title (тот же порядок, что secrets)
	secretListVers   []int32                       // номера текущих версий (тот же порядок, что secrets)
	cursor           int                           // курсор на секрете

	detailID               string                        // ID секрета
	detailPayload          *clientdata.Payload           // payload секрета
	detailReveal           bool                          // показать пароль и др. скрытые поля в просмотре
	detailCurrentVersionID string                        // ID текущей версии секрета
	detailShownVersionID   string                        // версия, которую сейчас показываем в payload-блоке
	detailVersions         []*gophkeeperpb.SecretVersion // список версий секрета
	detailVersionTitles    []string                      // расшифрованные title версий
	detailVersionCursor    int                           // курсор на версии секрета

	createKind      clientdata.Kind   // тип секрета
	createDraft     map[string]string // draft секрета
	editingSecretID string            // ID секрета

	errLine string // ошибка
	info    string // информация

	sessionPath string // путь к сессии
}

// стили для текста
var titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")) // желтый цвет для заголовков (жирный, розовый)
var errStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))              // красный цвет для ошибок (красный)
var hintStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))             // серый цвет для подсказок (серый)

// стили таблицы списка секретов
var listHeaderStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))                              // заголовки столбцов
var listDateColStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))                                       // даты
var listTitleColStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))                                      // название
var listRowSelectedStyle = lipgloss.NewStyle().Background(lipgloss.Color("238")).Foreground(lipgloss.Color("255")) // выбранная строка
var listTableBorderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))                                   // линия под заголовком

// New создаёт начальную модель TUI.
func New(cfg *config.ConfigClient) tea.Model {
	em := textinput.New()
	em.Placeholder = "email"
	em.CharLimit = 128
	em.Width = 48

	pw := textinput.New()
	pw.Placeholder = "пароль"
	pw.EchoMode = textinput.EchoPassword
	pw.CharLimit = 200
	pw.Width = 48

	pw2 := textinput.New()
	pw2.Placeholder = "пароль ещё раз"
	pw2.EchoMode = textinput.EchoPassword
	pw2.CharLimit = 200
	pw2.Width = 48

	m := &Model{
		cfg:         cfg,
		v:           viewMenu,
		emailTI:     em,
		passTI:      pw,
		passAgainTI: pw2,
		createDraft: map[string]string{},
	}
	if sp, err := clientsession.DefaultPath(); err == nil {
		m.sessionPath = sp
	}
	return m
}

// Init реализует tea.Model.
func (m *Model) Init() tea.Cmd {
	return textinput.Blink
}

// Update реализует tea.Model.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			if m.v == viewMenu {
				return m, tea.Quit
			}
			if m.v == viewDetail {
				m.v = viewList
				m.detailReveal = false
				m.detailVersions = nil
				m.detailVersionTitles = nil
				m.detailVersionCursor = 0
				m.errLine = ""
				return m, nil
			}
			if m.v == viewCreate {
				m.v = viewList
				m.resetCreateWizard()
				m.errLine = ""
				return m, nil
			}
			m.backToMenu()
			return m, nil
		}
	}

	switch m.v {
	case viewMenu:
		return m.updateMenu(msg)
	case viewLogin:
		return m.updateLogin(msg)
	case viewRegister:
		return m.updateRegister(msg)
	case viewList:
		return m.updateList(msg)
	case viewDetail:
		return m.updateDetail(msg)
	case viewCreate:
		return m.updateCreate(msg)
	}
	return m, nil
}

// backToMenu возвращает на меню
func (m *Model) backToMenu() {
	m.closeConn()
	m.v = viewMenu
	m.passwordSession = ""
	m.secrets = nil
	m.cursor = 0
	m.formFocus = 0
	m.errLine = ""
	m.info = ""
}

// closeConn закрывает соединение с сервером
func (m *Model) closeConn() {
	if m.api != nil {
		_ = m.api.Close()
	}
	m.api = nil
}

// updateMenu обновляет состояние при открытии меню
func (m *Model) updateMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "l", "L":
			m.v = viewLogin
			m.formFocus = 0
			m.emailTI.Focus()
			m.passTI.Blur()
			m.passAgainTI.Blur()
			if m.cfg.Email != "" {
				m.emailTI.SetValue(m.cfg.Email)
			}
			if f, err := clientsession.Load(m.sessionPath); err == nil && f != nil &&
				strings.TrimSpace(f.ServerAddress) == strings.TrimSpace(m.cfg.ServerAddress) && f.Email != "" {
				m.emailTI.SetValue(f.Email)
			}
			return m, textinput.Blink
		case "r", "R":
			m.v = viewRegister
			m.formFocus = 0
			m.emailTI.Focus()
			m.passTI.Blur()
			m.passAgainTI.Blur()
			return m, textinput.Blink
		case "q", "Q":
			return m, tea.Quit
		}
	}
	return m, nil
}

// updateLogin обновляет состояние при входе
func (m *Model) updateLogin(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "tab":
			m.formFocus = (m.formFocus + 1) % 2
			m.switchLoginFocus()
			return m, textinput.Blink
		case "enter":
			return m.submitLogin()
		}
	}
	var cmd tea.Cmd
	switch m.formFocus {
	case 0:
		m.emailTI, cmd = m.emailTI.Update(msg)
	default:
		m.passTI, cmd = m.passTI.Update(msg)
	}
	return m, cmd
}

// switchLoginFocus переключает фокус на следующее поле ввода
func (m *Model) switchLoginFocus() {
	switch m.formFocus {
	case 0:
		m.emailTI.Focus()
		m.passTI.Blur()
	default:
		m.emailTI.Blur()
		m.passTI.Focus()
	}
}

// submitLogin отправляет запрос на вход
func (m *Model) submitLogin() (tea.Model, tea.Cmd) {
	// проверяем, что email и пароль введены
	email := strings.TrimSpace(m.emailTI.Value())
	pass := m.passTI.Value()
	if email == "" || pass == "" {
		m.errLine = "Введите email и пароль"
		return m, nil
	}

	// устанавливаем таймаут для запроса
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	// открываем соединение с сервером
	conn, err := clientgrpc.Dial(ctx, clientgrpc.DialOptions{
		Address:   m.cfg.ServerAddress,
		Insecure:  m.cfg.GRPCInsecure,
		TLSCAFile: m.cfg.TLSCAFile,
	})

	if err != nil {
		m.errLine = fmt.Sprintf("Подключение: %v", err)
		return m, nil
	}

	// создаем клиент gRPC
	api := clientgrpc.NewClient(conn)

	// выполняем запрос на вход
	if err := api.Login(ctx, email, pass); err != nil {
		_ = api.Close()
		m.errLine = fmt.Sprintf("Вход: %v", err)
		return m, nil
	}

	// сохраняем сессию
	if err := clientsession.Save(m.sessionPath, &clientsession.File{
		ServerAddress: m.cfg.ServerAddress,
		Email:         email,
		RefreshToken:  api.RefreshToken(),
	}); err != nil {
		m.info = fmt.Sprintf("Сессия не записана: %v", err)
	}

	m.api = api
	m.passwordSession = pass
	m.errLine = ""
	m.v = viewList

	// обновляем список секретов
	if err := m.reloadSecrets(ctx); err != nil {
		m.errLine = fmt.Sprintf("Список: %v", err)
	}

	return m, nil
}

// reloadSecrets обновляет список секретов
func (m *Model) reloadSecrets(ctx context.Context) error {
	// проверяем, что соединение с сервером установлено
	if m.api == nil {
		return fmt.Errorf("нет соединения")
	}

	// проверяем, что токены не истекли
	if clientgrpc.NeedsRefreshSoon(m.api.AccessExpiresAt(), 30*time.Second, time.Now()) {
		_ = m.api.Refresh(ctx)
	}

	// получаем список секретов
	resp, err := m.api.ListSecrets(ctx, gophkeeperpb.SecretListScope_SECRET_LIST_SCOPE_ACTIVE_ONLY)
	if err != nil {
		return err
	}

	// сохраняем список секретов
	m.secrets = resp.GetSecrets()
	if m.cursor >= len(m.secrets) {
		m.cursor = max(0, len(m.secrets)-1)
	}
	m.fillSecretListTitles()
	m.fillSecretListVersions(ctx)

	return nil
}

// fillSecretListTitles расшифровывает title_encrypted по списку (один Argon2 на весь пакет).
func (m *Model) fillSecretListTitles() {
	//``
	m.secretListTitles = nil
	if m.api == nil || len(m.secrets) == 0 || m.passwordSession == "" {
		return
	}

	// получаем ключ для расшифровки данных
	key, err := secretcrypto.DeriveKeyFromPassword(m.passwordSession, m.api.SaltHex())
	if err != nil {
		m.secretListTitles = make([]string, len(m.secrets))
		for i, s := range m.secrets {
			m.secretListTitles[i] = shortID(s.GetId())
		}
		return
	}

	// получаем ключ для расшифровки данных
	m.secretListTitles = make([]string, len(m.secrets))
	for i, s := range m.secrets {
		// получаем зашифрованное название секрета
		te := s.GetTitleEncrypted()

		// проверяем, что зашифрованное название не пустое
		if len(te) == 0 {
			m.secretListTitles[i] = shortID(s.GetId())
			continue
		}

		// расшифровываем название секрета
		plain, err := secretcrypto.DecryptWithKey(te, key)

		// проверяем, что расшифровка прошла успешно
		if err != nil {
			m.secretListTitles[i] = shortID(s.GetId())
			continue
		}

		m.secretListTitles[i] = string(plain)
	}
}

// fillSecretListVersions заполняет номера текущих версий по списку секретов.
// Best-effort: при ошибке оставляет 0 для конкретной строки, не прерывая общий рендер списка.
func (m *Model) fillSecretListVersions(ctx context.Context) {
	m.secretListVers = nil
	if m.api == nil || len(m.secrets) == 0 {
		return
	}
	m.secretListVers = make([]int32, len(m.secrets))
	for i, s := range m.secrets {
		sec, err := m.api.GetSecret(ctx, s.GetId(), false)
		if err != nil || sec == nil || sec.GetCurrentVersion() == nil {
			m.secretListVers[i] = 0
			continue
		}
		m.secretListVers[i] = sec.GetCurrentVersion().GetVersion()
	}
}

// updateRegister обновляет состояние при регистрации
func (m *Model) updateRegister(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "tab":
			m.formFocus = (m.formFocus + 1) % 3
			m.switchRegisterFocus()
			return m, textinput.Blink
		case "enter":
			return m.submitRegister()
		}
	}

	// обновляем состояние полей ввода
	var cmd tea.Cmd
	switch m.formFocus {
	case 0:
		m.emailTI, cmd = m.emailTI.Update(msg)
	case 1:
		m.passTI, cmd = m.passTI.Update(msg)
	default:
		m.passAgainTI, cmd = m.passAgainTI.Update(msg)
	}

	return m, cmd
}

// switchRegisterFocus переключает фокус на следующее поле ввода
func (m *Model) switchRegisterFocus() {
	// сбрасываем фокус на все поля ввода
	m.emailTI.Blur()
	m.passTI.Blur()
	m.passAgainTI.Blur()

	// переключаем фокус на следующее поле ввода
	switch m.formFocus {
	case 0:
		m.emailTI.Focus()
	case 1:
		m.passTI.Focus()
	default:
		m.passAgainTI.Focus()
	}
}

// submitRegister отправляет запрос на регистрацию
func (m *Model) submitRegister() (tea.Model, tea.Cmd) {
	// проверяем, что пароли совпадают
	email := strings.TrimSpace(m.emailTI.Value())
	p1 := m.passTI.Value()
	p2 := m.passAgainTI.Value()
	if p1 != p2 {
		m.errLine = "Пароли не совпадают"
		return m, nil
	}

	// устанавливаем таймаут для запроса
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	// открываем соединение с сервером
	conn, err := clientgrpc.Dial(ctx, clientgrpc.DialOptions{
		Address:   m.cfg.ServerAddress,
		Insecure:  m.cfg.GRPCInsecure,
		TLSCAFile: m.cfg.TLSCAFile,
	})
	if err != nil {
		m.errLine = fmt.Sprintf("Подключение: %v", err)
		return m, nil
	}

	// создаем клиент gRPC
	api := clientgrpc.NewClient(conn)
	if err := api.Register(ctx, email, p1); err != nil {
		_ = api.Close()
		m.errLine = fmt.Sprintf("Регистрация: %v", err)
		return m, nil
	}

	// сохраняем сессию
	_ = clientsession.Save(m.sessionPath, &clientsession.File{
		ServerAddress: m.cfg.ServerAddress,
		Email:         email,
		RefreshToken:  api.RefreshToken(),
	})

	m.api = api
	m.passwordSession = p1
	m.errLine = ""
	m.v = viewList

	// обновляем список секретов
	if err := m.reloadSecrets(ctx); err != nil {
		m.errLine = fmt.Sprintf("Список: %v", err)
	}

	return m, nil
}

// updateList обновляет состояние при открытии списка секретов
func (m *Model) updateList(msg tea.Msg) (tea.Model, tea.Cmd) {
	// проверяем, что соединение с сервером установлено
	if m.api == nil {
		m.errLine = "Сессия сброшена"
		m.backToMenu()
		return m, nil
	}

	// обрабатываем нажатие клавиш
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "r", "R":
			ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
			defer cancel()
			if err := m.reloadSecrets(ctx); err != nil {
				m.errLine = err.Error()
			} else {
				m.info = "Список обновлён"
			}
			return m, nil
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil
		case "down", "j":
			if m.cursor < len(m.secrets)-1 {
				m.cursor++
			}
			return m, nil
		case "enter":
			if len(m.secrets) == 0 {
				return m, nil
			}
			return m.openDetail(m.secrets[m.cursor].GetId())
		case "n", "N":
			m.startCreateWizard("")
			return m, textinput.Blink
		case "d", "D":
			if len(m.secrets) == 0 {
				return m, nil
			}
			return m.deleteCurrent()
		case "q", "Q":
			m.backToMenu()
			return m, nil
		}
	}
	return m, nil
}

// deleteCurrent удаляет текущий секрет
func (m *Model) deleteCurrent() (tea.Model, tea.Cmd) {
	// получаем ID текущего секрета
	id := m.secrets[m.cursor].GetId()

	// устанавливаем таймаут для запроса
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	// выполняем запрос на удаление секрета
	if err := m.api.DeleteSecret(ctx, id); err != nil {
		m.errLine = fmt.Sprintf("Удаление: %v", err)
		return m, nil
	}

	// обновляем список секретов
	_ = m.reloadSecrets(ctx)
	m.info = "Секрет удалён (на сервере — мягкое удаление)"

	return m, nil
}

// openDetail открывает детали секрета
func (m *Model) openDetail(secretID string) (tea.Model, tea.Cmd) {
	if err := m.reloadDetail(secretID, ""); err != nil {
		m.errLine = fmt.Sprintf("Загрузка: %v", err)
		return m, nil
	}
	m.detailReveal = false
	m.v = viewDetail
	m.errLine = ""
	return m, nil
}

// reloadDetail перечитывает текущую версию секрета и список всех версий.
func (m *Model) reloadDetail(secretID, preferredVersionID string) error {
	// устанавливаем таймаут для запроса
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// выполняем запрос на получение секрета
	sec, err := m.api.GetSecret(ctx, secretID, false)
	if err != nil {
		return err
	}

	// получаем все версии секрета
	versions, err := m.api.ListSecretVersions(ctx, secretID)
	if err != nil {
		return err
	}

	// получаем текущую версию секрета
	cv := sec.GetCurrentVersion()
	if cv == nil {
		return fmt.Errorf("нет текущей версии")
	}

	// получаем ключ для расшифровки данных
	key, err := secretcrypto.DeriveKeyFromPassword(m.passwordSession, m.api.SaltHex())
	if err != nil {
		return fmt.Errorf("ключ: %w", err)
	}

	// расшифровываем данные секрета
	plain, err := secretcrypto.DecryptWithKey(cv.GetDataEncrypted(), key)
	if err != nil {
		return fmt.Errorf("расшифровка: %w", err)
	}

	// десериализуем данные секрета
	p, err := clientdata.UnmarshalJSONBytes(plain)
	if err != nil {
		return fmt.Errorf("формат данных: %w", err)
	}

	// сохраняем данные секрета
	m.detailID = secretID
	m.detailPayload = p
	m.detailCurrentVersionID = sec.GetCurrentSecretVersionId()

	// если текущая версия не установлена, устанавливаем её на основе ID текущей версии
	if m.detailCurrentVersionID == "" {
		m.detailCurrentVersionID = cv.GetId()
	}

	// устанавливаем ID версии, которую сейчас показываем в payload-блоке
	m.detailShownVersionID = m.detailCurrentVersionID

	// устанавливаем список версий секрета
	m.detailVersions = versions

	// заполняем заголовки версий секрета
	m.fillDetailVersionTitles(key)
	m.pickDetailVersionCursor(preferredVersionID)

	return nil
}

// fillDetailVersionTitles заполняет заголовки версий секрета
func (m *Model) fillDetailVersionTitles(key []byte) {
	m.detailVersionTitles = make([]string, len(m.detailVersions))
	for i, v := range m.detailVersions {
		fallback := fmt.Sprintf("Версия %d", v.GetVersion())
		if v.GetVersion() == 0 {
			fallback = shortID(v.GetId())
		}
		te := v.GetTitleEncrypted()
		if len(te) == 0 {
			m.detailVersionTitles[i] = fallback
			continue
		}
		plain, err := secretcrypto.DecryptWithKey(te, key)
		if err != nil || len(plain) == 0 {
			m.detailVersionTitles[i] = fallback
			continue
		}
		m.detailVersionTitles[i] = string(plain)
	}
}

// pickDetailVersionCursor выбирает курсор на версии секрета
func (m *Model) pickDetailVersionCursor(preferredVersionID string) {
	if len(m.detailVersions) == 0 {
		m.detailVersionCursor = 0
		return
	}
	targetID := preferredVersionID
	if targetID == "" {
		targetID = m.detailCurrentVersionID
	}
	for i, v := range m.detailVersions {
		if v.GetId() == targetID {
			m.detailVersionCursor = i
			return
		}
	}
	if m.detailVersionCursor >= len(m.detailVersions) {
		m.detailVersionCursor = len(m.detailVersions) - 1
	}
	if m.detailVersionCursor < 0 {
		m.detailVersionCursor = 0
	}
}

func (m *Model) selectedDetailVersion() *gophkeeperpb.SecretVersion {
	if m.detailVersionCursor < 0 || m.detailVersionCursor >= len(m.detailVersions) {
		return nil
	}
	return m.detailVersions[m.detailVersionCursor]
}

// loadDetailPayloadFromVersion загружает payload из выбранной версии секрета
func (m *Model) loadDetailPayloadFromVersion(v *gophkeeperpb.SecretVersion) error {
	// проверяем, что версия не выбрана
	if v == nil {
		return fmt.Errorf("версия не выбрана")
	}

	// получаем ключ для расшифровки данных
	key, err := secretcrypto.DeriveKeyFromPassword(m.passwordSession, m.api.SaltHex())
	if err != nil {
		return fmt.Errorf("ключ: %w", err)
	}

	// расшифровываем данные секрета
	plain, err := secretcrypto.DecryptWithKey(v.GetDataEncrypted(), key)
	if err != nil {
		return fmt.Errorf("расшифровка: %w", err)
	}

	// десериализуем данные секрета
	p, err := clientdata.UnmarshalJSONBytes(plain)
	if err != nil {
		return fmt.Errorf("формат данных: %w", err)
	}

	// сохраняем payload из выбранной версии секрета
	m.detailPayload = p

	// устанавливаем ID версии, которую сейчас показываем в payload-блоке
	m.detailShownVersionID = v.GetId()

	// возвращаем nil, если ошибки нет
	return nil
}

// updateDetail обновляет состояние при открытии деталей секрета
func (m *Model) updateDetail(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case "up", "k":
		if m.detailVersionCursor > 0 {
			m.detailVersionCursor--
		}
		return m, nil
	case "down", "j":
		if m.detailVersionCursor < len(m.detailVersions)-1 {
			m.detailVersionCursor++
		}
		return m, nil
	case "enter":
		return m.restoreSelectedDetailVersion()
	case " ", "space":
		v := m.selectedDetailVersion()

		// загружаем payload из выбранной версии секрета
		if err := m.loadDetailPayloadFromVersion(v); err != nil {
			m.errLine = fmt.Sprintf("Просмотр версии: %v", err)
			return m, nil
		}

		// устанавливаем сообщение об ошибке
		m.errLine = ""

		// устанавливаем сообщение об информации
		if v.GetId() == m.detailCurrentVersionID {
			m.info = "Просмотр текущей версии"
		} else {
			m.info = fmt.Sprintf("Просмотр версии %d (без назначения текущей)", v.GetVersion())
		}
		return m, nil
	case "x", "X":
		return m.deleteSelectedDetailVersion()
	case "z", "Z":
		return m.compressDetailSecret()
	case "e", "E":
		m.startCreateWizard(m.detailID)
		return m, textinput.Blink
	case "h", "H":
		m.detailReveal = !m.detailReveal
		m.info = ""
		m.errLine = ""
		return m, nil
	case "c", "C":
		if m.detailPayload == nil || m.detailPayload.Kind != clientdata.KindLoginPair {
			return m, nil
		}
		if err := clipboard.WriteAll(m.detailPayload.Password); err != nil {
			m.errLine = fmt.Sprintf("Буфер: %v", err)
		} else {
			m.errLine = ""
			m.info = "Пароль скопирован"
		}
		return m, nil
	default:
		return m, nil
	}
}

func (m *Model) restoreSelectedDetailVersion() (tea.Model, tea.Cmd) {
	sv := m.selectedDetailVersion()
	if sv == nil {
		return m, nil
	}
	if sv.GetId() == m.detailCurrentVersionID {
		m.info = "Эта версия уже текущая"
		m.errLine = ""
		return m, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := m.api.RestoreSecretVersion(ctx, m.detailID, sv.GetId()); err != nil {
		m.errLine = fmt.Sprintf("Восстановление версии: %v", err)
		return m, nil
	}
	if err := m.reloadDetail(m.detailID, sv.GetId()); err != nil {
		m.errLine = fmt.Sprintf("Обновление деталей: %v", err)
		return m, nil
	}
	if err := m.reloadSecrets(ctx); err != nil {
		m.errLine = fmt.Sprintf("Обновление списка: %v", err)
		return m, nil
	}
	m.errLine = ""
	m.info = "Выбранная версия сделана текущей"
	return m, nil
}

func (m *Model) deleteSelectedDetailVersion() (tea.Model, tea.Cmd) {
	sv := m.selectedDetailVersion()
	if sv == nil {
		return m, nil
	}
	if len(m.detailVersions) <= 1 {
		m.errLine = "Нельзя удалить единственную версию"
		return m, nil
	}
	if sv.GetId() == m.detailCurrentVersionID {
		m.errLine = "Нельзя удалить текущую версию. Сначала выберите другую и нажмите Enter"
		return m, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := m.api.DeleteSecretVersion(ctx, m.detailID, sv.GetId()); err != nil {
		m.errLine = fmt.Sprintf("Удаление версии: %v", err)
		return m, nil
	}
	if err := m.reloadDetail(m.detailID, ""); err != nil {
		m.errLine = fmt.Sprintf("Обновление деталей: %v", err)
		return m, nil
	}
	m.errLine = ""
	m.info = "Версия удалена"
	return m, nil
}

func (m *Model) compressDetailSecret() (tea.Model, tea.Cmd) {
	if len(m.detailVersions) <= 1 {
		m.errLine = ""
		m.info = "История уже сжата: осталась только текущая версия"
		return m, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	deleted := 0
	for _, v := range m.detailVersions {
		if v.GetId() == m.detailCurrentVersionID {
			continue
		}
		if err := m.api.DeleteSecretVersion(ctx, m.detailID, v.GetId()); err != nil {
			m.errLine = fmt.Sprintf("Compress прерван после %d удалений: %v", deleted, err)
			return m, nil
		}
		deleted++
	}

	if err := m.reloadDetail(m.detailID, m.detailCurrentVersionID); err != nil {
		m.errLine = fmt.Sprintf("Обновление деталей: %v", err)
		return m, nil
	}
	m.errLine = ""
	m.info = fmt.Sprintf("История сжата, удалено версий: %d", deleted)
	return m, nil
}

// resetCreateWizard сбрасывает состояние при создании секрета
func (m *Model) resetCreateWizard() {
	m.createKind = ""
	m.createDraft = map[string]string{}
	m.editingSecretID = ""
	m.createInputs = nil
	m.createFieldKeys = nil
	m.createFormFocus = 0
}

// fieldKeysForKind возвращает порядок полей формы для типа секрета.
func fieldKeysForKind(k clientdata.Kind) []string {
	switch k {
	case clientdata.KindLoginPair:
		return []string{"title", "login", "password", "url", "meta"}
	case clientdata.KindText:
		return []string{"title", "text", "meta"}
	case clientdata.KindBinary:
		return []string{"title", "binary_base64", "meta"}
	case clientdata.KindBankCard:
		return []string{"title", "card_holder", "card_number", "expiry", "cvc", "meta"}
	default:
		return nil
	}
}

func fieldPlaceholder(key string) string {
	return map[string]string{
		"meta":          "Метаданные (можно пусто)",
		"title":         "Заголовок",
		"login":         "Логин",
		"password":      "Пароль",
		"url":           "URL (можно пусто)",
		"text":          "Текст",
		"binary_base64": "Данные в Base64",
		"card_holder":   "Имя на карте",
		"card_number":   "Номер карты",
		"expiry":        "Срок (MM/YY)",
		"cvc":           "CVC",
	}[key]
}

func kindTitleRU(k clientdata.Kind) string {
	switch k {
	case clientdata.KindLoginPair:
		return "Логин / пароль"
	case clientdata.KindText:
		return "Текст"
	case clientdata.KindBinary:
		return "Бинарные данные (Base64)"
	case clientdata.KindBankCard:
		return "Банковская карта"
	default:
		return "Секрет"
	}
}

// buildCreateFormInputs собирает поля ввода для выбранного типа (значения из createDraft).
func (m *Model) buildCreateFormInputs() {
	// получаем поля для выбранного типа
	keys := fieldKeysForKind(m.createKind)
	m.createFieldKeys = keys
	m.createInputs = make([]textinput.Model, len(keys))

	// создаем поля ввода для каждого поля
	for i, key := range keys {
		ti := textinput.New()
		ti.Placeholder = fieldPlaceholder(key)
		ti.Width = 56
		ti.CharLimit = 0

		// устанавливаем режим отображения пароля для пароля и CVC
		if key == "password" || key == "cvc" {
			ti.EchoMode = textinput.EchoPassword
		} else {
			ti.EchoMode = textinput.EchoNormal
		}

		// устанавливаем значение для поля
		if v, ok := m.createDraft[key]; ok {
			ti.SetValue(v)
		}

		m.createInputs[i] = ti
	}

	m.createFormFocus = 0
	m.focusCreateForm()
}

func (m *Model) focusCreateForm() {
	for i := range m.createInputs {
		if i == m.createFormFocus {
			m.createInputs[i].Focus()
		} else {
			m.createInputs[i].Blur()
		}
	}
}

func (m *Model) syncCreateDraftFromInputs() {
	for i, key := range m.createFieldKeys {
		if i < len(m.createInputs) {
			m.createDraft[key] = m.createInputs[i].Value()
		}
	}
}

// startCreateWizard начинает новый секрет или правку (editSecretID непустой).
func (m *Model) startCreateWizard(editSecretID string) {
	// сбрасываем состояние при создании секрета
	m.resetCreateWizard()
	m.editingSecretID = editSecretID
	m.v = viewCreate
	m.errLine = ""
	m.info = ""

	// если ID секрета не пустой и payload не nil, то заполняем данные из payload
	if editSecretID != "" && m.detailPayload != nil {
		m.createKind = m.detailPayload.Kind
		m.prefillDraftFromPayload()
		m.buildCreateFormInputs()
		return
	}

	m.info = "Выберите тип: 1 — логин/пароль  2 — текст  3 — base64  4 — банковская карта"
}

// updateCreate обновляет состояние при создании секрета
func (m *Model) updateCreate(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.createKind == "" {
		if key, ok := msg.(tea.KeyMsg); ok {
			switch key.String() {
			case "1":
				m.pickKind(clientdata.KindLoginPair)
			case "2":
				m.pickKind(clientdata.KindText)
			case "3":
				m.pickKind(clientdata.KindBinary)
			case "4":
				m.pickKind(clientdata.KindBankCard)
			}
		}
		return m, textinput.Blink
	}

	if len(m.createInputs) == 0 {
		m.buildCreateFormInputs()
	}

	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "tab":
			if len(m.createInputs) == 0 {
				return m, nil
			}
			m.createFormFocus = (m.createFormFocus + 1) % len(m.createInputs)
			m.focusCreateForm()
			return m, textinput.Blink
		case "enter":
			return m.finishCreate()
		}
	}

	if len(m.createInputs) == 0 {
		return m, nil
	}
	var cmd tea.Cmd
	f := m.createFormFocus
	if f < 0 || f >= len(m.createInputs) {
		f = 0
	}
	m.createInputs[f], cmd = m.createInputs[f].Update(msg)
	return m, cmd
}

// pickKind выбирает тип секрета и строит одну форму на экран.
func (m *Model) pickKind(k clientdata.Kind) {
	m.createKind = k
	if m.editingSecretID != "" && m.detailPayload != nil {
		m.prefillDraftFromPayload()
	}
	m.buildCreateFormInputs()
	m.info = "" // иначе внизу останется строка «Выберите тип…» из глобального вывода info
}

// prefillDraftFromPayload заполняет draft из payload
func (m *Model) prefillDraftFromPayload() {
	p := m.detailPayload
	if p == nil {
		return
	}
	m.createDraft["title"] = p.Title
	switch m.createKind {
	case clientdata.KindLoginPair:
		m.createDraft["login"] = p.Login
		m.createDraft["password"] = p.Password
		m.createDraft["url"] = p.URL
		m.createDraft["meta"] = p.Meta
	case clientdata.KindText:
		m.createDraft["text"] = p.Text
		m.createDraft["meta"] = p.Meta
	case clientdata.KindBinary:
		m.createDraft["binary_base64"] = p.BinaryBase64
		m.createDraft["meta"] = p.Meta
	case clientdata.KindBankCard:
		m.createDraft["card_holder"] = p.CardHolder
		m.createDraft["card_number"] = p.CardNumber
		m.createDraft["expiry"] = p.Expiry
		m.createDraft["cvc"] = p.CVC
		m.createDraft["meta"] = p.Meta
	}
}

// finishCreate завершает создание секрета
func (m *Model) finishCreate() (tea.Model, tea.Cmd) {
	m.syncCreateDraftFromInputs()

	// создаем payload для данных секрета (заголовок, метаданные, заголовок)
	p := &clientdata.Payload{Kind: m.createKind, Meta: m.createDraft["meta"], Title: m.createDraft["title"]}

	// устанавливаем значения для полей
	switch m.createKind {
	case clientdata.KindLoginPair:
		p.Login = m.createDraft["login"]
		p.Password = m.createDraft["password"]
		p.URL = m.createDraft["url"]
	case clientdata.KindText:
		p.Text = m.createDraft["text"]
	case clientdata.KindBinary:
		p.BinaryBase64 = m.createDraft["binary_base64"]
	case clientdata.KindBankCard:
		p.CardHolder = m.createDraft["card_holder"]
		p.CardNumber = m.createDraft["card_number"]
		p.Expiry = m.createDraft["expiry"]
		p.CVC = m.createDraft["cvc"]
	}

	// сериализуем данные секрета
	raw, err := p.MarshalJSONBytes()
	if err != nil {
		m.errLine = err.Error()
		return m, nil
	}

	// получаем ключ для расшифровки данных
	listTitle := clientdata.VersionListTitle(p)
	key, err := secretcrypto.DeriveKeyFromPassword(m.passwordSession, m.api.SaltHex())
	if err != nil {
		m.errLine = fmt.Sprintf("Ключ: %v", err)
		return m, nil
	}

	// шифруем данные секрета
	ct, err := secretcrypto.EncryptWithKey(raw, key)
	if err != nil {
		m.errLine = fmt.Sprintf("Шифрование: %v", err)
		return m, nil
	}

	// шифруем название секрета
	titleCT, err := secretcrypto.EncryptWithKey([]byte(listTitle), key)
	if err != nil {
		m.errLine = fmt.Sprintf("Шифрование названия: %v", err)
		return m, nil
	}

	// проверяем, что title_encrypted не пустой
	if len(titleCT) == 0 {
		m.errLine = "внутренняя ошибка: пустой title_encrypted"
		return m, nil
	}

	// устанавливаем таймаут для запроса
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// выполняем запрос на создание или обновление секрета
	if m.editingSecretID != "" {
		_, err = m.api.UpdateSecret(ctx, m.editingSecretID, ct, titleCT, int32(secretcrypto.DataFormatVersion))
	} else {
		_, err = m.api.CreateSecret(ctx, ct, titleCT, int32(secretcrypto.DataFormatVersion))
	}

	// проверяем, что запрос выполнен успешно
	if err != nil {
		m.errLine = fmt.Sprintf("Отправка: %v", err)
		return m, nil
	}

	// обновляем список секретов
	_ = m.reloadSecrets(ctx)
	m.resetCreateWizard()
	m.v = viewList
	m.info = "Сохранено"
	return m, nil
}

// View реализует tea.Model.
func (m *Model) View() string {
	// создаем билдер для сборки строки
	var b strings.Builder

	// добавляем заголовок и пустую строку
	b.WriteString(titleStyle.Render(fmt.Sprintf("GophKeeper ver. %s by %s commit %s",
		buildinfo.OrNA(buildinfo.Version), buildinfo.OrNA(buildinfo.BuildDate), buildinfo.OrNA(buildinfo.BuildCommit))))
	b.WriteString("\n\n")

	// добавляем содержимое текущего экрана
	switch m.v {
	// экран меню
	case viewMenu:
		b.WriteString("l — вход по email и паролю\n")
		b.WriteString("r — регистрация нового пользователя\n")
		b.WriteString("q — выход\n")
		b.WriteString(hintStyle.Render(fmt.Sprintf("\nСервер (gRPC): %s; Insecure: %t", m.cfg.ServerAddress, m.cfg.GRPCInsecure)))
	// экран входа
	case viewLogin:
		b.WriteString("Вход\n\n")
		b.WriteString(m.emailTI.View() + "\n")
		b.WriteString(m.passTI.View() + "\n")
		b.WriteString(hintStyle.Render("\nTab — поле, Enter — войти, Esc — назад\n"))
	// экран регистрации
	case viewRegister:
		b.WriteString("Регистрация\n\n")
		b.WriteString(m.emailTI.View() + "\n")
		b.WriteString(m.passTI.View() + "\n")
		b.WriteString(m.passAgainTI.View() + "\n")
		b.WriteString(hintStyle.Render("\nTab — поле, Enter — создать аккаунт, Esc — назад\n"))
	// экран списка секретов
	case viewList:
		// Выводим Email пользователя
		b.WriteString(hintStyle.Render(fmt.Sprintf("Email: %s\n", m.emailTI.Value())))
		b.WriteString("\n")

		// Выводим заголовок списка секретов
		b.WriteString("Секреты (активные)\n\n")

		// Выводим список секретов
		if len(m.secrets) == 0 {
			b.WriteString("(пусто — нажмите n чтобы добавить)\n")
		} else {
			b.WriteString(renderSecretListTable(m))
		}

		// Выводим подсказки для экрана списка секретов
		b.WriteString(hintStyle.Render("\nj/k — курсор, Enter — открыть, n — новый, d — удалить, r — обновить, q — в меню\n"))
	// экран деталей секрета
	case viewDetail:
		// Выводим заголовок секрета
		b.WriteString(fmt.Sprintf("Секрет %s\n\n", shortID(m.detailID)))

		// Выводим payload секрета
		if m.detailPayload != nil {
			// Выводим сообщение о выбранной версии секрета
			if m.detailShownVersionID != "" && m.detailShownVersionID != m.detailCurrentVersionID {
				b.WriteString(hintStyle.Render("Просмотр: неактуальная версия\n"))
			} else {
				b.WriteString(hintStyle.Render("Просмотр: текущая версия\n"))
			}

			// Выводим payload секрета
			b.WriteString("\n")
			b.WriteString(renderPayload(m.detailPayload, m.detailReveal))
		}

		// Вывродим список версий секрета
		b.WriteString(hintStyle.Render("\nИСТОРИЯ ВЕРСИЙ СЕКРЕТА:\n"))
		b.WriteString("\n")
		if len(m.detailVersions) == 0 {
			b.WriteString("(версии не найдены)\n")
		} else {
			b.WriteString(renderSecretVersionsTable(m))
		}

		// добавляем подсказки для экрана деталей секрета
		hints := "j/k — выбор версии, Space — посмотреть, Enter — сделать текущей, x — удалить версию, z — compress, h — показать/скрыть, e — редактировать, Esc — к списку"

		// добавляем подсказки для экрана деталей секрета, если payload является логином/паролем
		if m.detailPayload != nil && m.detailPayload.Kind == clientdata.KindLoginPair {
			hints = "j/k — выбор версии, Space — посмотреть, Enter — сделать текущей, x — удалить версию, z — compress, h — показать/скрыть, c — скопировать пароль, e — редактировать, Esc — к списку"
		}

		b.WriteString(hintStyle.Render("\n" + hints + "\n"))
	// экран создания секрета
	case viewCreate:
		if m.createKind == "" {
			b.WriteString(m.info + "\n")
		} else {
			if m.editingSecretID != "" {
				b.WriteString("Правка секрета: ")
			} else {
				b.WriteString("Новый секрет: ")
			}
			b.WriteString(kindTitleRU(m.createKind) + "\n\n")
			for i := range m.createInputs {
				b.WriteString(m.createInputs[i].View() + "\n")
			}
			b.WriteString(hintStyle.Render("\nTab — поле, Enter — сохранить, Esc — отмена\n"))
		}
	}

	// добавляем ошибку, если она есть
	if m.errLine != "" {
		b.WriteString("\n" + errStyle.Render(m.errLine))
	}

	// добавляем информацию, если она есть и ошибки нет
	if m.info != "" && m.errLine == "" {
		b.WriteString("\n" + hintStyle.Render(m.info))
	}

	return b.String()
}

const (
	listDateColW    = 16   // "2006-01-02 15:04"
	listVersionColW = 7    // "Версия"/номер
	listColGapW     = 2    // зазор между столбцами
	currMark        = "● " // маркер текущей версии
)

// listTableTitleWidth — ширина колонки «Название» в ячейках дисплея.
func listTableTitleWidth(termW int, fixedTail int) int {
	if termW < 1 {
		termW = 100
	}
	used := 2 + fixedTail
	tw := termW - used
	if tw < 14 {
		tw = 14
	}
	return tw
}

// padListCell обрезает или дополняет строку до фиксированной ширины в терминале (корректно для wide runes).
func padListCell(s string, targetWidth int) string {
	if targetWidth < 1 {
		return ""
	}
	w := runewidth.StringWidth(s)
	if w > targetWidth {
		return runewidth.Truncate(s, targetWidth, "…")
	}
	return s + strings.Repeat(" ", targetWidth-w)
}

// renderSecretListTable рисует выровненные колонки с цветом и выделением курсора.
func renderSecretListTable(m *Model) string {
	termW := m.width
	if termW < 1 {
		termW = 100
	}
	tw := listTableTitleWidth(termW, listVersionColW+listColGapW+listDateColW+listColGapW+listDateColW+listColGapW)
	gap := strings.Repeat(" ", listColGapW)
	totalLineW := 2 + tw + listColGapW + listVersionColW + listColGapW + listDateColW + listColGapW + listDateColW
	sepW := totalLineW
	if sepW > termW-1 {
		sepW = termW - 1
	}
	if sepW < 8 {
		sepW = 8
	}

	var b strings.Builder
	h1 := listHeaderStyle.Render(padListCell("Название", tw))
	h2 := listHeaderStyle.Render(padListCell("Версия", listVersionColW))
	h3 := listHeaderStyle.Render(padListCell("Создан", listDateColW))
	h4 := listHeaderStyle.Render(padListCell("Изменён", listDateColW))
	b.WriteString("  ")
	b.WriteString(h1)
	b.WriteString(gap)
	b.WriteString(h2)
	b.WriteString(gap)
	b.WriteString(h3)
	b.WriteString(gap)
	b.WriteString(h4)
	b.WriteString("\n")
	b.WriteString(listTableBorderStyle.Render(strings.Repeat("─", sepW)))
	b.WriteString("\n")

	for i, s := range m.secrets {
		title := shortID(s.GetId())
		if i < len(m.secretListTitles) && m.secretListTitles[i] != "" {
			title = m.secretListTitles[i]
		}
		created := formatTs(s.GetCreatedAt())
		updated := formatTs(s.GetUpdatedAt())
		ver := "?"
		if i < len(m.secretListVers) && m.secretListVers[i] > 0 {
			ver = fmt.Sprintf("%d", m.secretListVers[i])
		}
		pref := "  "
		if i == m.cursor {
			pref = "> "
		}
		tCell := padListCell(title, tw)
		vCell := padListCell(ver, listVersionColW)
		cCell := padListCell(created, listDateColW)
		uCell := padListCell(updated, listDateColW)
		line := tCell + gap + vCell + gap + cCell + gap + uCell
		if i == m.cursor {
			b.WriteString(listRowSelectedStyle.Render(pref+line) + "\n")
		} else {
			b.WriteString(pref)
			b.WriteString(listTitleColStyle.Render(tCell))
			b.WriteString(gap)
			b.WriteString(listDateColStyle.Render(vCell))
			b.WriteString(gap)
			b.WriteString(listDateColStyle.Render(cCell))
			b.WriteString(gap)
			b.WriteString(listDateColStyle.Render(uCell))
			b.WriteString("\n")
		}
	}
	return b.String()
}

// renderSecretVersionsTable рисует таблицу версий секрета.
func renderSecretVersionsTable(m *Model) string {
	termW := m.width
	if termW < 1 {
		termW = 100
	}
	tw := listTableTitleWidth(termW, listVersionColW+listColGapW+listDateColW+listColGapW)
	gap := strings.Repeat(" ", listColGapW)
	totalLineW := 2 + listVersionColW + listColGapW + tw + listColGapW + listDateColW
	sepW := totalLineW
	if sepW > termW-1 {
		sepW = termW - 1
	}
	if sepW < 8 {
		sepW = 8
	}

	var b strings.Builder
	h1 := listHeaderStyle.Render(padListCell("Версия", listVersionColW))
	h2 := listHeaderStyle.Render(padListCell("Заголовок", tw))
	h3 := listHeaderStyle.Render(padListCell("Создана", listDateColW))
	b.WriteString("  ")
	b.WriteString(h1)
	b.WriteString(gap)
	b.WriteString(h2)
	b.WriteString(gap)
	b.WriteString(h3)
	b.WriteString("\n")
	b.WriteString(listTableBorderStyle.Render(strings.Repeat("─", sepW)))
	b.WriteString("\n")

	for i, v := range m.detailVersions {
		title := shortID(v.GetId())
		if i < len(m.detailVersionTitles) && m.detailVersionTitles[i] != "" {
			title = m.detailVersionTitles[i]
		}
		if v.GetId() == m.detailCurrentVersionID {
			title = currMark + title
		}
		ver := "?"
		if v.GetVersion() > 0 {
			ver = fmt.Sprintf("%d", v.GetVersion())
		}
		created := formatTs(v.GetCreatedAt())
		pref := "  "
		if i == m.detailVersionCursor {
			pref = "> "
		}
		vCell := padListCell(ver, listVersionColW)
		tCell := padListCell(title, tw)
		cCell := padListCell(created, listDateColW)
		line := vCell + gap + tCell + gap + cCell
		if i == m.detailVersionCursor {
			b.WriteString(listRowSelectedStyle.Render(pref+line) + "\n")
		} else {
			b.WriteString(pref)
			b.WriteString(listDateColStyle.Render(vCell))
			b.WriteString(gap)
			b.WriteString(listTitleColStyle.Render(tCell))
			b.WriteString(gap)
			b.WriteString(listDateColStyle.Render(cCell))
			b.WriteString("\n")
		}
	}
	return b.String()
}

// renderPayload рендерит payload в строку.
// reveal задаёт, показывать ли чувствительные поля (пароль и т.д.).
func renderPayload(p *clientdata.Payload, reveal bool) string {
	// создаем билдер для сборки строки
	var b strings.Builder

	fmt.Fprintf(&b, "Заголовок: %s\n", p.Title)

	switch p.Kind {
	// тип логин/пароль
	case clientdata.KindLoginPair:
		// По умолчанию пароль скрыт, если reveal = true (была нажата h), то показываем пароль
		pw := "***"

		if reveal {
			pw = p.Password
		}

		fmt.Fprintf(&b, "Тип: логин/пароль\nЛогин: %s\nПароль: %s\nURL: %s\n", p.Login, pw, p.URL)
	// тип текст
	case clientdata.KindText:
		fmt.Fprintf(&b, "Тип: текст\n%s\n", p.Text)
	// тип бинарные данные (base64)
	case clientdata.KindBinary:
		fmt.Fprintf(&b, "Тип: бинарные данные (base64)\n%s\n", trimMiddle(p.BinaryBase64, 120))
	// тип банковская карта
	case clientdata.KindBankCard:
		// По умолчанию CVC скрыт, если reveal = true (была нажата h), то показываем CVC
		cvc := "***"

		if reveal {
			cvc = p.CVC
		}

		fmt.Fprintf(&b, "Тип: банковская карта\nДержатель: %s\nНомер: %s\nСрок: %s\nCVC: %s\n", p.CardHolder, trimMiddle(p.CardNumber, 8), p.Expiry, cvc)
	}

	// добавляем метаданные, если они есть
	if p.Meta != "" {
		fmt.Fprintf(&b, "Метаданные: %s\n", p.Meta)
	}

	return b.String()
}

// trimMiddle обрезает строку по середине
func trimMiddle(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max/2] + "..." + s[len(s)-max/2:]
}

// shortID возвращает короткую версию ID
func shortID(id string) string {
	if len(id) <= 8 {
		return id
	}
	return id[:8] + "…"
}

// formatTs форматирует timestamp в строку
func formatTs(ts *timestamppb.Timestamp) string {
	if ts == nil {
		return ""
	}
	t := ts.AsTime()
	if t.IsZero() {
		return ""
	}
	return t.Local().Format("2006-01-02 15:04")
}
