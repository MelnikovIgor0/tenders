# Как запустить:

1) В .env файле правите переменные (DSN и порт на акутальные)

2) Запускаете докер-контейнер с базой (если у вас сторонняя база, то скипайте этот пункт): 

```sh
docker compose up -d
```

3) Применяете гусем (или чем хотите еще) миграции из директории backend/migrations/

```sh
goose -dir migrations postgres "{тут ваш DSN}" up
```

Только обратите внимание, там миграция init создает таблички и типы, а следующая - генерит какие-то данные в таблички, чтобы было проще функционал тестить. Если не хотите, вторую миграцию можете не применять, но первую надо по-любому.

4) Если есть желание, потыкайте в backend/config.yaml настройки конэкшна к бд.

5) Запускаете backend/main.go и на указаном в .env адресе крутится сервак.

PS: ручки как в описании, но добавил еще ручку /api/bids/:bidId/get_decision, чтобы все-таки решение по предложению можно было получить, не лазия в бд.
