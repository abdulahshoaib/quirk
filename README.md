# Quirk

Quirk is a data processing pipeline designed for production environments, automating corpus extraction, embedding generation using Cloudflare's embedding API, and storing the embeddings into a vector database (ChromaDB). It supports persistent token-based authentication using PostgreSQL.

The system accepts user data through two primary input methods:
- Direct submission via API endpoint
- Automated data retrieval through external API

Upon receiving data, the system processes and embeds it into CSV format, structuring the information into rows and columns for easy analysis and export.

## Features

- REST API for embedding processing and querying
- Export embeddings to JSON, CSV, or ChromaDB
- Supports Cloudflare Workers AI with `bge-large-en-v1.5`
- Automatically initializes and migrates PostgreSQL database if not already set up
- JWT-based authentication with persistent token storage
- Direct ChromaDB integration for vector storage
- **NEW!** functioning frontend for GUI interface [Frontend](https://github.com/abdulahshoaib/quirk-frontend)

## Core Functionality

Core functionality provides endpoints for data upload, batch processing, similarity search queries, and CSV export of embeddings with associated metadata. The system includes an exposed HTTP endpoint for external integration and interaction.

## Data Formats

Supported input formats:
- `.pdf`
- `.csv`
- `.txt`
- `.json`
- `.md`
- `.yml`
- `.xml`

## RESTful API

The system provides:
- RESTful API endpoints for data submission and processing
- External API integration for automated data fetching
- Exposed HTTP endpoints for external system interaction
- Export capabilities to multiple formats (JSON, CSV, ChromaDB)

## Tech Stack

### Go
Primary backend language

**Merits:**
- Built-in HTTP Server
- Support for CSV processing, HTTP handling, and file operations

### PostgreSQL
**Usage:**
- Stores auth tokens for user sessions
- Handles automatic DB creation and schema migration
- Tables are created on startup if they don't exist

### Embeddings
- **Provider:** [Cloudflare Workers AI](https://developers.cloudflare.com/workers-ai/)
- **Model:** @cf/baai/bge-large-en-v1.5
- **Embedding Dimensions:** 1024

### Docker
[**Docker Hub**](https://hub.docker.com/r/abdulahshoaib/quirk) - Automated deployment through Docker containers

## How to Use

You can run the API with Docker:

```bash
docker run -it -p 8080:8080 \
  --name quirk \
  -e DB_PORT=5432 \
  -e DB_HOST=postgres \
  -e DB_USER= \
  -e DB_PASSWORD='<password>' \
  -e DB_NAME=quirkDB \
  -e CLOUDFLARE_API_TOKEN=<your_token> \
  abdulahshoaib/quirk
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `DB_HOST` | Hostname or container name for the PostgreSQL database |
| `DB_USER` | PostgreSQL username with access rights |
| `DB_PASSWORD` | Password for the specified PostgreSQL user |
| `DB_NAME` | Database name for Quirk (will be created if not present) |
| `CLOUDFLARE_API_TOKEN` | Cloudflare API token with access to Workers AI |
| `CLOUDFLARE_ACCOUNT_ID` | Cloudflare account ID |

## Authentication

All protected endpoints require JWT authentication. Include the token in the `Authorization` header:

```
Authorization: Bearer <your_token>
```

Tokens are validated against the database and must match the stored token for the user's email.

## API Endpoints

### `POST /signup`
Registers a new API token.

**Request Body:**
```json
{
  "email": "sample@mail.com"
}
```

**Response:**
```json
{
  "email": "sample@mail.com",
  "token": "<token>"
}
```

**Error Responses:**
- `405 Method Not Allowed` - Invalid request method
- `400 Bad Request` - JSON format of incoming request is invalid, or doesn't match user credentials structure
- `409 Conflict` - Email already exists
- `500 Internal Server Error` - Occurs for multiple reasons:
  - Failure to generate the authentication token
  - Problems storing the user's token in the database

### `POST /process`
Processes raw document, extracts and chunks content, and returns embeddings.

**Headers:** `Authorization: Bearer <token>`

**Request Body:** `multipart/form-data` with files

**Response:**
```json
{
  "object_id": "161c0955-72a7-4e7f-9300-fc5d676102f2"
}
```

**Error Responses:**
- `401 Unauthorized` - Missing or invalid token
- `405 Method Not Allowed` - Invalid request method
- `400 Bad Request` - Occurs for multiple reasons:
  - `Failed to parse` - The server couldn't parse the multipart form data you sent, or file size exceeding server limit (50 MB)
  - `No files uploaded` - Sent files not found in the "files" field of your multipart form
  - `Unsupported file extension` - Uploaded file extension isn't supported (supported formats are .pdf, .csv, .txt, .json, .md, .yml, and .xml)
  - `Unsupported file type` - The server can't process the content type
- `500 Internal Server Error` - Occurs for multiple reasons:
  - `File open error` - The server had trouble opening one of the uploaded files after receiving it
  - `Read error` - The server failed to read the content of an uploaded file
  - `Failed to process file` - Occurred during the internal processing of the file

### `GET /status?object_id={object_id}`
Returns processing status of a given object.

**Headers:** `Authorization: Bearer <token>`

**Response:**
```json
{
  "error_message": "",
  "eta_seconds": 0,
  "status": "completed"
}
```

**Error Responses:**
- `401 Unauthorized` - Missing or invalid token
- `400 Bad Request` - object_id not provided in the query parameters
- `404 Not Found` - object_id not found

### `GET /result?object_id={object_id}`
Returns the embedding results (vector, triples, filename, filecontent).

**Headers:** `Authorization: Bearer <token>`

**Response:**
```json
{
  "Embeddings": [
    [
      -0.0177764892578125,
      "...",
      -0.0077056884765625
    ]
  ],
  "Triples": null,
  "Filenames": [
    "sample.pdf"
  ],
  "Filecontent": [
    "uploaded file content"
  ]
}
```

**Error Responses:**
- `401 Unauthorized` - Missing or invalid token
- `400 Bad Request` - object_id not provided in the query parameters
- `404 Not Found` - object_id not found
- `202 Accepted` - Processing still in progress

### `GET /export?object_id={object_id}&format={format}`
Exports embedding results in the specified format.

**Headers:** `Authorization: Bearer <token>`

**Query Parameters:**
- `object_id` - The ID of the processed object
- `format` - Export format (`csv` or `json`)

**Response:**
- For `csv`: Returns CSV file with embeddings and triples
- For `json`: Returns JSON file with complete result data

**Error Responses:**
- `401 Unauthorized` - Missing or invalid token
- `400 Bad Request` - Occurs for multiple reasons:
  - Missing object_id parameter
  - Format unrecognized (must be `csv` or `json`)
- `404 Not Found` - Result not found for the given object_id

### `POST /export-chroma?object_id={object_id}&operation={operation}`
Exports embeddings directly to ChromaDB.

**Headers:** `Authorization: Bearer <token>`

**Query Parameters:**
- `object_id` - The ID of the processed object
- `operation` - ChromaDB operation (`add` or `update`)

**Request Body:**
```json
{
  "req": {
    "collection_name": "my_collection",
    "host": "localhost:8000"
  },
  "payload": {
    "metadatas": [
      {
        "source": "document1.pdf"
      }
    ]
  }
}
```

**Response:**
```
Chroma operation succeeded
```

**Error Responses:**
- `401 Unauthorized` - Missing or invalid token
- `400 Bad Request` - Occurs for multiple reasons:
  - Missing object_id parameter
  - Missing operation parameter
  - Invalid operation parameter (must be `add` or `update`)
  - Invalid JSON body
- `404 Not Found` - Embedding not found for object_id
- `500 Internal Server Error` - ChromaDB operation failed
