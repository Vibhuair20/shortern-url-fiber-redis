Great! If you already have a Dockerfile for your database and a docker-compose.yml for the repo, I’ll help you update the README.md with instructions for using Docker to set up the entire project. Here’s how we can enhance your README.md with Docker-related steps.

⸻



# 🔗 URL Shortener with QR Code Generator

A blazing fast URL shortener API built with **Go (Fiber)**, **Redis**, and **QR Code** support. Features include custom short URLs, rate limiting, and automatic QR code generation.

## 🚀 Features

- Shorten long URLs quickly and efficiently
- Optional custom short URLs
- Auto-generated QR Code for each short URL (cached in Redis)
- Rate limiting per IP (customizable)
- URL expiry support
- Redis-based caching and storage

## 📸 Example Response

``` json
{
  "url": "https://google.com",
  "short_url": "http://localhost:3000/abcd12",
  "qr_code_url": "data:image/png;base64,iVBORw0K...",
  "expiry": 24,
  "rate_limit": 9,
  "rate_limit_reset": 29
}

📦 Tech Stack
	•	Go
	•	Fiber – Express-inspired web framework for Go
	•	Redis – Fast in-memory data store
	•	go-qrcode – QR Code generator for Go
	•	govalidator – URL validation
```

🛠️ Setup Instructions

1. Clone the Repository

git clone https://github.com/Vibhuair20/shortern-url-fiber-redis.git
cd shortern-url-fiber-redis

2. Install Go Modules

go mod tidy

3. Setup Docker

The repository includes Docker configurations for both the application and the Redis database. To run everything inside Docker containers, follow the steps below.
	•	Build and Run Docker Containers:
You can easily set up the entire environment using Docker Compose. This will set up the Go app and Redis container together.

docker-compose up --build

This command will:
	•	Build the Docker images.
	•	Set up the Redis container.
	•	Start the Go application.
If everything is set up correctly, the application will be accessible on http://localhost:3000.

4. Environment Variables

The .env file should contain the following environment variables:
	•	API_QUOTA=10
Max requests per IP within 30 minutes.
	•	DOMAIN=http://localhost:3000
The base domain for short URLs.

Create the .env file in the root of your project and add the variables.

5. Run the API with Docker

You can also manually run the Go API and Redis using Docker.
	•	Start Redis Container (if you’re not using Docker Compose):

docker run --name redis -p 6379:6379 -d redis


	•	Start Go Application:
If you want to run the Go API inside a Docker container, you can use this Dockerfile configuration:

docker build -t shortener-app .
docker run -p 3000:3000 --env-file .env shortener-app



Alternatively, you can use Docker Compose to manage both containers simultaneously, as mentioned in Step 3.

6. Test the API

You can test the URL shortening API using Postman or cURL. Here’s an example cURL request:

curl -X POST http://localhost:3000/api/v1 -H "Content-Type: application/json" \
-d '{
      "url": "https://example.com",
      "short_url": "go123",
      "expiry": 24
    }'

7. Rate Limiting

Each IP is allowed 10 requests per 30 minutes. After exceeding the limit, the API will return a cooldown time.

📎 License

MIT License. Feel free to fork, improve, and contribute!

⸻

🙌 Contributing

Contributions are welcome! Feel free to open an issue or submit a PR.
