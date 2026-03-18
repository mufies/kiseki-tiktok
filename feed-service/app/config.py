from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    model_config = SettingsConfigDict(env_file=".env", env_file_encoding="utf-8")

    # PostgreSQL (still used by ProfileRepository / history)
    database_url: str = "postgresql://postgres:postgres@localhost:5432/eventdb"

    # Redis (profile + watch history)
    redis_url: str = "redis://localhost:6379"

    # gRPC addresses of upstream services
    video_service_grpc: str = "localhost:9091"        # Video Service gRPC
    event_service_grpc: str = "localhost:5002"        # Event Service gRPC
    interaction_service_grpc: str = "localhost:9092"  # Interaction Service gRPC

    port: int = 8001


settings = Settings()
