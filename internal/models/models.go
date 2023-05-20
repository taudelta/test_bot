package models

// Theme тема опроса
type Theme struct {
	ID           int64               `json:"id"`
	Title        string              `json:"theme"`
	Code         string              `json:"code"`
	Questions    map[int64]*Question `json:"-"`
	QuestionList []*Question         `json:"questions"`
}

// Question вопрос
type Question struct {
	Text         string             `json:"text"`
	Answers      []Answer           `json:"answers"`
	ValidAnswers map[int64]struct{} `json:"-"`
}

// Answer ответ
type Answer struct {
	Text    string `json:"text"`
	IsValid bool   `json:"valid"`
}
