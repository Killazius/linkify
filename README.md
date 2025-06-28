# Linkify

RESTFUL API сервис для укорачивания ссылок.
Проект находится в разработке и находится в стадии тестирования. 
Разработал [killazius](https://t.me/killazDev).

##### Версия: v1.5

## Технологии

- [chi](https://github.com/go-chi/chi) - многофункциональный маршрутизатор для Go. 
- [swagger](https://github.com/swaggo/swag) - спецификация OpenAPI для HTTP-серверов.
- [jwt](https://github.com/golang-jwt/jwt) - JSON Web Token
- [redis](https://github.com/redis/go-redis) - хранилище типа ключ-значение.
- [gorm](https://gorm.io/) - ORM для Go.
- [zap](https://github.com/uber-go/zap) - библиотека логирования.

## Конфигурация

### Настройка окружения

1. Создайте файл `.env` на основе примера `.env.example`:
```env
CONFIG_PATH="config/config.yaml"
SERVER_IP="localhost"
POSTGRES_USER="myuser"
POSTGRES_PASSWORD="mypassword"
POSTGRES_DB="mydb"
POSTGRES_HOST="postgres"
POSTGRES_PORT="5432"
JWT_SECRET="SECRETKEY"
ALIAS_LENGTH="6"

REDIS_ADDR="redis:6379"
REDIS_PASSWORD=""
REDIS_DB=0

GF_SECURITY_ADMIN_USER="admin"
GF_SECURITY_ADMIN_PASSWORD="admin"
```
   Где `<name>` — название вашего конфигурационного файла. (по умолчанию config/config.yaml)

2. Создайте конфигурационные файлы в [shortener](./shortener/README.md) и [auth](./auth/README.md).


## Использование Makefile
В проекте предоставлен `Makefile` для упрощения сборки и запуска проекта. Доступные команды:
- `make docker` - запуск docker-compose.
- `make swag` — генерация документации для сервиса.
- `make lint` - проверка кода на соответствие стандартам.
- `make test` - запуск тестов.


## Запуск проекта
```bash
git clone https://github.com/killazius/linkify.git # клонирование репозитория
make docker # запуск docker-compose
```
