# API Documentation

## Authentication Service (Port: 8080)

### Register User
```http
POST http://localhost:8080/api/v1/auth/register
Content-Type: application/json

{
    "username": "string",
    "email": "string@gmail.com",
    "password": "string"
}
```

### Login User
```http
POST http://localhost:8080/api/v1/auth/login
Content-Type: application/json

{
    "email": "string@gmail.com",
    "password": "string"
}
```

### Refresh Token
```http
POST http://localhost:8080/api/v1/auth/refresh
Content-Type: application/json

{
    "refresh_token": "string"
}
```

### Logout User
```http
POST http://localhost:8080/api/v1/auth/logout
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
    "refresh_token": "string"
}
```

### Validate Token
```http
GET http://localhost:8080/api/v1/auth/validate
Authorization: Bearer <jwt_token>
```

## Forum Service (Port: 8081)

### Create Message
```http
POST http://localhost:8081/api/v1/messages
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
    "content": "string"
}
```

### Get All Messages
```http
GET http://localhost:8081/api/v1/messages
Authorization: Bearer <jwt_token>
```

### Get Message by ID
```http
GET http://localhost:8081/api/v1/messages/{message_id}
Authorization: Bearer <jwt_token>
```

### Update Message
```http
PUT http://localhost:8081/api/v1/messages/{message_id}
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
    "content": "string"
}
```

### Delete Message
```http
DELETE http://localhost:8081/api/v1/messages/{message_id}
Authorization: Bearer <jwt_token>
```

### Create Comment
```http
POST http://localhost:8081/api/v1/messages/{message_id}/comments
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
    "content": "string"
}
```

### Get Comments for Message
```http
GET http://localhost:8081/api/v1/messages/{message_id}/comments
Authorization: Bearer <jwt_token>
```

### Update Comment
```http
PUT http://localhost:8081/api/v1/comments/{comment_id}
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
    "content": "string"
}
```

### Delete Comment
```http
DELETE http://localhost:8081/api/v1/comments/{comment_id}
Authorization: Bearer <jwt_token>
```

## Notes

1. All endpoints requiring authentication need a valid JWT token in the Authorization header
2. JWT tokens can be obtained through the login endpoint
3. Refresh tokens can be used to get new JWT tokens when they expire
4. All timestamps are in UTC
5. All IDs are UUIDs
6. Error responses follow the format:
```json
{
    "error": "error message"
}
```

## Response Codes

- 200: Success
- 201: Created
- 400: Bad Request
- 401: Unauthorized
- 403: Forbidden
- 404: Not Found
- 500: Internal Server Error 