# Todo api

demo: https://supercut.ai/share/FhIpMNhxB-t1L9k0Am3xzx

Simple todo api backend implemented in Go.

## Register user

```
curl -X POST -d '{"username":"userA", "password":"paSSword123"}' http://localhost:3000/register
```

## Login

```
curl -X POST -d '{"username":"userA", "password":"paSSword123"}' http://localhost:3000/login
```

## Logout

```
curl --request POST \
  --url {endpoint}/logout \
  --header 'Authorization: Bearer auth-token' \
  --header 'content-type: application/json' \
  --data '{
  "username": "userA",
  "password": "paSSword123"
}'
```

## Access Todos

```
curl -H "X-Auth-Token: placeholder.jwt.token.for.userA" http://localhost:8080/users/:userA/todos
```

## Create Todo

```
curl --request POST \
  --url {endpoint}/users/:userA/todos \
  --header 'Authorization: Bearer auth-token' \
  --header 'content-type: application/json' \
  --data '{
  "task": "Build JWT API"
}'
```

## Update Todo

```
curl --request PUT \
  --url {endpoint}/users/:userA/todos/:todoId \
  --header 'Authorization: Bearer auth-token' \
  --header 'content-type: application/json' \
  --data '{
  "task": "Build JWT API",
  "completed": true
}'
```

## Delete Todo

```
curl --request DELETE \
  --url {endpoint}/users/:userA/todos/:todoId \
  --header 'Authorization: Bearer auth-token' \
  --header 'content-type: application/json' \
  --data '{
  "task": "Build JWT APIs"
}'
```
