# firestore-batch-incrementer

Iterates through a Firestore collection in batches and atomically increments a specified root‑level numeric field with configurable rate limiting.

## Environment Variables

- `PROJECT_ID` – GCP project ID  
- `COLLECTION` – Firestore collection to scan  
- `FIELD_KEY` – Root‑level field to increment  
- `BATCH_SIZE` – Number of documents per batch (default: 100)  
- `RATE_LIMIT` – Maximum updates per second (default: 50)  
- `GOOGLE_APPLICATION_CREDENTIALS` – Path to the service account JSON file  
