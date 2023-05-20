package bot

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"github.com/taudelta/test_bot/internal/models"
)

var cache = map[int64]*models.Theme{}

func FillQuestions(dbFile string) {
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	log.Println("Открываем базу данных с вопросами и ответами")
	rows, err := db.Query(`
select t.id as theme_id, t.title, q.id as question_id, q.txt, a.id as answer_id, a.txt, a.valid
from themes as t
left join questions as q on q.theme_id = t.id
left join answers as a on a.question_id = q.id
order by t.id, q.id, a.id
`)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

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

		if _, ok := cache[themeID]; !ok {
			log.Printf("Загружена тема (ID=%d): %s\n", themeID, title)
			cache[themeID] = &models.Theme{
				ID:        themeID,
				Title:     title,
				Questions: make(map[int64]*models.Question),
			}
		}

		if questionID != nil {
			qID := *questionID
			if _, ok := cache[themeID].Questions[qID]; !ok {
				log.Println("	Вопрос: ", *questionText)
				q := &models.Question{
					Text:    *questionText,
					Answers: make([]models.Answer, 0),
				}
				cache[themeID].Questions[qID] = q
				cache[themeID].QuestionList = append(cache[themeID].QuestionList, q)
			}
		}

		if questionID != nil && answerID != nil {
			log.Println("		Ответ: ", *answerText)
			answers := cache[themeID].Questions[*questionID].Answers
			answers = append(answers, models.Answer{
				Text:    *answerText,
				IsValid: *isValid,
			})

			validAnswers := make(map[int64]struct{})
			if *isValid {
				validAnswers[*answerID] = struct{}{}
			}

			cache[themeID].Questions[*questionID].Answers = answers
			cache[themeID].Questions[*questionID].ValidAnswers = validAnswers
		}
	}

	log.Println("Данные успешно прочитаны из базы данных")
}
