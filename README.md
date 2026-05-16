# Bank REST API (Golang)

REST API для банковского сервиса, реализованный на Go.  
Поддерживает регистрацию, аутентификацию, управление счетами, переводы, выпуск виртуальных карт, кредитование, аналитику, интеграцию с ЦБ РФ и отправку email-уведомлений.

## Технологии
- **Язык:** Go 1.23+
- **Маршрутизация:** gorilla/mux
- **База данных:** PostgreSQL 18
- **Драйвер БД:** lib/pq
- **Аутентификация:** JWT (golang-jwt/jwt/v5)
- **Логирование:** logrus
- **Шифрование:** bcrypt (пароли, CVV), AES-256 (данные карт), HMAC-SHA256 (целостность)
- **Интеграции:** ЦБ РФ (SOAP, etree), SMTP (gomail)
- **Планировщик:** обработка просроченных кредитных платежей каждые 12 часов

## Методы

Метод	URL	Описание

POST	/register	- Регистрация

POST	/login	- Аутентификация

POST	/accounts -	Создать новый счёт

GET	/accounts -	Список счетов пользователя

POST	/transfer -	Перевод между счетами

POST	/cards -	Выпуск виртуальной карты

GET	/cards	- Список карт (с расшифровкой)

POST	/credits - Оформление кредита

GET	/credits/{creditId}/schedule -	График платежей по кредиту

GET	/analytics -	Аналитика (доходы/расходы/прогноз)

GET	/accounts/{accountId}/predict?days=N -	Прогноз баланса на N дней (≤365)


## Примеры использования (curl)

Предварительно необходимо запустить процесс postgresql, создать и подключить БД bank
Выполнить скрипт .sql: psql -d bank -f migrations/001_init.sql
Запустить main.go в фоне: go run main.go &

Регистрация:
```
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","email":"test@mail.com","password":"123"}'
```
Ответ:
```
{"user_id":1}
```

Логин и сохранение токена:
```
TOKEN=$(curl -s -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"123"}' | jq -r '.token')
```

Создание 2 счетов:
```
curl -X POST http://localhost:8080/accounts -H "Authorization: Bearer $TOKEN"
curl -X POST http://localhost:8080/accounts -H "Authorization: Bearer $TOKEN"
```
Ответ:
```
{"account_id":1}
{"account_id":2}
```

Пополнение баланса через psql "UPDATE accounts SET balance=1000 WHERE id=1;"

Перевод 100 рублей с 1-го аккаунта на 2-й:
```
curl -X POST http://localhost:8080/transfer \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"from_account_id":1,"to_account_id":2,"amount":100}'
```
Ответ:
```
{"status":"ok"}
```

Выпуск карты:
```
curl -X POST http://localhost:8080/cards \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"account_id":1}'
```
Ответ:
```
{"id":1,"account_id":1,"hmac":"36b018fca5af1406389be356fec663804206414f00fb519e13667b3fee575c90","owner_id":1,"card_number":"4000005656666586","expiry":"05/29"}
```

Просмотр карт:
```
curl -X GET http://localhost:8080/cards -H "Authorization: Bearer $TOKEN"
```
Ответ:
```
[{"id":1,"account_id":1,"hmac":"36b018fca5af1406389be356fec663804206414f00fb519e13667b3fee575c90","owner_id":1,"card_number":"4000005656666586","expiry":"05/29"}]
```

Оформление кредита (5 тыс. на 12 мес.):
```
curl -X POST http://localhost:8080/credits \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"account_id":1,"amount":5000,"months":12}'
```
Ответ:
```
INFO[0070] Credit issued: user 1, amount 5000.00        
{"id":1,"user_id":1,"account_id":1,"amount":5000,"rate":19.5,"term_months":12,"monthly_payment":461.98,"status":"active"}
```

График платежей:
```
curl -X GET "http://localhost:8080/credits/1/schedule" -H "Authorization: Bearer $TOKEN"
```
ответ:
```
[{"id":1,"credit_id":1,"due_date":"2026-06-16T00:00:00Z","amount":461.98,"paid":false},{"id":2,"credit_id":1,"due_date":"2026-07-16T00:00:00Z","amount":461.98,"paid":false},{"id":3,"credit_id":1,"due_date":"2026-08-16T00:00:00Z","amount":461.98,"paid":false},{"id":4,"credit_id":1,"due_date":"2026-09-16T00:00:00Z","amount":461.98,"paid":false},{"id":5,"credit_id":1,"due_date":"2026-10-16T00:00:00Z","amount":461.98,"paid":false},{"id":6,"credit_id":1,"due_date":"2026-11-16T00:00:00Z","amount":461.98,"paid":false},{"id":7,"credit_id":1,"due_date":"2026-12-16T00:00:00Z","amount":461.98,"paid":false},{"id":8,"credit_id":1,"due_date":"2027-01-16T00:00:00Z","amount":461.98,"paid":false},{"id":9,"credit_id":1,"due_date":"2027-02-16T00:00:00Z","amount":461.98,"paid":false},{"id":10,"credit_id":1,"due_date":"2027-03-16T00:00:00Z","amount":461.98,"paid":false},{"id":11,"credit_id":1,"due_date":"2027-04-16T00:00:00Z","amount":461.98,"paid":false},{"id":12,"credit_id":1,"due_date":"2027-05-16T00:00:00Z","amount":461.98,"paid":false}]
```


