[![License](https://img.shields.io/badge/license-MIT-blue)](https://opensource.org/licenses/MIT) [![Work In Progress](https://img.shields.io/badge/Status-Work%20In%20Progress-yellow)](https://guide.unitvectorylabs.com/bestpractices/status/#work-in-progress)


# firestore-batch-incrementer

Iterates through a Firestore collection in batches and atomically increments a specified root‑level numeric field with configurable rate limiting.

## Purpose

Loops through every document in a specified Firestore collection and atomically increments a given root‑level numeric field—adding the field with a count of 1 if it doesn’t exist—to bulk-update all records so that downstream Eventarc triggers fire on each change.

This application runs scanning through the collection, updating each document at the specified rate limit, and then exits. It is intended to run as a Cloud Run Job.

## Environment Variables

- `PROJECT_ID` – GCP project ID  
- `COLLECTION` – Firestore collection to scan  
- `FIELD_KEY` – Root‑level field to increment  
- `DATABASE_ID` – Firestore database ID (default: "(default)")
- `BATCH_SIZE` – Number of documents per batch (default: 100)  
- `RATE_LIMIT` – Maximum updates per second (default: 50)
- `ATOMIC_UPDATES` – Use atomic transactions for updates (ensures document existence, default: false)
- `GOOGLE_APPLICATION_CREDENTIALS` – (Optional) Path to the service account JSON file. If not set, Application Default Credentials (ADC) will be used (e.g., when running in Cloud Run Jobs with an attached service account).
