# Todo api

Simple todo api backend implemented in Go.

## Register user

```
curl -X POST -d '{"username":"userA", "password":"paSSword123"}' http://localhost:3000/register
```

## Login

```
curl -X POST -d '{"username":"userA", "password":"paSSword123"}' http://localhost:3000/login
```

## Access Todos

```
curl -H "X-Auth-Token: placeholder.jwt.token.for.userA" http://localhost:8080/users/userA/todos
```

## Create Todo

## Update Todo

## Delete Todo
