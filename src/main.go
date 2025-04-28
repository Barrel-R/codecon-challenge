package main

import (
	"bufio"
	"encoding/json"
	"log"
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
	Status int         `json:"status"`
	Body   interface{} `json:"body"`
}

type StoreUsersResponse struct {
	Message    string `json:"message"`
	User_count int    `json:"user_count"`
}

type SuperuserResponse struct {
	Timestamp         time.Time     `json:"timestamp"`
	Execution_time_ms time.Duration `json:"execution_time_ms"`
	Data              []User        `json:"data"`
}

type TopCountriesResponse struct {
	Timestamp         time.Time     `json:"timestamp"`
	Execution_time_ms time.Duration `json:"execution_time_ms"`
	Countries         []Country     `json:"countries"`
}

type InsightsResponse struct {
	Timestamp         time.Time     `json:"timestamp"`
	Execution_time_ms time.Duration `json:"execution_time_ms"`
	Teams             []TeamInsight `json:"teams"`
}

type ActiveUsersPerDayResponse struct {
	Timestamp         time.Time     `json:"timestamp"`
	Execution_time_ms time.Duration `json:"execution_time_ms"`
	Logins            []Login       `json:"logins"`
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

const (
	TEST_USER_COUNT = 1000
)

var idCheck = make(map[uuid.UUID]bool)
var db = make([]User, TEST_USER_COUNT)

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

		if idCheck[user.Id] {
			continue
		}

		db = append(db, *user)
		idCheck[user.Id] = true
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

		data, err := json.Marshal(apiRes)

		if err != nil {
			log.Fatalf("Erro ao estruturar resposta da API: %v\n", err)
		}

		ctx.Data(200, "json", data)
	})
	r.GET("/superusers", func(ctx *gin.Context) {
		start := time.Now()
		users := getSuperusers()

		res := SuperuserResponse{
			Timestamp:         time.Now(),
			Execution_time_ms: time.Duration(time.Since(start).Milliseconds()),
			Data:              users,
		}

		apiRes := ApiResponse{
			200,
			res,
		}

		data, err := json.Marshal(apiRes)

		if err != nil {
			log.Fatalf("Erro ao estruturar dados de superusuários: %v\n", err)
		}

		ctx.Data(200, "json", data)
	})
	r.GET("/top-countries", func(ctx *gin.Context) {
		start := time.Now()
		countries := getTopCountries()

		res := TopCountriesResponse{
			Timestamp:         time.Now(),
			Execution_time_ms: time.Since(start),
			Countries:         countries,
		}

		apiRes := ApiResponse{
			200,
			res,
		}

		data, err := json.Marshal(apiRes)

		if err != nil {
			log.Fatalf("Erro ao estruturar dados dos países: %v", err)
		}

		ctx.Data(200, "json", data)
	})

	return r
}

func main() {
	r := setupRouter()
	// Listen and Server in 0.0.0.0:8080

	r.Run(":8080")
}
