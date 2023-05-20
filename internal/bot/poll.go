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

var arrayKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Да", "/answer"),
		tgbotapi.NewInlineKeyboardButtonData("Нет", "/cancel"),
	),
)

type Stat struct {
	ValidAnswers   int
	InvalidAnswers int
}

type Session struct {
	State         int
	Theme         models.Theme
	QuestionIndex int
	Statistics    Stat
}

var sessions = map[int64]*Session{}

func handleCallback(botAPI *tgbotapi.BotAPI, update tgbotapi.Update) {
	callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
	if _, err := botAPI.Request(callback); err != nil {
		panic(err)
	}

	cmd := strings.Trim(callback.Text, " ")
	cmd = strings.Trim(cmd, "\n")
	cmd = strings.Trim(cmd, "\r")

	log.Println("Получена команда", cmd)

	if cmd == "/answer" {
		handleCommand(botAPI, update, callback)
	} else if cmd == "/cancel" {
		delete(sessions, update.CallbackQuery.From.ID)
		send(botAPI, update.CallbackQuery.From.ID, "До свидания. Всего хорошего.")
	} else {
		handleCommand(botAPI, update, callback)
	}
}

// Start запуск бота, который читает входящие сообщения
// и обрабатывает команды
func Start(botAPI *tgbotapi.BotAPI) chan struct{} {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := botAPI.GetUpdatesChan(u)

	doneCh := make(chan struct{})

	go func() {
		defer close(doneCh)
		for update := range updates {
			if update.CallbackQuery != nil {
				handleCallback(botAPI, update)
				continue
			}

			if update.Message == nil {
				continue
			}

			userID := update.Message.From.ID
			userName := update.Message.From.UserName
			message := update.Message.Text

			log.Printf("[%s][id=%d] %s\n", userName, userID, message)

			if message == "/start" {
				sessions[update.Message.From.ID] = &Session{}
				showThemes(botAPI, userID)
			} else {
				send(botAPI, userID, "Неверная команда. Для прохождения опроса введите /start")
			}
		}
	}()

	return doneCh
}

func showThemes(botAPI *tgbotapi.BotAPI, userID int64) {
	themes := make([]*models.Theme, 0, len(cache))
	for _, v := range cache {
		themes = append(themes, v)
	}

	sort.Slice(themes, func(i, j int) bool {
		return themes[i].ID < themes[j].ID
	})

	msg := tgbotapi.NewMessage(userID, "Выберите тему для тестирования")

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

func startUserSession(botAPI *tgbotapi.BotAPI, update tgbotapi.Update, themeID int64) {
	msgTemplate := "Давайте проведем тестирование по теме: %s"

	userID := update.CallbackQuery.From.ID

	var msgText string
	theme, ok := cache[themeID]
	if !ok {
		msgText = "Заданная тема не найдена"
	} else {
		log.Println("Тема занятия найдена в кеше")
		msgText = fmt.Sprintf(msgTemplate, theme.Title)
	}

	msg := tgbotapi.NewMessage(userID, msgText)

	if ok {
		// показываем клавиатуру с кнопками Да/Нет
		msg.ReplyMarkup = arrayKeyboard

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
	msg := tgbotapi.NewMessage(userID, text)
	if _, err := botAPI.Send(msg); err != nil {
		log.Println("Ошибка", err)
	}
}

func complete(botAPI *tgbotapi.BotAPI, userID int64, session *Session) {
	log.Println("Завершаем сессию пользователя", userID)
	delete(sessions, userID)
	confirmTemplate := "Спасибо что прошли опрос. Ваш результат: %d правильных ответов, %d неправильных"
	txt := fmt.Sprintf(confirmTemplate, session.Statistics.ValidAnswers, session.Statistics.InvalidAnswers)
	send(botAPI, userID, txt)
}

func acceptAnswer(botAPI *tgbotapi.BotAPI, userID int64, answerText string, session *Session) {
	answerNumber, err := strconv.Atoi(answerText)
	if err != nil {
		send(botAPI, userID, "Ошибка")
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
		send(botAPI, userID, "Сессия не найдена")
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
		send(botAPI, userID, "Ошибка. Ответы не найдены")
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
