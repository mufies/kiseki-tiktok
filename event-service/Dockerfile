# Build stage
FROM mcr.microsoft.com/dotnet/sdk:8.0 AS build
WORKDIR /

# Copy proto files first (when context is parent dir)
COPY ["proto", "proto/"]

# Copy event service files
WORKDIR /src
COPY ["event-service/EventService.csproj", "./"]
RUN dotnet restore "EventService.csproj"

# Copy everything else and build
COPY ["event-service", "./"]
RUN dotnet build "EventService.csproj" -c Release -o /app/build

# Publish stage
FROM build AS publish
RUN dotnet publish "EventService.csproj" -c Release -o /app/publish /p:UseAppHost=false

# Runtime stage
FROM mcr.microsoft.com/dotnet/aspnet:8.0 AS final
WORKDIR /app
COPY --from=publish /app/publish .

ENV ASPNETCORE_URLS=http://+:8083
EXPOSE 5001 9092

ENTRYPOINT ["dotnet", "EventService.dll"]
