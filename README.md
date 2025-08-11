# 📊 Go Ad Tracking & Analytics System

A high-performance **Ad Click Tracking & Analytics** service built with **Go**, **Kafka**, and **Prometheus**, designed for **real-time click tracking**, **scalable event processing**, and **detailed analytics**.

## 🚀 Features

### Ad Management

* **List ads with pagination** - Browse through ads efficiently with configurable page sizes
* **Store ad metadata** - Comprehensive ad information for enhanced tracking capabilities

### Click Tracking

* **Non-blocking click recording** using** ****Kafka** for high throughput
* **High traffic handling** - Designed to handle burst loads without performance degradation
* **Data reliability** - Ensures** ****no data loss** with retry-safe processing mechanisms
* **Event deduplication** - Prevents duplicate click recordings

### Real-Time Analytics

* **Performance metrics** - Real-time ad performance data including click counts and CTR
* **Detailed insights** - Comprehensive analytics for data-driven decision making

### Observability & Monitoring

* **Prometheus metrics** endpoint (`/metrics`) for system monitoring
* **Structured logging** for effective debugging and troubleshooting
* **Health checks** for service monitoring

### Scalable Architecture

* **Asynchronous processing** with Kafka for better resource utilization
* **Microservices-ready** deployment for easy scaling
* **Containerized deployment** for consistent environments

## 🛠 Tech Stack

| Component                      | Technology   |
| ------------------------------ | ------------ |
| **Backend**              | Go (Golang)  |
| **Message Queue**        | Apache Kafka |
| **Metrics & Monitoring** | Prometheus   |
| **Containerization**     | Docker       |
| **Cloud Infrastructure** | AWS EC2      |

## 📦 Deployment

The project is** ****deployed on an AWS EC2 instance** running inside** ** **Docker containers** .

### 🌐 Server Information

* **Server IP:** `13.201.125.143`
* **API Base URL:** `http://13.201.125.143:8080/api/v1/ads`
* **Metrics Endpoint:** `http://13.201.125.143:8080/metrics`

## 🔗 API Endpoints

| Method   | Endpoint       | Description                         | Response                    |
| -------- | -------------- | ----------------------------------- | --------------------------- |
| `GET`  | `/`          | Get paginated list of ads           | List of ads with metadata   |
| `POST` | `/click`     | Record a click event (non-blocking) | Success confirmation        |
| `GET`  | `/analytics` | Fetch real-time ad analytics        | Analytics data with metrics |
| `GET`  | `/metrics`   | Prometheus metrics endpoint         | Prometheus format metrics   |

## 📌 API Usage Examples

### 1️⃣ Get Paginated List of Ads

```bash
curl -X GET "http://13.201.125.143:8080/api/v1/ads?page=1&limit=10" \
  -H "Content-Type: application/json"
```

**Query Parameters:**

* `page` (optional): Page number (default: 1)
* `limit` (optional): Number of ads per page (default: 10, max: 100)

### 2️⃣ Record a Click Event

```bash
curl -X POST "http://13.201.125.143:8080/api/v1/ads/click" \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "550e8400-e29b-41d4-a716-446655440000",
    "ad_id": 11,
    "user_ip": "192.168.1.100",
    "agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
    "play_time_secs": 30,
    "watched_percent": 80,
    "timestamp": 1691745600
  }'
```

**Request Body Fields:**

* `event_id` (string): Unique identifier for the click event (UUID format recommended)
* `ad_id` (integer): ID of the clicked ad
* `user_ip` (string): Client's IP address
* `agent` (string): User agent string
* `play_time_secs` (integer): Duration the ad was played in seconds
* `watched_percent` (integer): Percentage of ad watched (0-100)
* `timestamp` (integer): Unix timestamp of the event

### 3️⃣ Get Real-Time Analytics

```bash
curl -X GET "http://13.201.125.143:8080/api/v1/ads/analytics?ad_id=32&since=2025-07-01T00:00:00Z&until=2025-08-15T00:00:00Z"
```

**Optional Query Parameters:**

* `ad_id` (integer): Filter analytics for specific ad
* `since` (string): Start date in ISO 8601 format (e.g., "2025-07-01T00:00:00Z")
* `until` (string): End date in ISO 8601 format (e.g., "2025-08-15T00:00:00Z")

### 4️⃣ Access Prometheus Metrics

```bash
curl -X GET "http://13.201.125.143:8080/metrics"
```

## 📊 Sample Response Formats

### Ads List Response

```json
{
  "ads": [
    {
      "id": 1,
      "title": "Sample Ad",
      "description": "This is a sample advertisement",
      "url": "https://example.com",
      "created_at": "2024-01-15T10:30:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 10,
    "total": 25,
    "total_pages": 3
  }
}
```

### Analytics Response

```json
{
  "data": [
    {
      "ad_id": 32,
      "click_count": 2,
      "unique_clicks": 2,
      "avg_playback_time": 0,
      "avg_watch_percent": 0,
      "last_updated": "2025-07-22T16:16:12.879961Z"
    }
  ],
  "total": 1,
  "generated_at": "2025-08-11T07:49:18.342389686Z",
  "is_real_time": false
}
```

### Prometheus Metrics

The system exposes various metrics including:

* HTTP request duration and count
* Kafka message processing metrics
* Active connections
* Error rates
* Custom business metrics

## 🚀 Performance Features

* **High Throughput** : Handles thousands of concurrent click events
* **Low Latency** : Non-blocking API responses for click tracking
* **Fault Tolerant** : Automatic retry mechanisms and error handling
* **Scalable** : Horizontal scaling support with load balancing

## 🐳 Docker Deployment

The application runs in Docker containers with the following setup:

* **Application Container** : Go service with optimized runtime
* **Kafka Container** : Message queue for event processing
* **Prometheus Container** : Metrics collection and monitoring

## 📈 Scaling Recommendations

For production environments:

1. **Load Balancer** : Use AWS ALB or similar for traffic distributions
2. **Kafka Cluster** : Multi-node Kafka setup for high availability
3. **Monitoring** : Grafana dashboards for comprehensive monitoring
4. **Caching** : Redis for frequently accessed analytics data
