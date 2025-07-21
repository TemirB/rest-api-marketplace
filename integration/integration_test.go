package integration

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"

	"github.com/TemirB/rest-api-marketplace/internal/auth"
	"github.com/TemirB/rest-api-marketplace/internal/database"
	"github.com/TemirB/rest-api-marketplace/internal/middleware"
	"github.com/TemirB/rest-api-marketplace/internal/post"
	"github.com/TemirB/rest-api-marketplace/pkg/jwt"
)

var (
	testServer *httptest.Server
	db         *sql.DB
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	// 1) Старт контейнера Postgres
	pgC, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithExposedPorts("5432"),
		testcontainers.WithWaitStrategy(wait.ForListeningPort("5432/tcp")),
		postgres.WithInitScripts("../migrations/init.sql"),
	)
	if err != nil {
		panic("failed to start postgres container: " + err.Error())
	}
	defer func() { _ = pgC.Terminate(ctx) }()

	// 2) Подключение к БД
	dsn, err := pgC.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		panic(err)
	}
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}
	// даем БД немного прогреться
	time.Sleep(500 * time.Millisecond)

	// 3) Инициализация приложения
	repo := &database.Repository{DB: db, Logger: zap.NewNop()}
	userStore := auth.NewStorage(repo, zap.NewNop())
	postStore := post.NewStorage(repo, zap.NewNop())
	tokenManager := jwt.New("testsecret", time.Hour)
	authService := auth.NewService(userStore, tokenManager, zap.NewNop())
	postService := post.NewService(postStore, zap.NewNop())
	authHandler := auth.NewHandler(authService, zap.NewNop())
	postHandler := post.NewHandler(postService, zap.NewNop())

	mux := http.NewServeMux()
	mux.HandleFunc("/register", authHandler.Register)
	mux.HandleFunc("/login", authHandler.Login)

	protected := middleware.JWTAuthMiddleware(authService)
	mux.Handle("/posts", protected(http.HandlerFunc(postHandler.CreatePost)))
	mux.HandleFunc("/posts/feed", postHandler.GetPosts)

	testServer = httptest.NewServer(mux)
	code := m.Run()
	testServer.Close()
	os.Exit(code)
}

func TestFullScenario(t *testing.T) {
	client := testServer.Client()
	base := testServer.URL

	rnd := time.Now().UnixNano() % 1_000_000_000
	login := fmt.Sprintf("alice%d", rnd)

	// 1. Регистрация
	regBody := fmt.Sprintf(`{"login":"%s","password":"Password123!"}`, login)
	resp, err := client.Post(base+"/register", "application/json", strings.NewReader(regBody))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var reg map[string]string
	json.NewDecoder(resp.Body).Decode(&reg)
	resp.Body.Close()
	assert.Equal(t, login, reg["login"])

	// 2. Логин
	loginBody := fmt.Sprintf(`{"login":"%s","password":"Password123!"}`, login)
	resp, err = client.Post(base+"/login", "application/json", strings.NewReader(loginBody))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var lg map[string]string
	json.NewDecoder(resp.Body).Decode(&lg)
	resp.Body.Close()
	token := lg["token"]
	assert.NotEmpty(t, token)

	// 3. Создание первого поста
	post1 := `{"title":"First post","description":"Description","price":50,"image_url":"http://example.com/1.png"}`
	req1, _ := http.NewRequest(http.MethodPost, base+"/posts", strings.NewReader(post1))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("Authorization", "Bearer "+token)
	resp, err = client.Do(req1)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var p1 post.Post
	json.NewDecoder(resp.Body).Decode(&p1)
	resp.Body.Close()
	assert.Equal(t, "First post", p1.Title)
	assert.Equal(t, 50.0, p1.Price)
	assert.Equal(t, login, p1.Owner)
	assert.NotZero(t, p1.ID)

	// 4. Попытка без токена → 401
	reqNoAuth, _ := http.NewRequest(http.MethodPost, base+"/posts", strings.NewReader(post1))
	reqNoAuth.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(reqNoAuth)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()

	// 5. Создание второго поста
	post2 := `{"title":"Second post","description":"Another","price":150,"image_url":"http://example.com/2.png"}`
	req2, _ := http.NewRequest(http.MethodPost, base+"/posts", strings.NewReader(post2))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+token)
	resp, err = client.Do(req2)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	// 6. Лента без фильтра
	resp, err = client.Get(base + "/posts/feed")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var feed []post.Post
	json.NewDecoder(resp.Body).Decode(&feed)
	resp.Body.Close()
	assert.Len(t, feed, 2)
	found := map[string]bool{feed[0].Title: true, feed[1].Title: true}
	assert.True(t, found["First post"])
	assert.True(t, found["Second post"])

	// 7. Фильтрация по цене
	resp, err = client.Get(base + "/posts/feed?min_price=100")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var priceF []post.Post
	json.NewDecoder(resp.Body).Decode(&priceF)
	resp.Body.Close()
	assert.Len(t, priceF, 1)
	assert.Equal(t, 150.0, priceF[0].Price)

	// 8. Фильтрация по владельцу
	resp, err = client.Get(base + "/posts/feed?owner=" + login)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var ownerF []post.Post
	json.NewDecoder(resp.Body).Decode(&ownerF)
	resp.Body.Close()
	assert.Len(t, ownerF, 2)
	for _, p := range ownerF {
		assert.Equal(t, login, p.Owner)
	}

	// 9. Проверка в БД
	var cntUsers int
	err = db.QueryRow(`SELECT COUNT(*) FROM users WHERE login=$1`, login).Scan(&cntUsers)
	assert.NoError(t, err)
	assert.Equal(t, 1, cntUsers)

	var cntPosts int
	err = db.QueryRow(`SELECT COUNT(*) FROM posts WHERE owner=$1`, login).Scan(&cntPosts)
	assert.NoError(t, err)
	assert.Equal(t, 2, cntPosts)
}
