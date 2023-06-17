# Бот для telegram, где абитуриент проходит тест

Разработано для школы программирования Golang5
#(https://vk.com/golang5)

## Команды

### /start
Начинает опрос

### /lang, /l, /swl, /sl
Переключение языка

## Запуск

- Создать бота в Telegram и получить уникальный ключ
- Создать чистую базу данных командой migrate
- Загрузить данные из файла с вопросами командой import
- Запустить Telegram бота

## Создание базы данных

В параметре db указывается путь к новой базе данных

### Команда

```bash
go run cmd\migrate\main.go --db=test.db
```

## Импорт данных

В параметре test_file указывается путь к файлу с вопросами
Создайте по образцу файл с вопросами и положите его в корневую директорию проекта

В параметре db указывается путь к новой базе данных

Пример файл приводится ниже

```bash
go run cmd\import\import.go --test_file=test.json --db=test.db
```

## Запуск Telegram бота

Параметр bot_token - уникальный ключ, полученный при создании бота в Telegram

В параметре db указывается путь к базе данных с вопросами

```bash
go run cmd\bot\main.go --db=test.db --bot_token=token
```

## Пример файла с вопросами

```javascript
[
    {
        "theme": "Theme1",
        "code": "theme1",
        "questions": [
            {
                "text": "Question1",
                "answers": [
                    {
                        "text": "Answer1",
                        "valid": true
                    },
                    {
                        "text": "Answer2"
                    }
                ]
            },
            {
                "text": "Question2",
                "answers": [
                    {
                        "text": "Answer1"
                    },
                    {
                        "text": "Answer2",
                        "valid": true
                    }
                ]
            }
        ]
    }
]
```

## Инструкция для разработчика

### Запуск линтера

Установить golangci-lint

```bash
go get github.com/golangci/golangci-lint/cmd/golangci-lint
```

```bash
golangci-lint run
```
