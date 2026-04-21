# Stage 1: Build the application
FROM node:22 AS build

# Установка рабочей директории
WORKDIR /app

# Копирование package.json и package-lock.json
COPY package*.json ./

# Установка зависимостей
RUN npm install

# Глобальная установка serve (для обслуживания приложения)
RUN npm install -g serve

RUN node -v && npm -v

# Копирование остальных файлов проекта
COPY . .

# Сборка приложения
#RUN npm run build

# Открытие порта
EXPOSE 3000

# Запуск приложения с помощью serve
CMD ["serve", "-s", "dist"]