package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leamout/leamout/internal/security/secrets"
)

const rotationTimeout = 2 * time.Minute

type credentialRow struct {
	id       string
	outbound *string
	inbound  *string
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("rotate carrier credential encryption key: %v", err)
	}
}

func run() error {
	databaseURL := os.Getenv("DATABASE_URL")
	oldKey := os.Getenv("CARRIER_CREDENTIAL_ENCRYPTION_KEY_OLD")
	newKey := os.Getenv("CARRIER_CREDENTIAL_ENCRYPTION_KEY_NEW")
	if databaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if oldKey == "" || newKey == "" {
		return fmt.Errorf("CARRIER_CREDENTIAL_ENCRYPTION_KEY_OLD and CARRIER_CREDENTIAL_ENCRYPTION_KEY_NEW are required")
	}
	if oldKey == newKey {
		return fmt.Errorf("old and new carrier credential encryption keys must differ")
	}

	oldCipher, err := secrets.New(oldKey)
	if err != nil {
		return fmt.Errorf("initialize old key: %w", err)
	}
	newCipher, err := secrets.New(newKey)
	if err != nil {
		return fmt.Errorf("initialize new key: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), rotationTimeout)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return fmt.Errorf("begin rotation transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	rows, err := tx.Query(ctx, `
SELECT id::text, auth_secret_ciphertext, inbound_secret_ciphertext
FROM carrier_connections
WHERE auth_secret_ciphertext IS NOT NULL OR inbound_secret_ciphertext IS NOT NULL
ORDER BY id
FOR UPDATE`)
	if err != nil {
		return fmt.Errorf("lock carrier credentials: %w", err)
	}
	defer rows.Close()

	var credentials []credentialRow
	for rows.Next() {
		var row credentialRow
		if err := rows.Scan(&row.id, &row.outbound, &row.inbound); err != nil {
			return fmt.Errorf("scan carrier credential: %w", err)
		}
		credentials = append(credentials, row)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("read carrier credentials: %w", err)
	}
	rows.Close()

	updatedValues := 0
	alreadyRotated := 0
	for _, row := range credentials {
		outbound, outboundChanged, err := rotateValue(row.outbound, oldCipher, newCipher)
		if err != nil {
			return fmt.Errorf("carrier connection %s outbound credential: %w", row.id, err)
		}
		inbound, inboundChanged, err := rotateValue(row.inbound, oldCipher, newCipher)
		if err != nil {
			return fmt.Errorf("carrier connection %s inbound credential: %w", row.id, err)
		}
		if !outboundChanged && !inboundChanged {
			alreadyRotated++
			continue
		}
		if _, err := tx.Exec(ctx, `
UPDATE carrier_connections
SET auth_secret_ciphertext = $2,
    inbound_secret_ciphertext = $3,
    updated_at = now()
WHERE id = $1::uuid`, row.id, outbound, inbound); err != nil {
			return fmt.Errorf("update carrier connection %s: %w", row.id, err)
		}
		if outboundChanged {
			updatedValues++
		}
		if inboundChanged {
			updatedValues++
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit carrier credential rotation: %w", err)
	}
	log.Printf("carrier credential key rotation complete: connections=%d ciphertexts_reencrypted=%d already_on_new_key=%d", len(credentials), updatedValues, alreadyRotated)
	return nil
}

func rotateValue(value *string, oldCipher, newCipher *secrets.Cipher) (*string, bool, error) {
	if value == nil {
		return nil, false, nil
	}
	if _, err := newCipher.Decrypt(*value); err == nil {
		return value, false, nil
	}
	plaintext, err := oldCipher.Decrypt(*value)
	if err != nil {
		return nil, false, fmt.Errorf("ciphertext decrypts with neither old nor new key")
	}
	reencrypted, err := newCipher.Encrypt(plaintext)
	if err != nil {
		return nil, false, fmt.Errorf("encrypt with new key: %w", err)
	}
	if _, err := newCipher.Decrypt(reencrypted); err != nil {
		return nil, false, fmt.Errorf("verify new ciphertext: %w", err)
	}
	return &reencrypted, true, nil
}
