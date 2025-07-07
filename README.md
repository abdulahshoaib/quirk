# quirk

This is a RAG (Retrieval-Augmented Generation) data-processing pipeline that ingests raw data, converts
it to vector embeddings, and exports results as CSV format. The system accepts user data through two primary
input methods:
  * Direct submission via API endpoint
  * Automated data retrieval through external API

Upon receiving data, the system processes and embeds it into CSV format, structuring the information into rows and
columns for easy analysis and export.

## Core Functionality

Core functionality provides endpoints for data upload, batch processing, similarity search queries, and CSV export of embeddings
with associated metadata. The system includes an exposed HTTP endpoint for external integration and interaction.

## Data Formats

Supported input formats:
* .pdf
* .csv
* .json
* .doc

## RESTful API

The system provides:
* RESTful API endpoints for data submission and processing
* External API integration for automated data fetching
* Exposed HTTP endpoints for external system interaction

## Tech Stack

**Go**

Primary backend language

### merits:
* Built-in HTTP Server
* Support for CSV processing, HTTP handling, and file operations

**Docker**

Automated deployment through Docker containers
