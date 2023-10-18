# go-musthave-metrics-tpl

Шаблон репозитория для трека «Сервер сбора метрик и алертинга».

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m main template https://github.com/Yandex-Practicum/go-musthave-metrics-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/main .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).

Запуск postgres для тестов:
```bash
docker run -d \                                                                                                                       ✔ 
        --name postgre-go \
        -e POSTGRES_PASSWORD=secret \
        -e PGDATA=/var/lib/postgresql/data/pgdata \
        -v /var/tmp:/var/lib/postgresql/data \
        -p 5432:5432 postgres:15

psql postgres://postgres:secret@localhost:5432

postgres=# create database test;
postgres=# create user test with encrypted password 'test';
postgres=# grant all privileges on database metrics to test;
postgres=# grant all on SCHEMA public TO test;

psql postgres://test:test@localhost:5432/test

postgres=# 
CREATE TABLE Metrics (
"name" VARCHAR(250) NOT NULL,
"type" VARCHAR(100) NOT NULL,
"delta" DOUBLE PRECISION,
"value" BIGINT,
CONSTRAINT ID PRIMARY KEY (name,type) );

```

## Сборка 
```bash
go build -ldflags \
"-X main.buildVersion=v0.0.1 \
-X main.buildDate=$(date +'%d/%m/%Y') \
-X main.buildCommit=$(git rev-parse --short HEAD)"\
-o agent  ./cmd/agent/main.go 
```

## Создание ключей

```bash
go run ./cmd/keygen/main.go

INFO    asymcrypt/asymcrypt.go:46       Created Pivate key: key_priv.pem
INFO    asymcrypt/asymcrypt.go:47       Created PublicKey key: key_pub.pem
```
