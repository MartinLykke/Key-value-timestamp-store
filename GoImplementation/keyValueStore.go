package main

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/segmentio/kafka-go"
	bolt "go.etcd.io/bbolt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type keyValueTimestamp struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	Timestamp int64  `json:"timestamp"`
}

var db *bolt.DB

func main() {
	var err error
	db, err = bolt.Open("keyvaluetimestamp.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := mux.NewRouter()
	r.HandleFunc("/", putHandler).Methods("PUT")
	r.HandleFunc("/", getHandler).Methods("GET")

	log.Fatal(http.ListenAndServe(":8080", r))
}

func putHandler(w http.ResponseWriter, r *http.Request) {
	var kvt keyValueTimestamp
	err := json.NewDecoder(r.Body).Decode(&kvt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Produce the message to Kafka
	produceMessage(kvt)

	w.WriteHeader(http.StatusOK)
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Key       string `json:"key"`
		Timestamp int64  `json:"timestamp"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	value, err := getValue(req.Key, req.Timestamp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if value == "" {
		w.WriteHeader(http.StatusNotFound)
	} else {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, strconv.Quote(value))
	}
}

func produceMessage(kvt keyValueTimestamp) {
	// Kafka producer configuration
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "keyvaluetimestamp",
	})

	msg := kafka.Message{
		Key:   []byte(kvt.Key),
		Value: []byte(fmt.Sprintf("%s:%d", kvt.Value, kvt.Timestamp)),
	}

	err := writer.WriteMessages(context.Background(), msg)
	if err != nil {
		log.Println("Error producing message to Kafka:", err)
	}

	writer.Close()
}

func consumeMessages() {
	// Kafka consumer configuration
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{"localhost:9092"},
		Topic:     "keyvaluetimestamp",
		Partition: 0,
		MinBytes:  10e3, // 10KB
		MaxBytes:  10e6, // 10MB
	})

	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Println("Error consuming message from Kafka:", err)
			continue
		}

		key := string(msg.Key)
		valueTimestamp := string(msg.Value)
		sepIdx := strings.LastIndex(valueTimestamp, ":")
		value := valueTimestamp[:sepIdx]
		timestamp, _ := strconv.ParseInt(valueTimestamp[sepIdx+1:], 10, 64)

		err = storeValue(key, value, timestamp)
		if err != nil {
			log.Println("Error storing value:", err)
		}
	}
}

func storeValue(key, value string, timestamp int64) error {
	return db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(key))
		if err != nil {
			return err
		}
		tsBytes := make([]byte, 8)
		binary.BigEndian.PutUint64(tsBytes, uint64(timestamp))
		return bucket.Put(tsBytes, []byte(value))
	})
}

func getValue(key string, timestamp int64) (string, error) {
	var result string

	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(key))
		if bucket == nil {
			return nil
		}

		tsBytes := make([]byte, 8)
		binary.BigEndian.PutUint64(tsBytes, uint64(timestamp))

		c := bucket.Cursor()
		for k, v := c.Seek(tsBytes); k != nil; k, v = c.Prev() {
			result = string(v)
			break
		}
		return nil
	})

	return result, err
}
