package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"cloud.google.com/go/firestore"
	"github.com/kelseyhightower/envconfig"
	"golang.org/x/time/rate"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Config struct {
	ProjectID     string  `envconfig:"PROJECT_ID" required:"true"`
	Collection    string  `envconfig:"COLLECTION" required:"true"`
	FieldKey      string  `envconfig:"FIELD_KEY" required:"true"`
	DatabaseID    string  `envconfig:"DATABASE_ID" default:"(default)"`
	BatchSize     int     `envconfig:"BATCH_SIZE" default:"100"`
	RateLimit     float64 `envconfig:"RATE_LIMIT" default:"50"`
	AtomicUpdates bool    `envconfig:"ATOMIC_UPDATES" default:"false"`
}

func main() {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatalf("failed to process env vars: %v", err)
	}

	ctx := context.Background()

	var opts []option.ClientOption
	credsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if credsPath != "" {
		opts = append(opts, option.WithCredentialsFile(credsPath))
	}

	client, err := firestore.NewClientWithDatabase(ctx, cfg.ProjectID, cfg.DatabaseID, opts...)
	if err != nil {
		log.Fatalf("failed to create firestore client: %v", err)
	}
	defer client.Close()

	limiter := rate.NewLimiter(rate.Limit(cfg.RateLimit), int(cfg.RateLimit))

	var lastDoc *firestore.DocumentSnapshot
	for {

		q := client.Collection(cfg.Collection).OrderBy(firestore.DocumentID, firestore.Asc).Limit(cfg.BatchSize)
		if lastDoc != nil {
			q = q.StartAfter(lastDoc)
		}
		docs, err := q.Documents(ctx).GetAll()
		if err != nil {
			log.Fatalf("failed to fetch batch: %v", err)
		}
		if len(docs) == 0 {
			log.Println("no more documents, exiting")
			break
		}

		var wg sync.WaitGroup
		for _, d := range docs {
			wg.Add(1)

			if err := limiter.Wait(ctx); err != nil {
				log.Fatalf("rate limiter error: %v", err)
			}
			go func(doc *firestore.DocumentSnapshot) {
				defer wg.Done()

				if cfg.AtomicUpdates {

					err := client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {

						_, getErr := tx.Get(doc.Ref)
						if getErr != nil {
							if status.Code(getErr) == codes.NotFound {

								log.Printf("document %s does not exist, skipping update", doc.Ref.Path)
								return nil
							}

							return fmt.Errorf("failed to get document %s in transaction: %w", doc.Ref.Path, getErr)
						}

						updateErr := tx.Update(doc.Ref, []firestore.Update{{
							Path:  cfg.FieldKey,
							Value: firestore.Increment(1),
						}})
						if updateErr != nil {
							return fmt.Errorf("failed to update document %s in transaction: %w", doc.Ref.Path, updateErr)
						}
						return nil
					})

					if err != nil {

						log.Printf("transaction failed for %s: %v", doc.Ref.Path, err)
					}
				} else {

					_, err := doc.Ref.Update(ctx, []firestore.Update{{
						Path:  cfg.FieldKey,
						Value: firestore.Increment(1),
					}})
					if err != nil {
						log.Printf("direct update failed for %s: %v", doc.Ref.Path, err)
					}
				}
			}(d)
		}
		wg.Wait()

		log.Printf("Processed page with %d documents", len(docs))

		lastDoc = docs[len(docs)-1]
	}
}
