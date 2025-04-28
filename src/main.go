package main

import (
	"bufio"
	"encoding/json"
	"log"
	"maps"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Log struct {
	Data CustomTime `json:"data"`
	Acao string     `json:"acao"`
}

type Project struct {
	Nome      string `json:"nome"`
	Concluido bool   `json:"concluido"`
}

type Team struct {
	Nome     string    `json:"nome"`
	Lider    bool      `json:"lider"`
	Projetos []Project `json:"projetos"`
}

type TeamInsight struct {
	Team               string  `json:"team"`
	Total_members      int     `json:"total_members"`
	Leaders            int     `json:"leaders"`
	Completed_projects int     `json:"completed_projects"`
	Active_percentage  float64 `json:"active_percentage"`
}

type User struct {
	Id     uuid.UUID `json:"id"`
	Nome   string    `json:"nome"`
	Idade  int       `json:"idade"`
	Score  int       `json:"score"`
	Ativo  bool      `json:"ativo"`
	Pais   string    `json:"pais"`
	Equipe Team      `json:"equipe"`
	Logs   []Log     `json:"logs"`
}

type Country struct {
	Country string `json:"country"`
	Total   int    `json:"total"`
}

type Login struct {
	Date  CustomTime `json:"date"`
	Total int        `json:"total"`
}

type ApiResponse struct {
	Status int `json:"status"`
	Body   any `json:"body"`
}

type StoreUsersResponse struct {
	Message    string `json:"message"`
	User_count int    `json:"user_count"`
}

type SuperuserResponse struct {
	Timestamp         time.Time `json:"timestamp"`
	Execution_time_ms int64     `json:"execution_time_ms"`
	Data              []User    `json:"data"`
}

type TopCountriesResponse struct {
	Timestamp         time.Time `json:"timestamp"`
	Execution_time_ms int64     `json:"execution_time_ms"`
	Countries         []Country `json:"countries"`
}

type InsightsResponse struct {
	Timestamp         time.Time     `json:"timestamp"`
	Execution_time_ms int64         `json:"execution_time_ms"`
	Teams             []TeamInsight `json:"teams"`
}

type ActiveUsersPerDayResponse struct {
	Timestamp         time.Time `json:"timestamp"`
	Execution_time_ms int64     `json:"execution_time_ms"`
	Logins            []Login   `json:"logins"`
}

type Evaluation struct {
	Status         int  `json:"status"`
	Time_ms        int  `json:"time_ms"`
	Valid_response bool `json:"valid_response"`
}

type EvaluationEndpoints map[string]Evaluation

type EvaluationResponse struct {
	Tested_endpoints EvaluationEndpoints
}

type CustomTime struct {
	time.Time
}

func (t *CustomTime) UnmarshalJSON(b []byte) (err error) {
	date, err := time.Parse(`"2006-01-02"`, string(b))

	if err != nil {
		return err
	}

	t.Time = date
	return
}

var db = make(map[uuid.UUID]User)

func storeUsersInMemory(ctx *gin.Context) int {
	formfile, err := ctx.FormFile("arquivos")

	if err != nil {
		log.Fatalf("Erro ao ler arquivo JSON: %v\n", err)
	}

	file, err := formfile.Open()

	if err != nil {
		log.Fatalf("Erro ao abrir arquivo JSON: %v\n", err)
	}

	defer file.Close()

	reader := bufio.NewReader(file)
	decoder := json.NewDecoder(reader)

	token, err := decoder.Token()

	if err != nil {
		log.Fatalf("Erro ao ler inicío do JSON: %v\n", err)
	}

	if delim, ok := token.(json.Delim); !ok || delim != '[' {
		log.Fatalf("O array não foi inicializado corretamente.")
	}

	for decoder.More() {
		user := &User{}

		err := decoder.Decode(user)

		if err != nil {
			log.Fatalf("Erro ao salvar usuário na memória: %v\n", err.Error())
		}

		db[user.Id] = *user
	}

	token, err = decoder.Token()

	if err != nil {
		log.Fatalf("Erro ao ler final do array: %v\n", err)
	}

	if delim, ok := token.(json.Delim); !ok || delim != ']' {
		log.Fatalf("O array não foi finalizado corretamente.")
	}

	return len(db)
}

func getSuperusers() []User {
	var superusers []User

	for _, user := range db {
		if user.Score >= 900 && user.Ativo {
			superusers = append(superusers, user)
		}
	}

	return superusers
}

func getTopCountries() []Country {
	var countries []Country
	var countryMap = make(map[string]int)

	for _, user := range db {
		if user.Score >= 900 && user.Ativo {
			countryMap[user.Pais]++
		}
	}

	for country, count := range maps.All(countryMap) {
		countryObj := Country{country, count}
		countries = append(countries, countryObj)
	}

	return countries
}

func setupRouter() *gin.Engine {
	r := gin.Default()

	r.POST("/users", func(ctx *gin.Context) {
		count := storeUsersInMemory(ctx)

		res := StoreUsersResponse{
			"Arquivo salvo na memória com sucesso.",
			count,
		}

		apiRes := ApiResponse{
			200,
			res,
		}

		ctx.JSON(200, apiRes)
	})
	r.GET("/superusers", func(ctx *gin.Context) {
		start := time.Now()
		users := getSuperusers()

		res := SuperuserResponse{
			Timestamp:         time.Now(),
			Execution_time_ms: time.Since(start).Milliseconds(),
			Data:              users,
		}

		apiRes := ApiResponse{
			200,
			res,
		}

		ctx.JSON(200, apiRes)
	})
	r.GET("/top-countries", func(ctx *gin.Context) {
		start := time.Now()
		countries := getTopCountries()

		res := TopCountriesResponse{
			Timestamp:         time.Now(),
			Execution_time_ms: time.Since(start).Milliseconds(),
			Countries:         countries,
		}

		apiRes := ApiResponse{
			200,
			res,
		}

		ctx.JSON(200, apiRes)
	})

	return r
}

func main() {
	r := setupRouter()
	// Listen and Server in 0.0.0.0:8080

	r.Run(":8080")
}
