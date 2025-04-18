# firestore-batch-incrementer

Iterates through a Firestore collection in batches and atomically increments a specified root‑level numeric field with configurable rate limiting.

## Environment Variables

- `PROJECT_ID` – GCP project ID  
- `COLLECTION` – Firestore collection to scan  
- `FIELD_KEY` – Root‑level field to increment  
- `DATABASE_ID` – Firestore database ID (default: "(default)")
- `BATCH_SIZE` – Number of documents per batch (default: 100)  
- `RATE_LIMIT` – Maximum updates per second (default: 50)  
- `GOOGLE_APPLICATION_CREDENTIALS` – (Optional) Path to the service account JSON file. If not set, Application Default Credentials (ADC) will be used (e.g., when running in Cloud Run with an attached service account).
