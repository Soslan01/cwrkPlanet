**Данные, которые хранятся:**. 
- users.password_hash - хэш пароля с использованием bcrypt. Паролей в чистом виде в БД нет.  
- auth_sessions.token_hash - SHA-256 от `refresh_token` (сам refresh_token не хранится, только его   хэш). 
  
Access token(короткоживущий). 
- Формат: JWT(256). 
- Клеймы: `sub` (user_id), `iss`(issuer), `aud` (audience), `iat/nbf/exp`. 
- Срок жизни: задается `security.accessTTL` в конвиге. 
- Подпись: общий секрет в конфиге `SEC_JWT_SECRET`. 
- Проверка:  
    1. Разобрать JWT и проверить метод подписи == HS256. 
    2. Проверить подписьпо SEC_JWT_SECRET. 
    3. Проверить `iss`, `aud`, `exp/nbf/iat`. 
    4. Извлечь `sub` -> `user_id`. 
Refresh token(долгоживущий). 
- Формат: `opaque` случайная строка (примерно 32 байта, URL-safe base64). 
- Срок жизни: задается в конфиге `security.refreshTTL`. 
- Хранение: только хэш (SHA-256) в таблице `auth_sessions` + срок действия и метаданные. 
- Проверка: сверяем SHA-256(refresh_token) c `token_hash` в БД + проверяем `expires_at`. 
  
Почему не хранится как есть:  
**Пароль** - нельзя хранить в чистом виде - только хэш (например bcrypt) с параметрами (сейчас по умолчанию). 
**Refresh** - нельзя хранить в чистом виде - атакующий сможет использовать refresh без знания первичного зачения(поэтому хранится только SHA-256). 

Базовый адрес для сервиса: localhost:8081

**Endpoint**

Register - регистрация нового пользователя

POST /v1/auth/register

body:
```
{
  "email": "user1@example.com",
  "password": "123456",
  "displayName": "User One"
}
```

resp example:
```
{
  "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refreshToken": "SB1U1zRyx_qdDkAqhygziMPxP2Dlpsqxy0lPpFzGn1Y",
  "user": {
    "id": 1,
    "email": "user1@example.com",
    "displayName": "User One",
    "emailVerified": false,
    "createdAt": 1730000000,
    "updatedAt": 1730000000
  }
}
```

Login - вход по email и паролю

POST /v1/auth/login

body:
```
{
  "email": "user1@example.com",
  "password": "123456"
}
```

resp example:
```
{
  "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refreshToken": "E8yZwA31lDyROsVh5bOV9RW4v7JrsIvS24uxY7frjMA",
  "user": {
    "id": 1,
    "email": "user1@example.com",
    "displayName": "User One",
    "emailVerified": false,
    "createdAt": 1730000000,
    "updatedAt": 1730000000
  }
}
```

Refresh - обновление access-токена

POST /v1/auth/refresh

body:
```
{
  "refreshToken": "E8yZwA31lDyROsVh5bOV9RW4v7JrsIvS24uxY7frjMA"
}
```

resp example:
```
{
  "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refreshToken": "T3pN4fPkKhi3YH6t5yzUJ0I7e4bbixE8Z3oS6mWPEgY",
  "expires_in": "600"
}
```

Me - профиль текущего пользователя

GET /v1/auth/me

headers:
```
Authorization: Bearer <accessToken>
```

resp:
```
{
  "user": {
    "id": 1,
    "email": "user1@example.com",
    "displayName": "User One",
    "emailVerified": false,
    "createdAt": 1730000000,
    "updatedAt": 1730000000
  }
}
```