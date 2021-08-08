package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

var templates *template.Template

type PageBag struct {
	Favicon   string
	JsAssets  []string
	CssAssets []string
}

type Question struct {
	Id                      int64          `json:"id"`
	Title                   string         `json:"title"`
	FirstOptionDescription  string         `json:"first_option_description"`
	FirstOptionVotes        int64          `json:"first_option_votes"`
	SecondOptionDescription string         `json:"second_option_description"`
	SecondOptionVotes       int64          `json:"second_option_votes"`
	Details                 sql.NullString `json:"details"`
}

func main() {
	fmt.Println("Server starting..")

	templates = template.Must(template.ParseGlob("templates/*.html"))
	r := mux.NewRouter()
	assets := http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets/")))
	r.HandleFunc("/", indexHandler).Methods("GET")
	r.HandleFunc("/question", questionHandler).Methods("GET")
	r.HandleFunc("/question/answer", answerQuestionHandler).Methods("POST")
	r.PathPrefix("/assets/").Handler(assets)
	http.Handle("/", r)

	fmt.Println("Handlers setted")

	port := os.Getenv("PORT")
	fmt.Println("Will listen in port: ", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	p := PageBag{
		Favicon:   "assets/images/favicon.png",
		JsAssets:  []string{"assets/js/main.js"},
		CssAssets: []string{"assets/css/custom.css"},
	}
	templates.ExecuteTemplate(w, "index.html", p)
}

func questionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	question, err := getRandomQuestion()
	encoder := json.NewEncoder(w)
	if err != nil {
		w.WriteHeader(404)
		w.Write([]byte(`Question was not found. :(`))
		return
	}
	encoder.Encode(question)
}

func answerQuestionHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	formData := r.Form
	id, err := strconv.ParseInt(formData.Get("id"), 10, 64)
	option := formData.Get("option")
	question, err := getQuestion(id)
	if err != nil {
		w.WriteHeader(404)
		w.Write([]byte(`Question was not found. :(`))
		return
	}

	question, err = addQuestionVote(question, option)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(`Something goes wrong :(`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	if err != nil {
		w.WriteHeader(404)
		w.Write([]byte(`Question was not found. :(`))
		return
	}
	encoder.Encode(question)
}

func dbConnection() sql.DB {

	DBUser := os.Getenv("MYSQL_USER")
	DBPassword := os.Getenv("MYSQL_PASSWORD")
	DBHost := os.Getenv("MYSQL_HOST")
	DBPort := os.Getenv("MYSQL_PORT")
	DBName := os.Getenv("MYSQL_DB")

	for i, k := range []string{DBUser, DBPassword, DBHost, DBPort, DBName} {
		if k == "" {
			fmt.Println("Missing database env key ", i)
		}
	}

	db, err := sql.Open("mysql", DBUser+":"+DBPassword+"@tcp("+DBHost+":"+DBPort+")/"+DBName)
	if err != nil {
		fmt.Println("Something goes wrong at sql connection.", err)
	}
	return *db
}

func getRandomQuestion() (Question, error) {
	db := dbConnection()
	var question Question
	row := db.QueryRow("SELECT * FROM questions ORDER BY RAND() LIMIT 1")
	err := row.Scan(&question.Id, &question.Title, &question.FirstOptionDescription, &question.FirstOptionVotes, &question.SecondOptionDescription, &question.SecondOptionVotes, &question.Details)
	defer db.Close()
	if err != nil {
		fmt.Println("Could not get any random question :( ", err)
		return Question{}, err
	}
	return question, nil
}

func getQuestion(id int64) (Question, error) {
	db := dbConnection()
	var question Question
	row := db.QueryRow("SELECT * FROM questions WHERE ID = ?", id)
	err := row.Scan(&question.Id, &question.Title, &question.FirstOptionDescription, &question.FirstOptionVotes, &question.SecondOptionDescription, &question.SecondOptionVotes, &question.Details)
	defer db.Close()
	if err != nil {
		fmt.Println("Could not get question by ID: ", id)
		return Question{}, err
	}
	return question, nil
}

func addQuestionVote(question Question, option string) (Question, error) {
	db := dbConnection()

	var column string
	var votes int64
	if option == "first-option" {
		column = "first_option_votes"
		votes = question.FirstOptionVotes
		question.FirstOptionVotes++
	} else if option == "second-option" {
		column = "second_option_votes"
		votes = question.SecondOptionVotes
		question.SecondOptionVotes++
	} else {
		fmt.Println("Invalid option >:|")
	}

	stmt := "UPDATE questions SET " + column + "=? WHERE ID=?"
	result, err := db.Exec(stmt, votes+1, question.Id)
	defer db.Close()
	if err != nil {
		fmt.Println("Could not add vote :( ", err)
		return Question{}, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	if rows != 1 {
		log.Fatalf("expected to affect 1 row, affected %d", rows)
	}
	return question, nil
}
