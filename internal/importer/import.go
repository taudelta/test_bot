package importer

import (
	"database/sql"
	"encoding/json"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/taudelta/test_bot/internal/models"
)

func ImportTests(filePath, dbOutputFile string) error {
	log.Println("Начинается импорт базы данных вопросов и ответов в файл ", dbOutputFile)
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	log.Println("Файл с вопросами найден", filePath)

	fileBody, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var themes []models.Theme
	if err := json.Unmarshal(fileBody, &themes); err != nil {
		return err
	}

	log.Println("Файл проверен")

	db, err := sql.Open("sqlite3", dbOutputFile)
	if err != nil {
		return err
	}
	defer db.Close()

	log.Println("Производится очистка данных")

	if _, err := db.Exec("DELETE FROM answers"); err != nil {
		return err
	}
	if _, err := db.Exec("DELETE FROM questions"); err != nil {
		return err
	}
	if _, err := db.Exec("DELETE FROM themes"); err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
			log.Println(err)
		}
	}()

	for _, t := range themes {
		row := tx.QueryRow(`
			INSERT INTO themes (title) VALUES (?) RETURNING id
		`, t.Title)
		err = row.Err()
		if err != nil {
			return err
		}

		var themeID int64
		err = row.Scan(&themeID)
		if err != nil {
			return err
		}

		log.Printf("Добавлена тема: %s, id = %d\n", t.Title, themeID)

		for _, q := range t.QuestionList {
			questionRow := tx.QueryRow(
				`INSERT INTO questions (theme_id, txt) VALUES (?, ?) RETURNING id`,
				themeID, q.Text,
			)

			err = questionRow.Err()
			if err != nil {
				return err
			}

			log.Printf("Добавлен вопрос: %s\n", q.Text)

			var questionID int64
			err = questionRow.Scan(&questionID)
			if err != nil {
				return err
			}

			for _, a := range q.Answers {
				_, err = tx.Exec(
					"INSERT INTO answers (question_id, txt, valid) VALUES (?, ?, ?)",
					questionID, a.Text, a.IsValid,
				)
				if err != nil {
					return err
				}
				log.Printf("Добавлен ответ: %s, правильный: %t\n", a.Text, a.IsValid)
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	log.Println("Данные успешно импортированы")
	return nil
}
