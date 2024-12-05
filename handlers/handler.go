package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PatientRegistration struct {
	UID              string    `json:"uid"`
	MRN              string    `json:"patientMrn"`
	RegNum           string    `json:"regNumber"`
	Name             string    `json:"patientName"`
	RegistrationDate time.Time `json:"registration_date"`
	Status           string    `json:"status"`
}

func ExampleHandler(pgDB *pgxpool.Pool, redisClient *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()

		// Example: Query PostgreSQL
		row := pgDB.QueryRow(ctx, "SELECT 'Hello from PostgreSQL!'")
		var message string
		if err := row.Scan(&message); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Example: Set key in Redis
		err := redisClient.Set(ctx, "example_key", "Hello from Redis!", 0).Err()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Response
		response := map[string]string{
			"postgres_message": message,
			"redis_status":     "Key set successfully",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func GetPatientRegistrations(pgDB *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()

		// Parse startDate and endDate from query string
		startDateStr := r.URL.Query().Get("startDate")
		endDateStr := r.URL.Query().Get("endDate")

		fmt.Println("Hello", startDateStr)
		fmt.Println("Hello", endDateStr)
		// Prepare query with conditions
		query := `
			SELECT 
				tpr.uid,
				tpr.mrn, 
				tpr.reg_num, 
				tp.name,
				tpr.registration_date
			FROM 
				t_patient_registration tpr
			INNER JOIN 
				t_patient tp 
				ON tp.mrn = tpr.mrn
		`

		// Add date filters if provided
		if startDateStr != "" && endDateStr != "" {
			query += " WHERE tpr.registration_date BETWEEN $1 AND $2"
		} else if startDateStr != "" {
			query += " WHERE tpr.registration_date >= $1"
		} else if endDateStr != "" {
			query += " WHERE tpr.registration_date <= $1"
		}

		// Execute query
		var rows pgx.Rows
		var err error

		// If dates are provided, use them in query
		if startDateStr != "" && endDateStr != "" {
			// Parse the dates from string to time.Time
			startDate, err := time.Parse("2006-01-02", startDateStr)
			if err != nil {
				http.Error(w, "Invalid startDate format. Expected YYYY-MM-DD.", http.StatusBadRequest)
				return
			}

			endDate, err := time.Parse("2006-01-02", endDateStr)
			if err != nil {
				http.Error(w, "Invalid endDate format. Expected YYYY-MM-DD.", http.StatusBadRequest)
				return
			}

			// Query with parameters
			rows, err = pgDB.Query(ctx, query, startDate, endDate)
		} else if startDateStr != "" {
			startDate, err := time.Parse("2006-01-02", startDateStr)
			if err != nil {
				http.Error(w, "Invalid startDate format. Expected YYYY-MM-DD.", http.StatusBadRequest)
				return
			}

			rows, err = pgDB.Query(ctx, query, startDate)
		} else if endDateStr != "" {
			endDate, err := time.Parse("2006-01-02", endDateStr)
			if err != nil {
				http.Error(w, "Invalid endDate format. Expected YYYY-MM-DD.", http.StatusBadRequest)
				return
			}

			rows, err = pgDB.Query(ctx, query, endDate)
		} else {
			// Query without filters
			rows, err = pgDB.Query(ctx, query)
		}

		if err != nil {
			http.Error(w, "Failed to execute query: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		// Parse query result for patient registrations
		var registrations []PatientRegistration
		for rows.Next() {
			var reg PatientRegistration
			if err := rows.Scan(&reg.UID, &reg.MRN, &reg.RegNum, &reg.Name, &reg.RegistrationDate); err != nil {
				http.Error(w, "Failed to parse query result: "+err.Error(), http.StatusInternalServerError)
				return
			}

			// Query to check patient examination status
			queryExam := `SELECT is_acc FROM t_patient_examination WHERE uid_registration = $1`
			examRows, err := pgDB.Query(ctx, queryExam, reg.UID)
			if err != nil {
				http.Error(w, "Failed to execute status query: "+err.Error(), http.StatusInternalServerError)
				return
			}
			defer examRows.Close()

			var status string
			var allAccTrue = true
			var someAccFalse = false

			// Check the examination status for this registration
			for examRows.Next() {
				var isAcc bool
				if err := examRows.Scan(&isAcc); err != nil {
					http.Error(w, "Failed to parse examination status: "+err.Error(), http.StatusInternalServerError)
					return
				}

				if !isAcc {
					allAccTrue = false
					someAccFalse = true
				}
			}

			// Determine the status based on `is_acc` values
			if allAccTrue {
				status = "DONE"
			} else if someAccFalse {
				status = "PROCESS"
			} else {
				status = "NOT_YET"
			}

			// Add the status to the registration
			reg.Status = status

			// Add registration to the list
			registrations = append(registrations, reg)
		}

		// Check for errors in rows iteration
		if rows.Err() != nil {
			http.Error(w, "Error iterating over rows: "+rows.Err().Error(), http.StatusInternalServerError)
			return
		}

		// Prepare the response
		response := map[string]interface{}{
			"message": "Success",
			"data":    registrations,
		}

		// Write JSON response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
