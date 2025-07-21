package integration

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"go.uber.org/zap"

	"github.com/TemirB/rest-api-marketplace/internal/auth"
	"github.com/TemirB/rest-api-marketplace/internal/database"
	"github.com/TemirB/rest-api-marketplace/internal/middleware"
	"github.com/TemirB/rest-api-marketplace/internal/post"
	"github.com/TemirB/rest-api-marketplace/pkg/jwt"
)

var testServer *httptest.Server
var db *sql.DB

func TestMain(m *testing.M) {
	// 1. Запуск контейнера Postgres
	ctx := context.Background()
	pgContainer, err := postgres.Run(ctx, "postgres:13-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		postgres.WithInitScripts("../migrations/init.sql"),
	)
	if err != nil {
		panic(err)
	}

	uri, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		panic(err)
	}

	db, err = sql.Open("postgres", uri)
	if err != nil {
		panic(err)
	}

	repo := &database.Repository{DB: db, Logger: zap.NewNop()}
	userStore := auth.NewStorage(repo, zap.NewNop())
	postStore := post.NewStorage(repo, zap.NewNop())
	tokenManager := jwt.New("testsecret", time.Minute*60)
	authService := auth.NewService(userStore, tokenManager, zap.NewNop())
	postService := post.NewService(postStore, zap.NewNop())
	authHandler := auth.NewHandler(authService, zap.NewNop())
	postHandler := post.NewHandler(postService, zap.NewNop())

	mux := http.NewServeMux()
	mux.HandleFunc("/register", authHandler.Register)
	mux.HandleFunc("/login", authHandler.Login)

	protectedPost := middleware.JWTAuthMiddleware(authService)
	mux.Handle("/posts", protectedPost(http.HandlerFunc(postHandler.CreatePost)))
	mux.HandleFunc("/posts/feed", postHandler.GetPosts)

	testServer = httptest.NewServer(mux)
	code := m.Run()
	testServer.Close()
	pgContainer.Terminate(ctx)
	os.Exit(code)
}

func TestFullScenario(t *testing.T) {
	client := testServer.Client()
	baseURL := testServer.URL

	// 1. Регистрация
	regBody := strings.NewReader(`{"login":"alice","password":"password123"}`)
	regResp, err := client.Post(baseURL+"/register", "application/json", regBody)
	if err != nil {
		t.Fatalf("Register request failed: %v", err)
	}
	assert.Equal(t, http.StatusCreated, regResp.StatusCode)
	// можно проверить тело:
	var regRespBody map[string]string
	json.NewDecoder(regResp.Body).Decode(&regRespBody)
	assert.Equal(t, "alice", regRespBody["login"])
	regResp.Body.Close()

	// 2. Логин
	loginBody := strings.NewReader(`{"login":"alice","password":"password123"}`)
	loginResp, err := client.Post(baseURL+"/login", "application/json", loginBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, loginResp.StatusCode)
	var loginRespBody map[string]string
	json.NewDecoder(loginResp.Body).Decode(&loginRespBody)
	token := loginRespBody["token"]
	assert.NotEmpty(t, token, "JWT token should be returned")
	loginResp.Body.Close()

	// 3. Создание первого поста
	postReq, _ := http.NewRequest(http.MethodPost, baseURL+"/posts", strings.NewReader(`{
        "title": "First post",
        "description": "Description",
        "price": 50,
        "image_url": "http://example.com/1.png"
    }`))
	postReq.Header.Set("Content-Type", "application/json")
	postReq.Header.Set("Authorization", "Bearer "+token)
	postResp, err := client.Do(postReq)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, postResp.StatusCode)
	var postRespBody post.Post
	json.NewDecoder(postResp.Body).Decode(&postRespBody)
	assert.Equal(t, "First post", postRespBody.Title)
	assert.Equal(t, 50.0, postRespBody.Price)
	assert.Equal(t, "alice", postRespBody.Owner)
	assert.NotZero(t, postRespBody.ID)
	postResp.Body.Close()

	// 4. Создание второго поста (для проверки фильтров)
	postReq2, _ := http.NewRequest(http.MethodPost, baseURL+"/posts", strings.NewReader(`{
        "title": "Second post",
        "description": "Another post",
        "price": 150,
        "image_url": "http://example.com/2.png"
    }`))
	postReq2.Header.Set("Content-Type", "application/json")
	postReq2.Header.Set("Authorization", "Bearer "+token)
	postResp2, err := client.Do(postReq2)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, postResp2.StatusCode)
	postResp2.Body.Close()

	// 5. Получение всех постов (без фильтра, без токена)
	feedResp, err := client.Get(baseURL + "/posts/feed")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, feedResp.StatusCode)
	var feedPosts []post.Post
	json.NewDecoder(feedResp.Body).Decode(&feedPosts)
	feedResp.Body.Close()
	// Ожидаем 2 поста в ответе
	assert.Len(t, feedPosts, 2)
	// Проверим, что поля соответствуют тому, что мы создали
	titles := map[string]bool{feedPosts[0].Title: true, feedPosts[1].Title: true}
	assert.True(t, titles["First post"])
	assert.True(t, titles["Second post"])

	// 6. Фильтрация по цене
	feedResp2, err := client.Get(baseURL + "/posts/feed?min_price=100")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, feedResp2.StatusCode)
	var filtered []post.Post
	json.NewDecoder(feedResp2.Body).Decode(&filtered)
	feedResp2.Body.Close()
	// Должен прийти только пост с ценой 150
	assert.Len(t, filtered, 1)
	assert.Equal(t, 150.0, filtered[0].Price)

	// 7. Фильтрация по владельцу
	feedResp3, err := client.Get(baseURL + "/posts/feed?owner=alice")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, feedResp3.StatusCode)
	var ownerFiltered []post.Post
	json.NewDecoder(feedResp3.Body).Decode(&ownerFiltered)
	feedResp3.Body.Close()
	// Ожидаем, что оба поста пришли
	assert.Len(t, ownerFiltered, 2)
	for _, p := range ownerFiltered {
		assert.Equal(t, "alice", p.Owner)
	}

	// 8. Проверка данных в БД
	// Пользователь
	var countUsers int
	_ = db.QueryRow(`SELECT COUNT(*) FROM users WHERE login=$1`, "alice").Scan(&countUsers)
	assert.Equal(t, 1, countUsers)
	var storedPwd string
	_ = db.QueryRow(`SELECT password FROM users WHERE login=$1`, "alice").Scan(&storedPwd)
	assert.NotEqual(t, "password123", storedPwd)

	// Посты
	var countPosts int
	_ = db.QueryRow(`SELECT COUNT(*) FROM posts WHERE owner=$1`, "alice").Scan(&countPosts)
	assert.Equal(t, 2, countPosts)
}
