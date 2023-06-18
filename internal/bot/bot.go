package bot

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/taudelta/test_bot/internal/models"
)

const (
	StateBeforeStartQASession = 1
	StateQASessionStarted     = 2
	StateLanguageSelection    = 3
)

func NewReplyKeyboardBasic(userID int64, language string) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(t(YesReplyMessage, userID), "/answer"),
			tgbotapi.NewInlineKeyboardButtonData(t(NoReplyMessage, userID), "/cancel"),
		),
	)
}

type Stat struct {
	ValidAnswers   int
	InvalidAnswers int
}

// Session represent an Question&Answer session
type Session struct {
	State         int
	Theme         models.Theme
	QuestionIndex int
	Statistics    Stat
}

type State struct {
	Value int
}

type UserProfile struct {
	Language string
}

var (
	states   = map[int64]*State{}
	sessions = map[int64]*Session{}
	profiles = map[int64]*UserProfile{}
)

// Start запуск бота, который читает входящие сообщения
// и обрабатывает команды
func Start(botAPI *tgbotapi.BotAPI) chan struct{} {
	loadLanguagePack()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := botAPI.GetUpdatesChan(u)

	doneCh := make(chan struct{})

	go func() {
		defer close(doneCh)
		for update := range updates {
			// check for a click to the button
			if update.CallbackQuery != nil {
				handleCallback(botAPI, update)
				continue
			}

			if update.Message == nil {
				continue
			}

			// received standart command
			handleMessage(botAPI, update)
		}
	}()

	return doneCh
}

func handleMessage(botAPI *tgbotapi.BotAPI, update tgbotapi.Update) {
	userID := update.Message.From.ID
	userName := update.Message.From.UserName
	message := update.Message.Text

	log.Printf("[user=%s][id=%d] %s\n", userName, userID, message)

	if _, ok := states[userID]; !ok {
		states[userID] = &State{
			Value: StateBeforeStartQASession,
		}
	}

	if _, ok := profiles[userID]; !ok {
		profiles[userID] = &UserProfile{
			Language: "en",
		}
	}

	if message == "/start" {
		// start session
		states[userID].Value = StateQASessionStarted
		sessions[userID] = &Session{}
		showThemes(botAPI, userID)
		return
	}

	if message == "/lang" || message == "/sl" || message == "/swl" {
		if states[userID].Value == StateBeforeStartQASession {
			// switch language command
			showLanguage(botAPI, userID)
			states[userID].Value = StateLanguageSelection
			return
		}
	}

	send(botAPI, userID, IncorrectCommand)
}

func handleCallback(botAPI *tgbotapi.BotAPI, update tgbotapi.Update) {
	callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
	if _, err := botAPI.Request(callback); err != nil {
		panic(err)
	}

	cmd := strings.Trim(callback.Text, " ")
	cmd = strings.Trim(cmd, "\n")
	cmd = strings.Trim(cmd, "\r")

	log.Println("Получена команда", cmd)

	userID := update.CallbackQuery.From.ID
	state := states[userID]
	switch state.Value {
	case StateLanguageSelection:
		handleLanguageCommand(botAPI, update, callback, cmd)
	default:
		handleQACommand(botAPI, update, callback, cmd)
	}
}

func handleLanguageCommand(botAPI *tgbotapi.BotAPI, update tgbotapi.Update, callback tgbotapi.CallbackConfig, cmd string) {
	userID := update.CallbackQuery.From.ID
	languageCode := cmd
	if languageCode != "ru" && languageCode != "en" {
		send(botAPI, userID, IncorrectLanguage)
		return
	}
	states[userID].Value = StateBeforeStartQASession
	profiles[userID].Language = languageCode
	send(botAPI, userID, ChangeLanguageSuccess)
}

func handleQACommand(botAPI *tgbotapi.BotAPI, update tgbotapi.Update, callback tgbotapi.CallbackConfig, cmd string) {
	userID := update.CallbackQuery.From.ID
	if cmd == "/answer" {
		handleCommand(botAPI, update, callback)
	} else if cmd == "/cancel" {
		delete(sessions, userID)
		send(botAPI, userID, GoodbyeMessage)
	} else {
		handleCommand(botAPI, update, callback)
	}
}

func showThemes(botAPI *tgbotapi.BotAPI, userID int64) {
	language := profiles[userID].Language

	themes := make([]*models.Theme, 0, len(cache[language]))
	for _, v := range cache[language] {
		themes = append(themes, v)
	}

	sort.Slice(themes, func(i, j int) bool {
		return themes[i].ID < themes[j].ID
	})

	msg := tgbotapi.NewMessage(userID, t(ThemeSelectionMessage, userID))

	var markup [][]tgbotapi.InlineKeyboardButton
	for index, t := range themes {
		txt := fmt.Sprintf("%d. %s", index+1, t.Title)
		row := tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(txt, fmt.Sprintf("%v", t.ID)))
		markup = append(markup, row)
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(markup...)
	msg.ReplyMarkup = keyboard

	if _, err := botAPI.Send(msg); err != nil {
		log.Println("send error", err)
	}
}

func showLanguage(botAPI *tgbotapi.BotAPI, userID int64) {
	msg := tgbotapi.NewMessage(userID, "Select language")

	languages := []struct {
		code string
		name string
	}{
		{
			code: "en",
			name: "English",
		},
		{
			code: "ru",
			name: "Russian",
		},
	}

	var markup [][]tgbotapi.InlineKeyboardButton
	for _, t := range languages {
		txt := fmt.Sprintf("%s", t.name)
		row := tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(txt, t.code))
		markup = append(markup, row)
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(markup...)
	msg.ReplyMarkup = keyboard

	if _, err := botAPI.Send(msg); err != nil {
		log.Println("send error", err)
	}
}

func startUserSession(botAPI *tgbotapi.BotAPI, update tgbotapi.Update, themeID int64) {
	userID := update.CallbackQuery.From.ID

	language := profiles[userID].Language

	var msgText string
	theme, ok := cache[language][themeID]
	if !ok {
		msgText = t(ThemeNotFoundMessage, userID)
	} else {
		log.Println("Тема занятия найдена в кеше")
		tpl := t(ThemeStartConfirmMessage, userID) + "%s"
		msgText = fmt.Sprintf(tpl, theme.Title)
	}

	msg := tgbotapi.NewMessage(userID, msgText)

	if ok {
		// показываем клавиатуру с кнопками Да/Нет
		msg.ReplyMarkup = NewReplyKeyboardBasic(userID, language)

		sessions[userID].Theme = *theme
		sessions[userID].QuestionIndex = -1
	}

	if _, err := botAPI.Send(msg); err != nil {
		log.Println("Ошибка", err)
	}

	session := sessions[userID]
	session.State = 1
}

func send(botAPI *tgbotapi.BotAPI, userID int64, text string) {
	msg := tgbotapi.NewMessage(userID, t(text, userID))
	if _, err := botAPI.Send(msg); err != nil {
		log.Println("Ошибка", err)
	}
}

func sendNoTranslate(botAPI *tgbotapi.BotAPI, userID int64, text string) {
	msg := tgbotapi.NewMessage(userID, text)
	if _, err := botAPI.Send(msg); err != nil {
		log.Println("Ошибка", err)
	}
}

func complete(botAPI *tgbotapi.BotAPI, userID int64, session *Session) {
	log.Println("Завершаем сессию пользователя", userID)
	delete(sessions, userID)
	confirmTemplate := fmt.Sprintf(
		"%s %s %s, %s %s",
		t(SessionEndStatisticsMessage, userID),
		"%d",
		t(RightAnswerCountMessage, userID),
		"%d",
		t(WrongAnswerCountMessage, userID),
	)

	txt := fmt.Sprintf(confirmTemplate, session.Statistics.ValidAnswers, session.Statistics.InvalidAnswers)
	states[userID].Value = StateBeforeStartQASession
	sendNoTranslate(botAPI, userID, txt)
}

func acceptAnswer(botAPI *tgbotapi.BotAPI, userID int64, answerText string, session *Session) {
	answerNumber, err := strconv.Atoi(answerText)
	if err != nil {
		send(botAPI, userID, ErrorMessage)
		return
	}

	lastQuestion := session.Theme.QuestionList[session.QuestionIndex]
	confirmedAnswer := lastQuestion.Answers[answerNumber-1]

	log.Println("пользователь дал ответ на вопрос: ", lastQuestion.Text)
	if confirmedAnswer.IsValid {
		log.Println("правильный - ", confirmedAnswer.Text)
		session.Statistics.ValidAnswers++
	} else {
		log.Println("неправильный - ", confirmedAnswer.Text)
		session.Statistics.InvalidAnswers++
	}
}

func handleCommand(botAPI *tgbotapi.BotAPI, update tgbotapi.Update, callback tgbotapi.CallbackConfig) {
	session, ok := sessions[update.CallbackQuery.From.ID]

	userID := update.CallbackQuery.From.ID
	if !ok {
		send(botAPI, userID, SessionNotFoundMessage)
		return
	}

	if session.State == 0 {
		themeID, _ := strconv.ParseInt(callback.Text, 10, 64)
		startUserSession(botAPI, update, themeID)
		return
	}

	log.Println("Получен ответ на вопрос #", session.QuestionIndex)

	if session.QuestionIndex >= 0 {
		acceptAnswer(botAPI, userID, callback.Text, session)
	}

	// получен ответ на последний вопрос
	if len(session.Theme.Questions)-1 == session.QuestionIndex {
		complete(botAPI, userID, session)
		return
	}

	session.QuestionIndex++

	// берем следующий вопрос
	question := session.Theme.QuestionList[session.QuestionIndex]

	if len(question.Answers) == 0 {
		send(botAPI, userID, AnswersNotFoundMessage)
		return
	}

	// формируем сообщение с клавиатурой
	msg := tgbotapi.NewMessage(update.CallbackQuery.From.ID, question.Text)

	var questionsMarkup [][]tgbotapi.InlineKeyboardButton
	for index, answer := range question.Answers {
		txt := fmt.Sprintf("%d. %s", index+1, answer.Text)
		row := tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(txt, fmt.Sprintf("%v", index+1)))
		questionsMarkup = append(questionsMarkup, row)
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(questionsMarkup...)

	msg.ReplyMarkup = keyboard
	if _, err := botAPI.Send(msg); err != nil {
		log.Println("Ошибка", err)
	}
}
