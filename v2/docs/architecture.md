# Architecture

## Overview

The system consists of two main components:
- **Frontend**: Web application located in `web/`
- **Backend**: API and webhook server located in `api/`

## Backend Architecture

The backend serves dual purposes:

### 1. API Server for Frontend
- Provides REST/HTTP endpoints for the web application
- Handles all frontend requests and business logic
- Translates API calls into Kubernetes Custom Resource (CR) operations

### 2. Webhook Server for Kubernetes
- Receives validation webhook requests from Kubernetes API server
- Receives mutation webhook requests from Kubernetes API server
- Validates and mutates CRs before they are persisted

## Data Persistence

All application data is stored as Kubernetes Custom Resources (CRs). The flow works as follows:

1. Frontend makes API calls to the backend
2. Backend creates/updates/deletes corresponding CRs
3. Kubernetes intercepts CR operations and calls backend webhooks
4. Backend validates and/or mutates the CRs
5. Kubernetes persists the CRs if validation passes

## Request Flow

```
Frontend (web/)
    ↓ HTTP/REST
Backend API Server (api/)
    ↓ CR operations
Kubernetes API Server
    ↓ Webhook calls
Backend Webhook Server (api/)
    ↓ Validation/Mutation
Kubernetes etcd (CR persistence)
```
