package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"bytes"
	"encoding/json"
	"github.com/rs/xid"

	_ "modernc.org/sqlite"
)

var port = ":3000"

func newKey() string {
	key1 := xid.New().String()
	key2 := xid.New().String()
	final := key1 + "LATEPATCHES" + key2
	return final
}

func registerWebhook(db *sql.DB, webhook string, auth string) string {
	key := newKey()
	fmt.Println("[SERVER] Registering -> " + webhook + " | " + key)
	_, err := db.Exec(`
INSERT INTO webhooks VALUES (?, ?, ?)`, webhook, key, auth)
	if err != nil {
		fmt.Println("[SERVER] Error registering webhook: " + err.Error())
		return "Failed"
	}

	return key
}

func retrieveWebhook(db *sql.DB, key string) string {
	var webhook string

	err := db.QueryRow(
		`SELECT webhook FROM webhooks WHERE key = ?`,
		key,
	).Scan(&webhook)

	if err != nil {
		return "Failed To Retrieve"
	}

	return webhook
}

func verifyAuth(db *sql.DB, key string, auth string) bool {
    var correct string

    err := db.QueryRow(
        `SELECT auth FROM webhooks WHERE key = ?`,
        key,
    ).Scan(&correct); if err != nil { return false }

    return correct == auth
}

func main() {
	db, err := sql.Open("sqlite", "hooks.db"); if err != nil { return }
	mux := http.NewServeMux()
	defer db.Close()
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS webhooks (
        webhook TEXT UNIQUE,
        key TEXT UNIQUE,
	   auth TEXT
    )`)
	if err != nil {
		fmt.Println("[SERVER] Error creating table: " + err.Error())
	}

	mux.HandleFunc(
		"GET /register",
		func(w http.ResponseWriter, r *http.Request) {
			query := r.URL.Query()
			webhook := query.Get("webhook")
			auth := query.Get("auth")
			req := registerWebhook(db, webhook, auth)
			w.Write([]byte("Webhook Has Been Registered as: " + req + "\nAuthorization Header: " + auth))
		},
	)

	mux.HandleFunc("POST /webhook", func(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	auth := r.Header.Get("Authorization")

	if !verifyAuth(db, key, auth) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	webhook := retrieveWebhook(db, key)

	var data map[string]any
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	body, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "JSON encoding failed", http.StatusInternalServerError)
		return
	}
	req, err := http.NewRequest(http.MethodPost, webhook, bytes.NewReader(body))
	if err != nil {
		http.Error(w, "Invalid webhook URL", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "Webhook request failed", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(http.StatusAccepted)
	})
	http.ListenAndServe(port, mux)
}
