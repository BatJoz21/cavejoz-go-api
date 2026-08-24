package models

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"time"

	"github.com/BatJoz21/cavejoz-go-api/databases"
)

type WSTicket struct {
	Value    string
	ExpireAt time.Time
}

func GenerateWSTicket() (*WSTicket, error) {
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	token := base64.RawURLEncoding.EncodeToString(randomBytes)

	return &WSTicket{
		Value:    token,
		ExpireAt: time.Now().Add(30 * time.Second),
	}, nil
}

func IssueWSTicket(uID int64) (string, error) {
	wst, err := GenerateWSTicket()
	if err != nil {
		return "", err
	}

	query := `INSERT INTO ws_tickets(ticket, user_id, expire_at) VALUES (?, ?, ?)`
	_, err = databases.DB.Exec(query, wst.Value, uID, wst.ExpireAt)
	if err != nil {
		return "", err
	}

	return wst.Value, nil
}

func ConsumeWSTicket(ticket string) (int64, bool) {
	var uID int64
	var expireAt time.Time

	err := databases.DB.QueryRow(`SELECT user_id, expire_at FROM ws_tickets WHERE ticket = ?`, ticket).Scan(&uID, &expireAt)
	if err != nil {
		return 0, false
	}

	if time.Now().After(expireAt) {
		return 0, false
	}

	result, err := databases.DB.Exec(`DELETE FROM ws_tickets WHERE ticket = ?`, ticket)
	if err != nil {
		return 0, false
	}

	rows, _ := result.RowsAffected()
	return uID, rows == 1
}

// DeleteExpiredWSTickets reaps tickets whose expiry has passed. Consumed tickets
// are already deleted by ConsumeWSTicket; this clears the ones nobody ever used.
func DeleteExpiredWSTickets() (int64, error) {
	result, err := databases.DB.Exec(`DELETE FROM ws_tickets WHERE expire_at < ?`, time.Now())
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// StartWSTicketSweeper sweeps expired tickets once, then on every tick for the
// lifetime of the process.
func StartWSTicketSweeper(every time.Duration) {
	go func() {
		ticker := time.NewTicker(every)
		defer ticker.Stop()

		for {
			deleted, err := DeleteExpiredWSTickets()
			if err != nil {
				log.Printf("ws ticket sweep failed: %v", err)
			} else if deleted > 0 {
				log.Printf("ws ticket sweep removed %d expired tickets", deleted)
			}

			<-ticker.C
		}
	}()
}
