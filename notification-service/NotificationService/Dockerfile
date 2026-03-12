# Build stage
FROM mcr.microsoft.com/dotnet/sdk:8.0 AS build
WORKDIR /src

# Copy csproj and restore dependencies
COPY ["NotificationService.csproj", "./"]
RUN dotnet restore "NotificationService.csproj"

# Copy everything else and build
COPY . .
WORKDIR "/src"
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
