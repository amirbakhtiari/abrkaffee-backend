package bootstrap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/amirbakhtiari/abrcoffee/config"
	"github.com/amirbakhtiari/abrcoffee/internal/adapters/repository/pg_auth/db_models"
	"github.com/amirbakhtiari/abrcoffee/internal/adapters/repository/pg_user"
	"github.com/amirbakhtiari/abrcoffee/pkg/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupIntegrationTest(t *testing.T) (*gorm.DB, *gin.Engine) {
	t.Helper()

	cfg := &config.Config{
		Server: config.ServerConfig{
			Port: "8080",
			Env:  "test",
		},
		Database: config.DatabaseConfig{
			Host:     "127.0.0.1",
			User:     "abrcoffee_test",
			Password: "test_only_password",
			Name:     "abrcoffee_test",
			Port:     5434,
			SSLMode:  "disable",
		},
		JWT: config.JWTConfig{
			Secret:      "integration-test-secret",
			ExpireHours: 24,
		},
	}

	// اتصال فقط به دیتابیس تست
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		cfg.Database.Host,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.Port,
		cfg.Database.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("connect to test database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql.DB: %v", err)
	}

	if err := sqlDB.Ping(); err != nil {
		t.Fatalf("ping test database: %v", err)
	}

	t.Cleanup(func() {
		if err := sqlDB.Close(); err != nil {
			t.Errorf("close test database: %v", err)
		}
	})

	// فقط مدل‌های موردنیاز Auth و User
	if err := db.AutoMigrate(
		&pg_user.UserModel{},
		&db_models.AccountDB{},
	); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}

	// پاک‌سازی دیتابیس تست قبل از هر تست
	if err := db.Exec(
		"TRUNCATE TABLE accounts, users RESTART IDENTITY CASCADE",
	).Error; err != nil {
		t.Fatalf("truncate test tables: %v", err)
	}

	// ساخت پروفایل کاربر
	profile := pg_user.UserModel{
		Name:     "Integration",
		Family:   "Admin",
		Email:    "integration.admin@example.com",
		Mobile:   "09123456789",
		IsActive: true,
	}

	if err := db.Create(&profile).Error; err != nil {
		t.Fatalf("create test user: %v", err)
	}

	// ساخت Hash واقعی رمز عبور
	passwordHash, err := utils.NewPasswordHash("Admin@123456")
	if err != nil {
		t.Fatalf("hash test password: %v", err)
	}

	// ساخت حساب Admin
	account := db_models.AccountDB{
		UserID:       profile.ID,
		Username:     "integration_admin",
		PasswordHash: passwordHash,
		Role:         "Admin",
		IsActive:     true,
	}

	if err := db.Create(&account).Error; err != nil {
		t.Fatalf("create test account: %v", err)
	}

	gin.SetMode(gin.TestMode)

	router := NewRouter(db, cfg)

	return db, router
}

func TestIntegration_AuthLogin_Success(t *testing.T) {
	_, router := setupIntegrationTest(t)

	requestBody := map[string]string{
		"username": "integration_admin",
		"password": "Admin@123456",
	}

	body, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/auth/login",
		bytes.NewReader(body),
	)

	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d. Response: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	var response struct {
		Message string `json:"message"`
		Token   string `json:"token"`
	}

	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response.Token == "" {
		t.Fatal("expected JWT token, got empty token")
	}

	claims, err := utils.ValidateToken(
		response.Token,
		"integration-test-secret",
	)
	if err != nil {
		t.Fatalf("validate generated JWT: %v", err)
	}

	if claims.Username != "integration_admin" {
		t.Errorf(
			"expected username integration_admin, got %s",
			claims.Username,
		)
	}

	if claims.Role != "Admin" {
		t.Errorf("expected role Admin, got %s", claims.Role)
	}

	if claims.UserID == 0 {
		t.Error("expected non-zero user ID")
	}
}
