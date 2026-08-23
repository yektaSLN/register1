A simple authentication REST API built with Go.

Used in this project:

- Go
- Gin
- GORM
- PostgreSQL
- JWT
- bcrypt
- Redis
- Apache Kafka
- ELK Stack (Elasticsearch, Logstash, Kibana)

Features included:

- User registration
- User login
- JWT authentication (access token & refresh token)
- Get current user (profile)
- Forgot password
- Reset password
- Password hashing
- Input validation
- Error handling
- Product CRUD
- Rate limiting
- Event streaming with Kafka
- Centralized logging with ELK Stack
- Log collection, processing, storage, and visualization
- Real-time monitoring and log analysis

Architecture:

Client
  │
  ▼
Go REST API (Gin)
  │
  ├── PostgreSQL
  ├── Redis
  └── Kafka
        │
        ▼
      Logstash
        │
        ▼
   Elasticsearch
        │
        ▼
      Kibana