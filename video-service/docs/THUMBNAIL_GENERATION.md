# Video Thumbnail Generation

## Overview

The video service automatically generates thumbnails for uploaded videos by extracting the first frame of each video using FFmpeg.

## Features

- Automatic thumbnail generation on video upload
- First frame extraction from uploaded videos
- Thumbnails scaled to 320px width (height auto-calculated to maintain aspect ratio)
- Separate storage bucket for thumbnails
- Presigned URLs for thumbnail access
- JPEG format for thumbnails

## Architecture

### Components

1. **Thumbnail Generator** (`internal/thumbnail/generator.go`)
   - Uses FFmpeg to extract first frame
   - Scales thumbnail to 320px width
   - Returns JPEG image data

2. **Video Service Updates**
   - Reads uploaded video into memory
   - Generates thumbnail during upload
   - Uploads both video and thumbnail to separate buckets
   - Stores thumbnail filename in database
   - Generates presigned URLs for thumbnails on retrieval

3. **Storage**
   - Videos stored in configured bucket (MINIO_BUCKET)
   - Thumbnails stored in separate bucket (MINIO_THUMBNAILS_BUCKET)

## Environment Variables

Add the following to your `.env` file:

```env
# Optional - defaults to {MINIO_BUCKET}-thumbnails if not set
MINIO_THUMBNAILS_BUCKET=video-thumbnails
```

## Requirements

- FFmpeg must be installed in the runtime environment
- In Docker: Already included in Dockerfile.dev
- Local development: Install FFmpeg
  - macOS: `brew install ffmpeg`
  - Ubuntu/Debian: `apt-get install ffmpeg`
  - Alpine: `apk add ffmpeg`

## Database Schema

The `Video` model includes:

```go
type Video struct {
    // ... other fields
    VideoThumbnail    string `json:"videoThumbnail"`    // Presigned URL (generated on retrieval)
    ThumbnailFileName string `gorm:"-" json:"-"`         // Storage filename (not stored in DB)
}
```

## API Response

When retrieving videos, the response includes a `videoThumbnail` field with a presigned URL:

```json
{
  "id": "uuid",
  "title": "My Video",
  "videoThumbnail": "https://minio.example.com/thumbnails/uuid_thumbnail.jpg?signature=...",
  ...
}
```

## Error Handling

- If thumbnail generation fails, the video upload continues successfully
- Warnings are logged for thumbnail generation failures
- Videos without thumbnails will have empty `videoThumbnail` field
- This ensures video uploads are not blocked by thumbnail issues

## Performance Considerations

- Videos are read into memory for thumbnail generation
- For large videos, ensure adequate memory is available
- Thumbnail generation adds latency to upload process
- Consider async thumbnail generation for production use with large files
