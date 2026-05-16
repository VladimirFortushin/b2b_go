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

'''
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"username":"ivan","email":"test@mail.com","password":"123"}'
  '''
