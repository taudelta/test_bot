package bot

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

const (
	ErrorMessage                = "Something went wrong"
	SessionNotFoundMessage      = "Session not found"
	IncorrectCommand            = "IncorrectCommand"
	IncorrectLanguage           = "IncorrectLanguage"
	ChangeLanguageSuccess       = "ChangeLanguageSuccess"
	GoodbyeMessage              = "GoodbyeMessage"
	ThemeSelectionMessage       = "ThemeSelectionMessage"
	ThemeStartConfirmMessage    = "ThemeStartConfirmMessage"
	ThemeNotFoundMessage        = "ThemeNotFoundMessage"
	SessionEndStatisticsMessage = "SessionEndStatisticsMessage"
	RightAnswerCountMessage     = "RightAnswerCountMessage"
	WrongAnswerCountMessage     = "WrongAnswerCountMessage"
	AnswersNotFoundMessage      = "AnswersNotFoundMessage"
)

type LanguagePack struct {
	Language     string
	Translations map[string]string
}

var langPacks = map[string]LanguagePack{}

func t(s string, userID int64) string {
	profile := profiles[userID]
	log.Println("translation", profile.Language, s, langPacks[profile.Language].Translations[s])
	t := langPacks[profile.Language].Translations[s]
	return t
}

func loadLanguagePack() {
	var loadedLanguages []string

	err := filepath.Walk("./internal/bot/translations", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		var langPack LanguagePack

		if err := json.NewDecoder(f).Decode(&langPack); err != nil {
			return err
		}

		if langPack.Language == "" {
			return fmt.Errorf("language pack language is empty")
		}

		langPacks[langPack.Language] = langPack
		loadedLanguages = append(loadedLanguages, langPack.Language)

		log.Println("load translation pack", langPack.Language)
		return nil
	})

	if err != nil {
		log.Println("error loading translations:", err)
		return
	}

	log.Println("language packs", loadedLanguages)
}
