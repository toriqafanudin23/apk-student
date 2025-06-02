package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type Student struct {
	Id        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Address   string    `json:"address"`
	BirthDate time.Time `json:"birth_date"`
	Gender    string    `json:"gender"`
}

var db *sql.DB

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	db = connectDb()
	defer db.Close()

	router := gin.Default()

	router.GET("/students", getAllStudents)
	router.GET("/students/:id", getStudentByID)
	router.POST("/students", addStudent)
	router.PUT("/students/:id", updateStudent)
	router.DELETE("/students/:id", deleteStudent)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	router.Run(":" + port)
}

func connectDb() *sql.DB {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to database")
	return db
}

// Handlers

func getAllStudents(c *gin.Context) {
	rows, err := db.Query("SELECT * FROM mst_student")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var students []Student

	for rows.Next() {
		var s Student
		if err := rows.Scan(&s.Id, &s.Name, &s.Email, &s.Address, &s.BirthDate, &s.Gender); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		students = append(students, s)
	}

	c.JSON(http.StatusOK, students)
}

func getStudentByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var s Student
	err = db.QueryRow("SELECT * FROM mst_student WHERE id = $1", id).Scan(&s.Id, &s.Name, &s.Email, &s.Address, &s.BirthDate, &s.Gender)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Student not found"})
		return
	}

	c.JSON(http.StatusOK, s)
}

func addStudent(c *gin.Context) {
	var s Student
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := "INSERT INTO mst_student (id, name, email, address, birth_date, gender) VALUES ($1, $2, $3, $4, $5, $6)"
	_, err := db.Exec(query, s.Id, s.Name, s.Email, s.Address, s.BirthDate, s.Gender)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, s)
}

func updateStudent(c *gin.Context) {
	id := c.Param("id")
	var s Student
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := "UPDATE mst_student SET name = $2, email = $3, address = $4, birth_date = $5, gender = $6 WHERE id = $1"
	_, err := db.Exec(query, id, s.Name, s.Email, s.Address, s.BirthDate, s.Gender)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, s)
}

func deleteStudent(c *gin.Context) {
	id := c.Param("id")

	_, err := db.Exec("DELETE FROM mst_student WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
