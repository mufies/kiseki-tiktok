# Build stage
FROM mcr.microsoft.com/dotnet/sdk:8.0 AS build
WORKDIR /

# Copy proto files first (when context is parent dir)
COPY ["proto", "proto/"]

# Copy notification service files
WORKDIR /src
COPY ["notification-service/NotificationService/NotificationService.csproj", "./"]
RUN dotnet restore "NotificationService.csproj"

# Copy everything else and build
COPY ["notification-service/NotificationService", "./"]
RUN dotnet build "NotificationService.csproj" -c Release -o /app/build

# Publish stage
FROM build AS publish
RUN dotnet publish "NotificationService.csproj" -c Release -o /app/publish /p:UseAppHost=false

# Runtime stage
FROM mcr.microsoft.com/dotnet/aspnet:8.0 AS final
WORKDIR /app
COPY --from=publish /app/publish .

ENV ASPNETCORE_URLS=http://+:8085
EXPOSE 8085 9093

ENTRYPOINT ["dotnet", "NotificationService.dll"]
