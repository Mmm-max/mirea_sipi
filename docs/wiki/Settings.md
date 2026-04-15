# Настройки профиля

## Профиль пользователя

**Просмотр профиля:** GET `/users/me`

Ответ содержит: email, имя, часовой пояс.

**Обновление профиля:** PATCH `/users/me`:
```json
{
  "display_name": "Иван Иванов",
  "timezone": "Europe/Moscow"
}
```

> Часовой пояс используется для корректного отображения времени и учитывается при поиске слотов. Рекомендуется установить его сразу после регистрации.

## Рабочие часы

Рабочие часы определяют, в какое время вы доступны для встреч. Если рабочие часы заданы, слоты за их пределами будут иметь жёсткий конфликт.

**Установить рабочие часы:** PUT `/users/me/working-hours`:
```json
[
  { "weekday": 1, "is_working_day": true, "start_time": "09:00", "end_time": "18:00" },
  { "weekday": 2, "is_working_day": true, "start_time": "09:00", "end_time": "18:00" },
  { "weekday": 3, "is_working_day": true, "start_time": "09:00", "end_time": "18:00" },
  { "weekday": 4, "is_working_day": true, "start_time": "09:00", "end_time": "18:00" },
  { "weekday": 5, "is_working_day": true, "start_time": "09:00", "end_time": "17:00" },
  { "weekday": 6, "is_working_day": false },
  { "weekday": 0, "is_working_day": false }
]
```

> `weekday`: 0 — воскресенье, 1 — понедельник, ..., 6 — суббота (стандарт Go/Unix).

**Просмотр рабочих часов:** GET `/users/me/working-hours`

## Периоды недоступности

Периоды недоступности блокируют вас для встреч на всё указанное время (жёсткий конфликт).

**Создать период:** POST `/users/me/unavailability`:
```json
{
  "title": "Отпуск",
  "type": "vacation",
  "start_at": "2025-07-01T00:00:00Z",
  "end_at": "2025-07-14T23:59:59Z"
}
```

**Типы недоступности:**
- `vacation` — отпуск
- `sick_leave` — больничный
- `business_trip` — командировка
- `custom` — произвольная причина

**Список периодов:** GET `/users/me/unavailability`

**Удалить период:** DELETE `/users/me/unavailability/:id`
