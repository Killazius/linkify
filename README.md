# Linkify

RESTFUL API сервис для укорачивания ссылок.
Проект находится в разработке и находится в стадии тестирования. 
Разработал [killazius](https://t.me/killazDev).

##### Версия: v1.2

## Технологии

- [chi](https://github.com/go-chi/chi) - многофункциональный маршрутизатор для Go. 
- [swagger](https://swagger.io/) - спецификация OpenAPI для HTTP-серверов.
- [redis](https://redis.io/) - хранилище типа ключ-значение.
- [gorm](https://gorm.io/) - ORM для Go.
- [uber-automaxprocs](https://github.com/uber-go/automaxprocs) - установка максимального количества потоков.
- [slog](https://pkg.go.dev/log/slog) - библиотека логирования.

## Конфигурация

### Настройка окружения

1. Создайте файл `.env` на основе примера `.env.example`:
```env
CONFIG_PATH = "config/<name>.yaml"


POSTGRES_USER="postgres_user"
POSTGRES_PASSWORD="postgres_password"
POSTGRES_DB="postgres_db"
POSTGRES_HOST="postgres"
POSTGRES_PORT="5432"

ALIAS_LENGTH="6"

REDIS_ADDR="redis:6379"
REDIS_PASSWORD=""
DB=0

STORAGE_URL="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}"

```
   Где `<name>` — название вашего конфигурационного файла. (по умолчанию config/config.yaml)

2. Создайте конфигурационный файл по пути, указанному в `.env`. Пример содержимого конфигурационного файла:
##### config/config.yaml
```yaml
env: "local" # "local", "prod"

http_server:
  address: "localhost:8080" # for localhost if env = prod "0.0.0.0:8080"
  timeout: "4s" # server RW timeout
  idle_timeout: "60s" # server idle timeout
```

## Использование Makefile
В проекте предоставлен `Makefile` для упрощения сборки и запуска проекта. Доступные команды:
- `make build` — сборка бинарного файла.
- `make exec` — запуск бинарного файла.
- `make clear` - очистка от бинарного файла.
- `make run` — запуск docker-compose.
- `make swag` — генерация документации для сервиса.
- `make lint` - проверка кода на соответствие стандартам.
- `make test` - запуск тестов.


## Endpoints

### URL

- `POST /url` - сохранение URL.

**Пример запроса:**
```json
{
    "url": "https://example.com"
}
```

**Пример ответа:**
```json
{
    "status": "OK",
    "alias": "123456",
    "created_at": "2023-06-01T00:00:00Z"
}
```
- `GET /url/{alias}` - перенаправление по сохраненному URL.

**Пример запроса:**
`curl http://localhost:8080/url/H2vga5`

**Пример ответа:**
`302 Found`

- `DELETE /url/{alias}` - удаление сохраненного URL.
**Пример запроса:**

`curl http://localhost:8080/url/H2vga5`

**Пример ответа:**
`204 No Content`


## Запуск проекта
```bash
git clone https://github.com/killazius/linkify.git # клонирование репозитория
cd shortener # переход в директорию проекта
make run # запуск docker-compose
```