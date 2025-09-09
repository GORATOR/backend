# Client Reports Support

Этот backend теперь поддерживает client reports - специальные сообщения, которые отправляют клиентские библиотеки Sentry, когда не могут доставить обычные события из-за недоступности сервера.

## Что такое Client Reports

Client reports - это механизм, позволяющий клиентским библиотекам Sentry информировать сервер о событиях, которые не удалось отправить. Это происходит когда:
- Backend недоступен во время отправки события
- Произошла сетевая ошибка
- Превышен таймаут запроса

## Формат Client Report

Client report состоит из трех частей, разделенных символом новой строки:

1. **Header**: `{}` - пустой JSON объект
2. **Type**: `{"type":"client_report"}` - указывает тип сообщения
3. **Data**: JSON с информацией о потерянных событиях

Пример:
```
{}
{"type":"client_report"}
{"timestamp":1740239731.564,"discarded_events":[{"reason":"network_error","category":"internal","quantity":1}]}
```

## Обработка Client Reports

Backend автоматически определяет client reports по наличию `"type":"client_report"` во второй строке payload'а.

### Процесс обработки:

1. Парсинг payload и определение типа сообщения
2. Извлечение ключа проекта из URL параметров или заголовков
3. Проверка существования проекта и валидности ключа
4. Сохранение client report в базу данных
5. Возврат пустого JSON ответа `{}`

### Модель данных

Client reports сохраняются в таблице `client_reports` со следующими полями:
- `id` - уникальный идентификатор
- `project_id` - ID проекта
- `timestamp` - временная метка события
- `discarded_events` - JSON массив с информацией о потерянных событиях
- `envelope_key` - ключ проекта для валидации
- `created_at`, `updated_at`, `deleted_at` - стандартные поля GORM

### Поля discarded_events:
- `reason` - причина потери (например, "network_error")
- `category` - категория события (например, "internal")
- `quantity` - количество потерянных событий

## Тестирование

Используйте скрипт `test_client_report.sh` для тестирования функциональности:

```bash
./test_client_report.sh
```

Или отправьте client report вручную:

```bash
curl -v \
  -X POST \
  -H "Content-Type: text/plain" \
  -d $'{}
{"type":"client_report"}
{"timestamp":1740239731.564,"discarded_events":[{"reason":"network_error","category":"internal","quantity":1}]}' \
  "http://localhost:8080/api/1/envelope/?sentry_key=YOUR_PROJECT_KEY&sentry_version=7"
```

## Воспроизведение сценария из issue

1. Создайте проект через API:
```bash
curl http://localhost:8080/project -d '{"Name":"Project test","TeamId": 1,"Avatar":"","Active":true}'
```

2. Обновите DSN в клиентском коде, используя полученный `envelope_key`

3. Запустите клиент с выключенным backend'ом

4. Через несколько секунд включите backend - client report будет обработан корректно

## Совместимость

Данная реализация совместима с поведением оригинального Sentry и корректно обрабатывает client reports от всех поддерживаемых SDK (JavaScript, Python, etc.).