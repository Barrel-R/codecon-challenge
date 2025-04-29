package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"log"
	"maps"
	"net/http"
	"slices"
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

type ErrorResponse struct {
	Error         string `json:"error"`
	DetailedError string `json:"detailed_error"`
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
	Count             int       `json:"count"`
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
	Status         int   `json:"status"`
	Time_ms        int64 `json:"time_ms"`
	Valid_response bool  `json:"valid_response"`
}

type EvaluationMap map[string]Evaluation

type EvaluationErrorMap map[string]ErrorResponse

type EvaluationResponse struct {
	Timestamp             time.Time          `json:"timestamp"`
	Execution_time_ms     int64              `json:"execution_time_ms"`
	Tested_endpoints      EvaluationMap      `json:"tested_endpoints"`
	Endpoints_with_errors EvaluationErrorMap `json:"endpoints_with_errors"`
}

type HasExecutionTime interface {
	GetExecutionTime() int64
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

func (r SuperuserResponse) GetExecutionTime() int64 {
	return r.Execution_time_ms
}

func (r TopCountriesResponse) GetExecutionTime() int64 {
	return r.Execution_time_ms
}

func (r InsightsResponse) GetExecutionTime() int64 {
	return r.Execution_time_ms
}

func (r ActiveUsersPerDayResponse) GetExecutionTime() int64 {
	return r.Execution_time_ms
}

var db = make(map[uuid.UUID]User)

func storeUsersInMemory(ctx *gin.Context) (int, error) {
	formfile, err := ctx.FormFile("arquivos")

	if err != nil {
		return 0, errors.New("Erro ao ler arquivo JSON")
	}

	file, err := formfile.Open()

	if err != nil {
		return 0, errors.New("Erro ao abrir arquivo JSON")
	}

	defer file.Close()

	reader := bufio.NewReader(file)
	decoder := json.NewDecoder(reader)

	token, err := decoder.Token()

	if err != nil {
		return 0, errors.New("Erro ao ler inicío do JSON")
	}

	if delim, ok := token.(json.Delim); !ok || delim != '[' {
		return 0, errors.New("O array não foi inicializado corretamente.")
	}

	for decoder.More() {
		user := &User{}

		err := decoder.Decode(user)

		if err != nil {
			log.Printf("Erro ao salvar usuário na memória: %v\n", err.Error())
			continue
		}

		db[user.Id] = *user
	}

	token, err = decoder.Token()

	if err != nil {
		return 0, errors.New("Erro ao ler final do array.")
	}

	if delim, ok := token.(json.Delim); !ok || delim != ']' {
		return 0, errors.New("O array não foi finalizado corretamente.")
	}

	return len(db), nil
}

func getSuperusers() ([]User, error) {
	var superusers []User

	if len(db) == 0 {
		return superusers, errors.New("O arquivo de usuários não foi enviado ou processado, não há usuários na memória.")
	}

	for _, user := range db {
		if user.Score >= 900 && user.Ativo {
			superusers = append(superusers, user)
		}
	}

	return superusers, nil
}

func getTopCountries() ([]Country, error) {
	var countries []Country

	if len(db) == 0 {
		return countries, errors.New("O arquivo de usuários não foi enviado ou processado, não há usuários na memória.")
	}

	var countryMap = make(map[string]int)

	for _, user := range db {
		if user.Score >= 900 && user.Ativo {
			countryMap[user.Pais]++
		}
	}

	for country, count := range maps.All(countryMap) {
		if len(countries) >= 5 {
			return countries, nil
		}

		countryObj := Country{country, count}
		countries = append(countries, countryObj)
	}

	return countries, nil
}

func getTeamInsights() ([]TeamInsight, error) {
	var teamInsights []TeamInsight

	if len(db) == 0 {
		return teamInsights, errors.New("O arquivo de usuários não foi enviado ou processado, não há usuários na memória.")
	}

	var teamNames []string
	var memberCount = make(map[string]int)
	var leaderCount = make(map[string]int)
	var completedProjectsCount = make(map[string]int)
	var activeUsers = make(map[string]int)

	for _, user := range db {
		if !slices.Contains(teamNames, user.Equipe.Nome) {
			teamNames = append(teamNames, user.Equipe.Nome)
		}

		memberCount[user.Equipe.Nome]++

		if user.Equipe.Lider {
			leaderCount[user.Equipe.Nome]++
		}

		filteredProjects := Filter(user.Equipe.Projetos, func(project Project) bool {
			return project.Concluido
		})

		completedProjectsCount[user.Equipe.Nome] += len(filteredProjects)

		if user.Ativo {
			activeUsers[user.Equipe.Nome]++
		}
	}

	for _, teamName := range teamNames {

		activePercentage := float64(activeUsers[teamName]) / float64(memberCount[teamName]) * 100

		teamInsight := TeamInsight{
			Team:               teamName,
			Total_members:      memberCount[teamName],
			Leaders:            leaderCount[teamName],
			Completed_projects: completedProjectsCount[teamName],
			Active_percentage:  float64(activePercentage),
		}

		teamInsights = append(teamInsights, teamInsight)
	}

	return teamInsights, nil
}

func getActiveUsersPerDay() ([]Login, error) {
	var logins []Login

	if len(db) == 0 {
		return logins, errors.New("O arquivo de usuários não foi enviado ou processado, não há usuários na memória.")
	}

	var dayMap = make(map[CustomTime]int)

	for _, user := range db {
		for _, log := range user.Logs {
			if log.Acao != "login" {
				continue
			}
			dayMap[log.Data]++
		}
	}

	for date, count := range maps.All(dayMap) {
		login := Login{
			Date:  date,
			Total: count,
		}

		logins = append(logins, login)
	}

	return logins, nil
}

func storeUsersHandler(ctx *gin.Context) {
	count, err := storeUsersInMemory(ctx)

	var status int
	var res any

	if err != nil {
		status = http.StatusBadRequest

		res = ErrorResponse{
			Error: err.Error(),
		}
	} else {
		status = http.StatusOK

		res = StoreUsersResponse{
			"Arquivo salvo na memória com sucesso.",
			count,
		}
	}

	apiRes := ApiResponse{
		status,
		res,
	}

	ctx.JSON(status, apiRes)
}

func superusersHandler(ctx *gin.Context) {
	start := time.Now()
	users, err := getSuperusers()

	var status int
	var res any

	if err != nil {
		status = http.StatusBadRequest

		res = ErrorResponse{
			Error: err.Error(),
		}
	} else {
		status = http.StatusOK

		res = SuperuserResponse{
			Timestamp:         time.Now(),
			Execution_time_ms: time.Since(start).Milliseconds(),
			Data:              users,
			Count:             len(users),
		}
	}

	apiRes := ApiResponse{
		status,
		res,
	}

	ctx.JSON(status, apiRes)
}

func topCountriesHandler(ctx *gin.Context) {
	start := time.Now()
	countries, err := getTopCountries()

	var status int
	var res any

	if err != nil {
		status = http.StatusBadRequest

		res = ErrorResponse{
			Error: err.Error(),
		}
	} else {
		status = http.StatusOK

		res = TopCountriesResponse{
			Timestamp:         time.Now(),
			Execution_time_ms: time.Since(start).Milliseconds(),
			Countries:         countries,
		}
	}

	apiRes := ApiResponse{
		status,
		res,
	}

	ctx.JSON(status, apiRes)
}

func teamInsightsHandler(ctx *gin.Context) {
	start := time.Now()
	insights, err := getTeamInsights()

	var status int
	var res any

	if err != nil {
		status = http.StatusBadRequest

		res = ErrorResponse{
			Error: err.Error(),
		}
	} else {
		status = http.StatusOK

		res = InsightsResponse{
			Timestamp:         time.Now(),
			Execution_time_ms: time.Since(start).Milliseconds(),
			Teams:             insights,
		}
	}

	apiRes := ApiResponse{
		status,
		res,
	}

	ctx.JSON(status, apiRes)
}

func activeUsersHandler(ctx *gin.Context) {
	start := time.Now()
	activeUsers, err := getActiveUsersPerDay()

	var status int
	var res any

	if err != nil {
		status = http.StatusBadRequest

		res = ErrorResponse{
			Error: err.Error(),
		}
	} else {
		status = http.StatusOK

		res = ActiveUsersPerDayResponse{
			Timestamp:         time.Now(),
			Execution_time_ms: time.Since(start).Milliseconds(),
			Logins:            activeUsers,
		}
	}

	apiRes := ApiResponse{
		status,
		res,
	}

	ctx.JSON(status, apiRes)
}

func performEvaluation(endpoint string) (Evaluation, error) {
	start := time.Now()
	res, err := http.Get("http://localhost:8080" + endpoint)

	if err != nil {
		log.Printf("Erro ao enviar requisição de avaliação: %v\n", err)
		return Evaluation{}, err
	}

	defer res.Body.Close()

	evaluation := Evaluation{
		Status:         res.StatusCode,
		Time_ms:        time.Since(start).Milliseconds(),
		Valid_response: err == nil,
	}

	return evaluation, err
}

func evaluateEndpoints() (EvaluationMap, EvaluationErrorMap) {
	endpoints := []string{
		"/superusers",
		"/top-countries",
		"/team-insights",
		"/active-users-per-day",
	}

	evalMap := make(EvaluationMap)
	errMap := make(EvaluationErrorMap)

	for _, endpoint := range endpoints {
		eval, err := performEvaluation(endpoint)

		if err != nil {
			errMap[endpoint] = ErrorResponse{
				Error:         "Não foi possível obter a avaliação do endpoint " + endpoint,
				DetailedError: err.Error(),
			}
		}

		evalMap[endpoint] = eval
	}

	return evalMap, errMap
}

func evaluationHandler(ctx *gin.Context) {
	if len(db) == 0 {
		res := ErrorResponse{
			Error: "O arquivo de usuários não foi enviado ou processado, não há usuários na memória.",
		}

		apiRes := ApiResponse{
			http.StatusBadRequest,
			res,
		}

		ctx.JSON(http.StatusBadRequest, apiRes)
		return
	}

	start := time.Now()
	evalMap, errMap := evaluateEndpoints()

	res := EvaluationResponse{
		Timestamp:             time.Now(),
		Execution_time_ms:     time.Since(start).Milliseconds(),
		Tested_endpoints:      evalMap,
		Endpoints_with_errors: errMap,
	}

	apiRes := ApiResponse{
		http.StatusOK,
		res,
	}

	ctx.JSON(http.StatusOK, apiRes)
}

func setupRouter() *gin.Engine {
	r := gin.Default()

	r.POST("/users", storeUsersHandler)
	r.GET("/superusers", superusersHandler)
	r.GET("/top-countries", topCountriesHandler)
	r.GET("/team-insights", teamInsightsHandler)
	r.GET("/active-users-per-day", activeUsersHandler)
	r.GET("/evaluation", evaluationHandler)

	return r
}

func main() {
	r := setupRouter()
	// Listen and Server in 0.0.0.0:8080

	r.Run(":8080")
}
