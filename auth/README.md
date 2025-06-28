# Auth Microservice

## Конфигурация

Создайте конфигурационный файл в папке config. Пример содержимого конфигурационного файла:
##### config/config.yaml
```yaml
logger_path: "config/logger.json"
grpc_server:
  port: 50051
  timeout: 5s
http_server:
  port: "8085"
  timeout: 5s
  idle_timeout: 60s
migrations_path: "migrations"
access_token_ttl: 15m
refresh_token_ttl: 24h
```
Создайте конфигурационный файл для логирования в папке config. Пример содержимого конфигурационного файла:
##### config/logger.json
```json
{
  "level": "debug",
  "encoding": "json",
  "outputPaths": ["stdout"],
  "errorOutputPaths": ["stderr"],
  "encoderConfig": {
    "timeKey": "timestamp",
    "timeEncoder": "rfc3339",
    "messageKey": "message",
    "levelKey": "level",
    "levelEncoder": "lowercase",
    "callerKey": "caller",
    "callerEncoder": "short"
  }
}
```

## Endpoints


- `POST /auth/register` - Регистрация нового пользователя

**Пример запроса:**
```json
{
  "email": "user@example.com",
  "password": "securepassword123"
}
```

**Пример ответа (201 Created):**
```json
{
  "user_id": 123
}
```
400 Bad Request: неверный формат запроса
409 Conflict: пользователь уже существует
500 Internal Server Error: ошибка сервера при регистрации

- `POST /auth/login` - Вход в систему

**Пример запроса:**
```json
{
  "email": "user@example.com",
  "password": "securepassword123"
}
```

**Пример ответа:**
```json
{
    "access_token_expires_in": 3600,
    "refresh_token_expires_in": 86400
}
```
Sets Cookies:
access_token: JWT for API authentication
refresh_token: JWT for token renewal

Error Responses:
400 Bad Request: неверный формат запроса
401 Unauthorized: неверные учетные данные
500 Internal Server Error:  ошибка сервера при входе

- `GET /auth/refresh` - Refresh токенов

Требуемые cookie:
refresh_token

**Пример ответа (200 OK):**
```json
{
    "access_token_expires_in": 3600,
    "refresh_token_expires_in": 86400
}
```
Updates Cookies:
access_token: New JWT for API authentication
refresh_token: New JWT for token renewal

Error Responses:
401 Unauthorized: отсутствующий или недействительный токен обновления
500 Internal Server Error: ошибка сервера при обновлении токенов

- `DELETE /auth/logout` - Выход из системы

Требуемые cookie:
refresh_token

**Пример ответа (204 No content)**

Clears Cookies:
access_token
refresh_token

Error Responses:
401 Unauthorized: отсутствует токен обновления
500 Internal Server Error: ошибка сервера при выходе из системы

- `DELETE /auth/account` - Удаление аккаунта

Требуемые cookie:
refresh_token

**Пример ответа (204 No content)**

Clears Cookies:
access_token
refresh_token

Error Responses:
401 Unauthorized: отсутствует или недействителен токен обновления
400 Bad Request: неверный идентификатор пользователя в токене
500 Internal Server Error: ошибка сервера при удалении учетной записи

