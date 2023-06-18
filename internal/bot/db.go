package bot

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/taudelta/test_bot/internal/models"
)

var cache = map[string]map[int64]*models.Theme{}

const questionsQuery = `
select t.id as theme_id, t.title, q.id as question_id, q.txt, a.id as answer_id, a.txt, a.valid
from themes as t
left join questions as q on q.theme_id = t.id
left join answers as a on a.question_id = q.id
order by t.id, q.id, a.id
`

func LoadData(dbPath string) error {
	return filepath.Walk(dbPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if filepath.Ext(info.Name()) != ".db" {
			return nil
		}

		language := strings.ReplaceAll(info.Name(), ".db", "")

		FillQuestions(path, language)
		return nil
	})
}

func FillQuestions(dbFile, language string) {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	log.Println("Открываем базу данных с вопросами и ответами")
	rows, err := db.Query(questionsQuery)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	if _, ok := cache[language]; !ok {
		cache[language] = map[int64]*models.Theme{}
	}

	for rows.Next() {
		var (
			themeID      int64
			title        string
			questionID   *int64
			questionText *string
			answerID     *int64
			answerText   *string
			isValid      *bool
		)

		err = rows.Scan(&themeID, &title, &questionID, &questionText, &answerID, &answerText, &isValid)
		if err != nil {
			panic(err)
		}

		if _, ok := cache[language][themeID]; !ok {
			log.Printf("Загружена тема (ID=%d): %s\n", themeID, title)
			cache[language][themeID] = &models.Theme{
				ID:        themeID,
				Title:     title,
				Questions: make(map[int64]*models.Question),
			}
		}

		if questionID != nil {
			qID := *questionID
			if _, ok := cache[language][themeID].Questions[qID]; !ok {
				log.Println("	Вопрос: ", *questionText)
				q := &models.Question{
					Text:    *questionText,
					Answers: make([]models.Answer, 0),
				}
				cache[language][themeID].Questions[qID] = q
				cache[language][themeID].QuestionList = append(cache[language][themeID].QuestionList, q)
			}
		}

		if questionID != nil && answerID != nil {
			log.Println("		Ответ: ", *answerText)
			answers := cache[language][themeID].Questions[*questionID].Answers
			answers = append(answers, models.Answer{
				Text:    *answerText,
				IsValid: *isValid,
			})

			validAnswers := make(map[int64]struct{})
			if *isValid {
				validAnswers[*answerID] = struct{}{}
			}

			cache[language][themeID].Questions[*questionID].Answers = answers
			cache[language][themeID].Questions[*questionID].ValidAnswers = validAnswers
		}
	}

	log.Println("Данные успешно прочитаны из базы данных")
}
