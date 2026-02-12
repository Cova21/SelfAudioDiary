# Voice Diary with AI 🎙️🤖

Полнофункциональное микросервисное приложение для записи голосовых заметок с автоматической транскрипцией, AI-анализом эмоций и полнотекстовым поиском.

## 🏗️ Архитектура

Приложение состоит из 9 микросервисов:

### Go Микросервисы
- **auth-service** - Аутентификация и авторизация (JWT + Redis)
- **diary-service** - Основной сервис управления записями
- **storage-service** - Работа с MinIO для хранения аудиофайлов
- **search-service** - Полнотекстовый поиск через Elasticsearch
- **notification-service** - WebSocket уведомления в реальном времени
- **api-gateway** - Единая точка входа для всех запросов

### Python AI Сервисы
- **transcription-service** - Транскрипция аудио через OpenAI Whisper
- **nlp-service** - Анализ эмоций и ключевых моментов через Hugging Face

### Frontend
- **Next.js 14** - Современный фронтенд с темной темой и красными акцентами

## 🚀 Технологический стек

### Backend
- Go 1.21+ с Clean Architecture
- gRPC + Protocol Buffers
- PostgreSQL (отдельная БД на каждый сервис)
- Redis для сессий и кеша
- RabbitMQ для асинхронных задач
- MinIO для хранения аудиофайлов
- Elasticsearch для полнотекстового поиска

### AI Services
- FastAPI
- OpenAI Whisper для транскрипции
- Hugging Face Transformers для NLP

### Frontend
- Next.js 14 с App Router
- TypeScript + CSS Modules
- Web Audio API
- NextAuth.js

### Infrastructure
- Docker + Docker Compose
- Nginx Reverse Proxy
- Prometheus + Grafana (опционально)

## 📋 Требования

- Docker 20.10+
- Docker Compose 2.0+
- Make (опционально, для удобства)

## 🛠️ Быстрый старт

### 1. Клонирование и настройка

```bash
# Создайте .env файл из примера
cp .env.example .env

# Отредактируйте .env и установите надежные пароли
nano .env
```

### 2. Генерация protobuf кода (если нужно)

```bash
make proto
```

### 3. Запуск всех сервисов

```bash
# Сборка и запуск
docker-compose up --build -d

# Или используя Makefile
make build
make up
```

### 4. Проверка здоровья сервисов

```bash
make health
```

### 5. Доступ к приложению

- **Frontend**: http://localhost
- **API Gateway**: http://localhost/api
- **WebSocket**: ws://localhost/ws
- **RabbitMQ Management**: http://localhost:15672 (guest/guest)
- **MinIO Console**: http://localhost:9001 (minioadmin/minioadmin)
- **Elasticsearch**: http://localhost:9200

## 📖 Основные команды

```bash
make help              # Показать все доступные команды
make setup             # Первоначальная настройка проекта
make up                # Запустить все сервисы
make down              # Остановить все сервисы
make logs              # Показать логи всех сервисов
make logs-diary        # Показать логи конкретного сервиса
make restart           # Перезапустить все сервисы
make clean             # Очистить все данные
make ps                # Показать статус сервисов
```

## 🎯 Основные функции

1. ✅ Регистрация и авторизация пользователей
2. ✅ Запись аудио через браузер (до 10 минут)
3. ✅ Автоматическая транскрипция через Whisper
4. ✅ AI-анализ эмоций и ключевых моментов
5. ✅ Воспроизведение аудио с синхронизацией текста
6. ✅ Полнотекстовый поиск по записям
7. ✅ Фильтрация по эмоциям и тегам
8. ✅ Real-time уведомления через WebSocket

## 🔧 Разработка

### Структура проекта

```
voice-diary/
├── backend/                 # Go микросервисы
│   ├── internal/
│   │   ├── protobuf/       # .proto файлы
│   │   ├── gen/            # Сгенерированный код
│   │   └── pkg/            # Общие утилиты
│   ├── auth-service/
│   ├── diary-service/
│   ├── storage-service/
│   ├── search-service/
│   ├── notification-service/
│   └── api-gateway/
├── ai-services/            # Python AI сервисы
│   ├── transcription-service/
│   └── nlp-service/
├── frontend/               # Next.js приложение
├── deployments/            # Конфигурации развертывания
│   └── nginx/
└── docker-compose.yml
```

### Запуск тестов

```bash
make test-backend          # Тесты Go сервисов
make test-ai               # Тесты Python сервисов
```

## 🔐 Безопасность

- JWT токены для аутентификации
- Хеширование паролей через bcrypt
- HTTPS через Nginx (в продакшене)
- Валидация на всех уровнях
- Rate limiting в API Gateway

## 📊 Мониторинг

Все сервисы предоставляют:
- Health check endpoints
- Structured logging в JSON
- Метрики (готовность к Prometheus)

## 🐛 Устранение неполадок

### Проблемы с запуском

```bash
# Проверьте логи конкретного сервиса
make logs-diary

# Проверьте статус всех контейнеров
docker-compose ps

# Перезапустите проблемный сервис
docker-compose restart diary-service
```

### Очистка данных

```bash
# Полная очистка (ВНИМАНИЕ: удалит все данные!)
make clean
```

## 📝 Лицензия

MIT License

## 🤝 Вклад в проект

Приветствуются pull requests!

## 📧 Контакты

Для вопросов и предложений создавайте issue в репозитории.
