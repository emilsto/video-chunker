# Video Chunker API

POC for chunking video files into smaller parts and uploading them to S3 (todo).

## Docker Setup

For ez deployment, you can use Docker to run the application.

Run the entire application stack using Docker Compose:

```bash
docker-compose up --build
```

This will start:
- Backend API on port 5000
- Frontend server on port 8080

Access the application at http://127.0.0.1:8080 

## Development

Set up config.json with the following structure:

```json
{
  "server": {
    "port": 5000
  }
}
```

Recommended to use air to run the go server. Install it with:

```bash
go get -u github.com/cosmtrek/air
```

Run the server with:

```bash
air
``` 

Make sure you have ffmpeg installed on your machine, sice this project uses it to chunk the video files.


##  Running the frontend

Simply use npx live-server to run the frontend. Make sure that backend is running on port 5000.

```bash
npx live-server
```
