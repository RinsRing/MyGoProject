# Distributed Recruitment Backend System (分布式招聘后端系统)

A high-performance, asynchronous recruitment backend system built with **Golang**. This project focuses on handling high-concurrency resume submissions and implementing modern backend patterns like rate limiting, caching, and message-based decoupling.

## 🌟 Highlights (核心亮点)

- **Distributed Rate Limiting (分布式限流)**: Implemented using **Redis (SetNX)** to protect core endpoints (Resume Upload) from malicious requests and spam, ensuring system stability.
- **Asynchronous Task Processing (异步任务处理)**: Utilizes **RabbitMQ** to decouple the resume submission flow. This allows the system to handle bursts of traffic (Peak Shaving) by moving heavy tasks (persistence/parsing) to a background consumer.
- **Caching Strategy (缓存优化)**: Implemented a **Cache-Aside Pattern** with Redis for hot job data. Applied **Randomized TTL** to prevent Cache Avalanche (缓存雪崩) and improved read throughput by 10x.
- **Secure Authentication (安全鉴权)**: Built a stateless authentication system using **JWT (JSON Web Tokens)** and custom **Gin Middleware** for identity propagation and access control.
- **Data Integrity (数据安全)**: Passwords are securely hashed using **Bcrypt** with salt, adhering to industry security standards.

## 🛠 Tech Stack (技术栈)

- **Language**: Go (Golang)
- **Web Framework**: Gin
- **Database**: MySQL (GORM as ORM)
- **Cache**: Redis (go-redis)
- **Message Queue**: RabbitMQ (amqp091-go)
- **Auth**: JWT (golang-jwt)
- **Containerization**: Docker & Docker-Compose

## 🏗 System Architecture (系统架构)

1. **User/HR** -> **Gin API Gateway** (Auth & Rate Limit)
2. **Gin** -> **Redis** (Cache Check & Token Validation)
3. **Gin** -> **RabbitMQ** (Producer: Dispatch Resume Tasks)
4. **RabbitMQ** -> **Worker** (Consumer: Persistent Data to MySQL)
5. **Worker** -> **MySQL** (Final Storage)

## 🚀 How to Run (如何运行)

1. **Clone the repo**:
   ```bash
   git clone https://github.com/RinsRing/MyGoProject.git
