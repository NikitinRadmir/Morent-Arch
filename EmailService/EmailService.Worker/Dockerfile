# See https://aka.ms/customizecontainer
FROM mcr.microsoft.com/dotnet/aspnet:8.0 AS base
USER app
WORKDIR /app

FROM mcr.microsoft.com/dotnet/sdk:8.0 AS build
ARG BUILD_CONFIGURATION=Release
WORKDIR /src

# 🔹 Копируем sln
COPY ["EmailService.Api/EmailService.Api.sln", "EmailService.Api.sln"]

# Копируем csproj файлы
COPY ["EmailService.Worker/EmailService.Worker.csproj", "EmailService.Worker/EmailService.Worker.csproj"]
COPY ["EmailService.Core/EmailService.Core.csproj", "EmailService.Core/EmailService.Core.csproj"]
COPY ["EmailService.Infrastructure/EmailService.Infrastructure.csproj", "EmailService.Infrastructure/EmailService.Infrastructure.csproj"]

# 🔹 Restore с правильным путём
RUN dotnet restore "EmailService.Worker/EmailService.Worker.csproj"

# Копируем исходный код
COPY ["EmailService.Worker/", "EmailService.Worker/"]
COPY ["EmailService.Core/", "EmailService.Core/"]
COPY ["EmailService.Infrastructure/", "EmailService.Infrastructure/"]

# 🔹 Меняем WORKDIR перед build/publish
WORKDIR "/src/EmailService.Worker"
RUN dotnet build "EmailService.Worker.csproj" -c $BUILD_CONFIGURATION -o /app/build

FROM build AS publish
RUN dotnet publish "EmailService.Worker.csproj" -c $BUILD_CONFIGURATION -o /app/publish /p:UseAppHost=false

FROM base AS final
WORKDIR /app
COPY --from=publish /app/publish .
ENTRYPOINT ["dotnet", "EmailService.Worker.dll"]