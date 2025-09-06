package data

import (
	"database/sql"
	"errors"
	"github.com/aikwen/greenlight/internal/validator"
	"github.com/lib/pq"
	"time"
)

type Movie struct {
	ID       int64     `json:"id"`
	CreateAt time.Time `json:"create_at"`
	Title    string    `json:"title"`
	Year     int32     `json:"year"`
	Runtime  Runtime   `json:"runtime"`
	Genres   []string  `json:"genres"`
	Version  int32     `json:"version"`
}

func ValidateMovie(v *validator.Validator, movie *Movie) {
	// title
	v.Check(movie.Title != "", "title", "must be provided")
	v.Check(len(movie.Title) <= 500, "title", "must not be more than 500 bytes")

	// year
	v.Check(movie.Year != 0, "year", "must be provided")
	v.Check(movie.Year >= 1888, "year", "must be greater than 1888")
	v.Check(movie.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	// runtime
	v.Check(movie.Runtime != 0, "runtime", "must be provided")
	v.Check(movie.Runtime > 0, "runtime", "must be a positive integer")

	// genres
	v.Check(movie.Genres != nil, "genres", "must be provided")
	v.Check(len(movie.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(movie.Genres) <= 5, "genres", "must not contain more than 5 genres")
	v.Check(validator.Unique(movie.Genres), "genres", "must not contain duplicate genres")
}

// MovieModel Define a MovieModel
type MovieModel struct {
	DB *sql.DB
}

// Insert method for inserting a new record in the movie table
func (m MovieModel) Insert(movie *Movie) error {
	query := `
			INSERT INTO movies (title, year, runtime, genres)
			VALUES ($1, $2, $3, $4)
			RETURNING id, created_at, version`
	args := []interface{}{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}

	return m.DB.QueryRow(query, args...).Scan(&movie.ID, &movie.CreateAt, &movie.Version)
}

// Get method for fetching a specific record
func (m MovieModel) Get(id int64) (*Movie, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, created_at, title, year, runtime, genres, version
		FROM movies
		WHERE id = $1`

	var movie Movie
	err := m.DB.QueryRow(query, id).Scan(
		&movie.ID,
		&movie.CreateAt,
		&movie.Title,
		&movie.Year,
		&movie.Runtime,
		pq.Array(&movie.Genres),
		&movie.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &movie, nil
}

// Update method for updating a specific record
func (m MovieModel) Update(movie *Movie) error {
	query := `
		UPDATE movies
		SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1
		WHERE id = $5
		RETURNING version`

	args := []interface{}{
		movie.Title,
		movie.Year,
		movie.Runtime,
		pq.Array(movie.Genres),
		movie.ID}

	return m.DB.QueryRow(query, args...).Scan(&movie.Version)
}

// Delete method for deleting a specific record
func (m MovieModel) Delete(id int64) error {
	// check id
	if id < 1 {
		return ErrRecordNotFound
	}
	// sql
	query := `DELETE FROM movies 
       WHERE id = $1`
	// execute the sql query
	result, err := m.DB.Exec(query, id)
	if err != nil {
		return err
	}
	// get the number of rows affected by the query
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	// no row were affected
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}
