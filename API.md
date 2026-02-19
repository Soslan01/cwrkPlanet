# Документация API

## Базовый URL

- **Docker**: `http://localhost:3000/api` (через nginx)
- **Локальная разработка**: `http://localhost:3000/api` (через Vite proxy)
- **Напрямую**: `http://localhost:8080` (api-gateway)

Все эндпоинты ниже указаны относительно базового URL. Клиент по умолчанию использует префикс `/api`, который nginx/Vite пробрасывает на api-gateway.

## Аутентификация

API использует JWT: access token и refresh token. Access token передаётся в заголовке:

```
Authorization: Bearer <access_token>
```

При истечении access token клиент может обновить пару токенов через `POST /v1/auth/refresh`.

---

## Эндпоинты

### 1. Регистрация

**POST** `/v1/auth/register`

Создаёт нового пользователя и возвращает токены.

**Тело запроса:**
```json
{
  "email": "string",           // обязательное, email
  "username": "string",        // обязательное, 3-50 символов
  "password": "string",        // обязательное, минимум 6 символов
  "displayName": "string",     // обязательное
  "avatarUrl": "string"        // опционально
}
```

**Ответ 200 OK:**
```json
{
  "accessToken": "string",
  "refreshToken": "string",
  "user": {
    "id": 0,
    "email": "string",
    "username": "string",
    "displayName": "string",
    "avatarUrl": "string | null",
    "createdAt": 0,
    "updatedAt": 0
  }
}
```

**Коды ошибок:**
- `400` — неверный формат запроса
- `409` — пользователь уже существует (email или username заняты)

---

### 2. Вход

**POST** `/v1/auth/login`

Аутентифицирует пользователя по email и паролю.

**Тело запроса:**
```json
{
  "email": "string",     // обязательное
  "password": "string"   // обязательное
}
```

**Ответ 200 OK:**
```json
{
  "accessToken": "string",
  "refreshToken": "string",
  "user": {
    "id": 0,
    "email": "string",
    "username": "string",
    "displayName": "string",
    "avatarUrl": "string | null",
    "createdAt": 0,
    "updatedAt": 0
  }
}
```

**Коды ошибок:**
- `400` — неверный формат запроса
- `401` — неверные учётные данные

---

### 3. Обновление токенов

**POST** `/v1/auth/refresh`

Выдаёт новую пару access и refresh токенов по действующему refresh token. Не требует авторизации.

**Тело запроса:**
```json
{
  "refreshToken": "string"   // обязательное
}
```

**Ответ 200 OK:**
```json
{
  "accessToken": "string",
  "refreshToken": "string",
  "expires_in": 0   // секунды до истечения access token
}
```

**Коды ошибок:**
- `400` — неверный формат запроса
- `401` — недействительный или истёкший refresh token

---

### 4. Выход

**POST** `/v1/auth/logout`

Инвалидирует сессию по refresh token. Не требует заголовка Authorization.

**Тело запроса:**
```json
{
  "refreshToken": "string"   // обязательное
}
```

**Ответ 200 OK:**
```json
{
  "success": true
}
```

---

### 5. Текущий пользователь

**GET** `/v1/auth/me`

Возвращает данные авторизованного пользователя.

**Заголовки:**
```
Authorization: Bearer <access_token>
```

**Ответ 200 OK:**
```json
{
  "user": {
    "id": 0,
    "email": "string",
    "username": "string",
    "displayName": "string",
    "avatarUrl": "string | null",
    "createdAt": 0,
    "updatedAt": 0
  }
}
```

**Коды ошибок:**
- `401` — отсутствует или недействителен токен

---

### 6. Проверка состояния

**GET** `/health`

Проверка доступности api-gateway. Не требует авторизации.

**Ответ 200 OK:**
```
OK
```

---

## Формат ошибок

При ошибках API возвращает JSON:

```json
{
  "error": "Описание ошибки",
  "details": "Дополнительные детали (опционально)"
}
```

## Общие коды состояния

| Код | Значение |
|-----|----------|
| 200 | Успех |
| 400 | Неверный запрос |
| 401 | Требуется аутентификация / токен недействителен |
| 403 | Доступ запрещён |
| 404 | Не найдено |
| 500 | Внутренняя ошибка сервера |
| 503 | Сервис недоступен |

## CORS

Разрешены все origins (`*`). Методы: GET, POST, PUT, DELETE, OPTIONS. Заголовки: Accept, Authorization, Content-Type, X-Request-ID. Credentials разрешены.
