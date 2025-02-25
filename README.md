# Log Monitoring API

![Log Monitoring API](assets/log_monitoring.png)

This is a **Log Monitoring API** built with **Go**, **Gin**, and **PostgreSQL**. It provides functionalities for logging, authentication, and monitoring, making it a great example of a modern backend application.

---

## Features

- **JWT Authentication**: Secure login and token-based authentication.
- **CRUD Operations**: Create, Read, Update, and Delete log entries.
- **Prometheus Metrics**: Monitor HTTP requests and response times.
- **Swagger Documentation**: Interactive API documentation.
- **Graceful Shutdown**: Ensures the server shuts down properly.
- **Environment Variables**: Easy configuration using a `.env` file.

---

## Technologies Used

- **Go**: The primary programming language.
- **Gin**: A high-performance HTTP web framework.
- **PostgreSQL**: A powerful relational database.
- **JWT**: JSON Web Tokens for secure authentication.
- **Prometheus**: For monitoring and metrics.
- **Swagger**: For API documentation.
- **Docker**: For containerization (optional).

---

## Getting Started

### Prerequisites

- Go 1.20 or higher
- PostgreSQL
- Git

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/mustafa-bagci/LogMonitoringAPI.git
   cd LogMonitoringAPI

2. Set up environment variables:
   ```bash
   DATABASE_URL=postgres://user:password@localhost/dbname?sslmode=disable
   JWT_SECRET=your-secret-key
   PORT=8080
   ```

3. Install dependencies:
   ```bash
   go mod download
   ```

4. Run the application:
   ```bash
   go run main.go
   ```

5. Access the API:
   - Swagger UI: http://localhost:8080/swagger/index.html
   - Prometheus Metrics: http://localhost:8080/metrics
   - Health Check: http://localhost:8080/health

### API Endpoints

#### Authentication
- POST /login: Authenticate and get a JWT token.

#### Logs
- GET /logs: Retrieve all log entries.
- POST /logs: Add a new log entry.
- PUT /logs/{id}: Update a log entry by ID.
- PATCH /logs/{id}: Partially update a log entry by ID.
- DELETE /logs/{id}: Delete a log entry by ID.

### Monitoring with Prometheus and Grafana

#### Step 1: Set Up Prometheus and Grafana
Create a `docker-compose.yml` file in the root directory:
   ```yaml
   version: '3'
   services:
     prometheus:
       image: prom/prometheus
       ports:
         - "9090:9090"
       volumes:
         - ./prometheus.yml:/etc/prometheus/prometheus.yml
     grafana:
       image: grafana/grafana
       ports:
         - "3000:3000"
   ```

Create a `prometheus.yml` file to configure Prometheus:
   ```yaml
   global:
     scrape_interval: 15s

   scrape_configs:
     - job_name: 'log_monitoring_api'
       static_configs:
         - targets: ['host.docker.internal:8080']
   ```

Start Prometheus and Grafana using Docker Compose:
   ```bash
   docker-compose up -d
   ```

Access the services:
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000 (default credentials: admin/admin)

#### Step 2: Configure Grafana
Log in to Grafana at http://localhost:3000.

Add Prometheus as a data source:
1. Go to Configuration > Data Sources.
2. Click Add data source.
3. Select Prometheus.
4. Set the URL to http://prometheus:9090.
5. Click Save & Test.

Create a dashboard to visualize metrics:
1. Go to Create > Dashboard.
2. Add a new panel and query your metrics (e.g., http_requests_total).
3. Save the dashboard.

### Running Tests
To run the unit tests:
   ```bash
   go test -v ./...
   ```

### Deployment

#### Docker
Build the Docker image:
   ```bash
   docker build -t log-monitoring-api .
   ```

Run the Docker container:
   ```bash
   docker run -p 8080:8080 --env-file .env log-monitoring-api
   ```

#### Kubernetes (Optional)
If you're familiar with Kubernetes, you can deploy the application using a Kubernetes cluster.

### Contributing
Contributions are welcome! Please follow these steps:
1. Fork the repository.
2. Create a new branch (`git checkout -b feature/log-filtering`).
3. Commit your changes (`git commit -m 'Add some feature'`).
4. Push to the branch (`git push origin feature/log-filtering`).
5. Open a pull request.

### License
This project is licensed under the MIT License. See the LICENSE file for details.