# Authentication REST API

A simple authentication REST API built with Go.

## Tech Stack

* Go
* Gin
* GORM
* PostgreSQL
* JWT
* bcrypt
* Redis
* Apache Kafka
* Elasticsearch
* Logstash
* Kibana
* Docker
* Docker Compose

## Features

* User registration
* User login
* JWT authentication (access token & refresh token)
* Get current user (profile)
* Forgot password
* Reset password
* Password hashing with bcrypt
* Input validation
* Error handling
* Product CRUD
* Rate limiting with Redis
* Event streaming with Apache Kafka
* Centralized logging with ELK Stack
* Log collection, processing, storage, and visualization
* Real-time monitoring and log analysis
* Containerized application and infrastructure with Docker
* Multi-container orchestration with Docker Compose

## Architecture

```text
                         Docker Environment
┌─────────────────────────────────────────────────────────────┐
│                                                             │
│   Client                                                    │
│     │                                                       │
│     ▼                                                       │
│   Go REST API (Gin)                                         │
│     │                                                       │
│     ├──────────────► PostgreSQL                             │
│     │                                                       │
│     ├──────────────► Redis                                  │
│     │                                                       │
│     └──────────────► Kafka                                  │
│                            │                                │
│                            ▼                                │
│                        Logstash                             │
│                            │                                │
│                            ▼                                │
│                      Elasticsearch                          │
│                            │                                │
│                            ▼                                │
│                         Kibana                               │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## Dockerization

The entire application infrastructure is containerized using Docker and Docker Compose.

Docker Compose manages the following services:

* Go API
* PostgreSQL
* Redis
* Apache Kafka
* Elasticsearch
* Logstash
* Kibana

The Go API is built using a **multi-stage Docker build**, where the application is compiled in a Go builder image and then copied into a lightweight Alpine runtime image.

Logstash is configured with separate pipelines for:

* Collecting logs from application log files
* Consuming application events from Kafka

Both pipelines process the logs and send them to Elasticsearch, where they can be searched, analyzed, and visualized through Kibana.
