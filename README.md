# Tasks API

REST API для списка задач на Go (`net/http`) и PostgreSQL (`database/sql` + драйвер `pgx`).
Учебный проект для отработки HTTP, работы с базой данных, контекстов и конфигурации через переменные окружения.

## Возможности

| Метод    | Путь           | Описание                                  |
|----------|----------------|-------------------------------------------|
| `POST`   | `/tasks`       | Создать задачу                            |
| `GET`    | `/tasks`       | Получить список всех задач                |
| `GET`    | `/tasks/{id}`  | Получить задачу по id                     |
| `PATCH`  | `/tasks/{id}`  | Отметить задачу выполненной или нет       |
| `DELETE` | `/tasks/{id}`  | Удалить задачу                            |

## Модель задачи

```json
{
  "id": 1,
  "title": "Купить продукты",
  "done": false,
  "created_at": "2026-11-08T23:02:59.470011Z"
}
```

| Поле         | Описание                                                        |
|--------------|-----------------------------------------------------------------|
| `id`         | Выдаётся базой автоматически                                    |
| `title`      | Название, обязательное, не длиннее 100 символов                 |
| `done`       | Выполнена ли задача, по умолчанию `false`                       |
| `created_at` | Время создания, проставляется базой автоматически               |

## Тела запросов

Создание (`POST /tasks`):

```json
{"title": "Купить молоко"}
```

Обновление статуса (`PATCH /tasks/{id}`):

```json
{"done": true}
```

## Валидация

- `title` не должен быть пустым
- `title` не должен быть длиннее 100 символов (считаются символы, а не байты, поэтому русские названия проверяются корректно)
- `id` в пути должен быть числом
- тело запроса должно быть корректным JSON

## Коды ответов

| Код | Когда возвращается                                                         |
|-----|----------------------------------------------------------------------------|
| 200 | Успешное чтение или обновление                                             |
| 201 | Задача создана                                                             |
| 204 | Задача удалена (ответ без тела)                                            |
| 400 | Некорректный JSON, неверный `id` или не прошла валидация                   |
| 404 | Задача с таким `id` не найдена                                             |
| 500 | Ошибка на стороне сервера или базы данных                                  |

## Технологии и приёмы

- `net/http`: маршрутизация с указанием метода прямо в пути (Go 1.22+)
- `database/sql` + `pgx`: работа с PostgreSQL
- параметризованные запросы (`$1`, `$2`) для защиты от SQL-инъекций
- `INSERT ... RETURNING` и `UPDATE ... RETURNING` для получения готовой строки одним запросом
- `context` с таймаутом 3 секунды на каждый запрос к базе (`QueryContext`, `QueryRowContext`, `ExecContext`)
- конфигурация через переменные окружения и файл `.env` (`godotenv`)
- обработка ошибок с правильными HTTP-статусами, различие «не найдено» (`sql.ErrNoRows`) и ошибки базы

## Требования

- Go 1.22 или новее
- PostgreSQL
- Docker (необязательно, для запуска PostgreSQL в контейнере)

## Запуск

### 1. PostgreSQL

Пример запуска в Docker (порт `5433` на компьютере, `5432` внутри контейнера):

```bash
docker run --name my-postgres -e POSTGRES_PASSWORD=your_password -p 5433:5432 -d postgres:16
```

### 2. База и таблица

Создайте базу и таблицу, например через `psql` или pgAdmin:

```sql
CREATE DATABASE "tasksApi";
```

В созданной базе:

```sql
CREATE TABLE tasks (
    id SERIAL PRIMARY KEY,
    title VARCHAR(100) NOT NULL,
    done BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW()
);
```

### 3. Настройка окружения

Скопируйте `.env.example` в `.env` и укажите свои значения:

```
DB_HOST=localhost
DB_PORT=5433
DB_NAME=tasksApi
DB_USER=postgres
DB_PASSWORD=your_password
```

Файл `.env` добавлен в `.gitignore` и не попадает в репозиторий.

### 4. Запуск сервера

```bash
go run main.go
```

Сервер стартует на `http://localhost:8080`.

## Примеры запросов

### PowerShell

Создать задачу:

```powershell
Invoke-RestMethod -Method Post -Uri http://localhost:8080/tasks -ContentType "application/json" -Body '{"title":"Buy milk"}'
```

Для русского текста в Windows PowerShell 5.1 передавайте тело байтами в UTF-8:

```powershell
Invoke-RestMethod -Method Post -Uri http://localhost:8080/tasks -ContentType "application/json" -Body ([System.Text.Encoding]::UTF8.GetBytes('{"title":"Купить молоко"}'))
```

Получить все задачи:

```powershell
Invoke-RestMethod http://localhost:8080/tasks
```

Получить задачу по id:

```powershell
Invoke-RestMethod http://localhost:8080/tasks/1
```

Отметить задачу выполненной:

```powershell
Invoke-RestMethod -Method Patch -Uri http://localhost:8080/tasks/1 -ContentType "application/json" -Body '{"done":true}'
```

Удалить задачу:

```powershell
Invoke-RestMethod -Method Delete -Uri http://localhost:8080/tasks/1
```

### curl (Linux, macOS, Git Bash)

```bash
curl -X POST http://localhost:8080/tasks -d '{"title":"Купить молоко"}'
curl http://localhost:8080/tasks
curl http://localhost:8080/tasks/1
curl -X PATCH http://localhost:8080/tasks/1 -d '{"done":true}'
curl -X DELETE http://localhost:8080/tasks/1
```

## Что можно улучшить

- тесты обработчиков через `net/http/httptest`
- разделить код по пакетам (обработчики, работа с базой, конфигурация)
- логировать все ошибки базы, а не только часть
- плавное завершение сервера (graceful shutdown)
- `docker-compose` для запуска API и базы одной командой
- частичное обновление: возможность менять `title` через `PATCH` и различать «поле не прислали» и `false`
- миграции базы данных вместо ручного создания таблицы
