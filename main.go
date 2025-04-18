package main

import (
	"context"
	"log"
	"os"
	"time"

	"sync"

	"cloud.google.com/go/firestore"
	"github.com/kelseyhightower/envconfig"
	"golang.org/x/time/rate"
	"google.golang.org/api/option"
)

type Config struct {
	ProjectID  string  `envconfig:"PROJECT_ID" required:"true"`
	Collection string  `envconfig:"COLLECTION" required:"true"`
	FieldKey   string  `envconfig:"FIELD_KEY" required:"true"`
	DatabaseID string  `envconfig:"DATABASE_ID" default:"(default)"` // Add DatabaseID field
	BatchSize  int     `envconfig:"BATCH_SIZE" default:"100"`
	RateLimit  float64 `envconfig:"RATE_LIMIT" default:"50"`
}

func main() {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatalf("failed to process env vars: %v", err)
	}

	ctx := context.Background()

	// Conditionally add credentials option
	var opts []option.ClientOption
	credsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if credsPath != "" {
		opts = append(opts, option.WithCredentialsFile(credsPath))
	}

	// Update NewClient call to use cfg.DatabaseID and conditional options
	client, err := firestore.NewClientWithDatabase(ctx, cfg.ProjectID, cfg.DatabaseID, opts...)
	if err != nil {
		log.Fatalf("failed to create firestore client: %v", err)
	}
	defer client.Close()

	limiter := rate.NewLimiter(rate.Limit(cfg.RateLimit), int(cfg.RateLimit))

	// Pagination cursor
	var lastDoc *firestore.DocumentSnapshot
	for {
		q := client.Collection(cfg.Collection).Limit(cfg.BatchSize)
		if lastDoc != nil {
			q = q.StartAfter(lastDoc.Ref)
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
			// rate-limit
			if err := limiter.Wait(ctx); err != nil {
				log.Fatalf("rate limiter error: %v", err)
			}
			go func(doc *firestore.DocumentSnapshot) {
				defer wg.Done()
				_, err := doc.Ref.Update(ctx, []firestore.Update{{
					Path:  cfg.FieldKey,
					Value: firestore.Increment(1),
				}})
				if err != nil {
					log.Printf("update failed for %s: %v", doc.Ref.Path, err)
				}
			}(d)
		}
		wg.Wait()

		// prepare next cursor
		lastDoc = docs[len(docs)-1]
		// small pause to allow any pending logs
		time.Sleep(200 * time.Millisecond)
	}
}
