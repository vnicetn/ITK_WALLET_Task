# Запустить все сервисы с пересборкой

make up-build

# Показать логи

make logs

# Остановить сервисы

make down

# Все тесты (unit + integration + concurrency)

make test-all

# Только unit-тесты (быстро, без БД)

make test-unit

# Интеграционные тесты

make test-integration

# Тесты на конкурентность

make test-concurrency
