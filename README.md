# Neighbor Finder Chat Service

Чат сервис для приложения рекомендации соседей.
Данный сервис служит для обмена сообщений между пользователями одной группы.

## Установка

### Зависимости

Для запуска сервиса требуется установить следующие пакеты:

    - golang
    - docker
    - docker-compose

### Переменные окружения

    - CONFIG_PATH - путь к файлу конфигурации (рекомендуется использовать из config)

    остальные переменные окружения используются если нет конфигурационного файла

    - GRPC_HOST - адрес сервера для gRPC
    - GRPC_PORT - порт сервера для gRPC
    - USER_SERVICE_HOST - адрес сервера для User Service
    - USER_SERVICE_PORT - порт сервера для User Service
    - GROUP_SERVICE_HOST - адрес сервера для Group Service
    - GROUP_SERVICE_PORT - порт сервера для Group Service
    - POSTGRES_HOST - адрес сервера для Postgres
    - POSTGRES_PORT - порт сервера для Postgres
    - POSTGRES_USER - пользователь Postgres
    - POSTGRES_PASSWORD - пароль Postgres
    - POSTGRES_DB_NAME - имя базы данных Postgres

### Запуск

Запуск через docker-compose (рекомендуется)

```bash
git clone https://github.com/savelij/nbf-chat-service.git
cd nbf-chat-service
docker-compose up -d
```

Или запуск через golang

```bash
git clone https://github.com/savelij/nbf-chat-service.git
cd nbf-chat-service
go run cmds/main.go
```

Через Taskfile

```bash
git clone https://github.com/savelij/nbf-chat-service.git
cd nbf-chat-service
task run
```

## Использование

После запуска сервиса доступен на порту `50052` по протоколу gRPC

На сервер установлен Reflection для определния функций без protobuf.

Данная версия является тестовой, поэтому все сервисы замоканы.

### User Service

Пользователи:

```json
[
    {
        "id": "d5313639-46cf-42d1-9c23-1cd19c8dcfb9",
        "name": "test1"
    },
    {
        "id": "d5313639-46cf-42d1-9c23-1cd19c8dcfb8",
        "name": "test2"
    },
    {
        "id": "d5313639-46cf-42d1-9c23-1cd19c8dcfb7",
        "name": "test3"
    }
]
```

### Group Service

Группы:

`id` - любое

```json

{
    "id": "d5313639-46cf-42d1-9c23-1cd19c8dcfb9",
    "name": "test",
    "members": [
        "d5313639-46cf-42d1-9c23-1cd19c8dcfb9",
        "d5313639-46cf-42d1-9c23-1cd19c8dcfb8",
        "d5313639-46cf-42d1-9c23-1cd19c8dcfb7"
    ]
}
```

### Message Repository

Сообщения в данной версии хранятся in-memory.
