## Запуск

1. Скопировать .env.example в .env и заполнить;
2. docker compose -f docker-compose.yml up -d;
3. Зайти в контейнер с приложением (backend) и выполнить миграции ```./backend -s```.

## Запуск в режиме разработки
1. Скопировать .env.example в .env и заполнить;
2. docker compose -f docker-compose.yml run -p 5432:5432 db;
3. Запустить приложение, например, по F5 через VSCode.

## Текущий статус
Есть апи для логина и получения сессии

## Роадмап
not implemented yet


