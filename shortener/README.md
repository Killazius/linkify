# Shortener Microservice

## Конфигурация

Создайте конфигурационный файл в папке config. Пример содержимого конфигурационного файла:
##### config/config.yaml
```yaml
http_server:
  address: "0.0.0.0:8080"
  timeout: "4s"
  idle_timeout: "60s"
  alias_length: 6
prometheus:
  address: "0.0.0.0:8083"
  timeout: "4s"
  idle_timeout: "60s"
logger_path: "config/logger.json"
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

### URL

- `POST /api/url` - сохранение URL.

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
  "alias": "H2vga5",
  "created_at": "2023-06-01T00:00:00Z"
}
```
- `GET /{alias}` - перенаправление по сохраненному URL.

**Пример запроса:**
`GET /H2vga5`

**Пример ответа:**
`302 Found
Location: https://original-url.com`

- `DELETE /api/url/{alias}` - удаление сохраненного URL.
**Пример запроса:**

`DELETE /api/url/H2vga5`

**Пример ответа:**
`204 No Content`
