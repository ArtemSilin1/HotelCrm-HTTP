package users

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ArtemSilin1/HotelCrm-HTTP/internal/storage/data"
	"github.com/golang-jwt/jwt/v5"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type Users struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	UserRole string `json:"userRole"`
}

// ============
//func (u *Users) Test(db *pgxpool.Pool) error {
//	u.Username = "master_admin"
//	u.Password = "123456"
//	u.UserRole = data.Admin_manager
//
//	hashedPassword, err := u.hashPassword()
//	if err != nil {
//		fmt.Println(err.Error())
//		return err
//	}
//
//	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
//	defer cancel()
//
//	q := "INSERT INTO Users(username, hash_password, user_role) VALUES ($1, $2, $3)"
//
//	_, err = db.Exec(ctx, q, u.Username, hashedPassword, u.UserRole)
//	if err != nil {
//		fmt.Println(err.Error())
//		return err
//	}
//
//	return err
//}
// ============

func (u *Users) hashPassword() (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashedPassword), nil
}

func (u *Users) checkUserExist(db *pgxpool.Pool) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	checkUserExistQ := "SELECT COUNT(*) FROM users WHERE id = $1"

	var count int
	if err := db.QueryRow(ctx, checkUserExistQ, u.Id).Scan(&count); err != nil {
		return false
	}

	return count > 0
}

func (u *Users) checkValidPassword(db *pgxpool.Pool) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	checkValidPasswordQ := "SELECT COUNT(*) FROM users WHERE hash_password = $1"

	var count int
	if err := db.QueryRow(ctx, checkValidPasswordQ, u.Password).Scan(&count); err != nil {
		return false
	}

	return count > 0
}

func (u *Users) CreateUser(db *pgxpool.Pool) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	hashedPassword, err := u.hashPassword()
	createUserQ := "INSERT INTO users (username, hash_password, user_role) VALUES ($1, $2, $3)"

	_, err = db.Exec(ctx, createUserQ, u.Username, hashedPassword, u.UserRole)
	if err != nil {
		return "", err
	}

	var userToken JWTToken
	token, err := userToken.GenerateToken(u)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (u *Users) LoginUser(db *pgxpool.Pool) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := "SELECT id, username, hash_password, user_role FROM users WHERE username = $1"

	var userFromDatabase Users

	if err := db.QueryRow(ctx, query, u.Username).Scan(
		&userFromDatabase.Id,
		&userFromDatabase.Username,
		&userFromDatabase.Password,
		&userFromDatabase.UserRole,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf(data.UserNotFound)
		}
		return "", err
	}

	// Сравниваем пароли
	if err := bcrypt.CompareHashAndPassword(
		[]byte(userFromDatabase.Password),
		[]byte(u.Password),
	); err != nil {
		return "", fmt.Errorf(data.WrongPassword)
	}

	var jwtManager JWTToken
	// Передаем адрес структуры в генератор токена
	token, err := jwtManager.GenerateToken(&userFromDatabase)
	if err != nil {
		return "", err
	}

	return token, nil
}

type JWTToken struct {
	Token  string `json:"token"`
	secret string `env:"SERVER_SECRET"`

	// Для ротации секретов
	currentSecret    []byte
	previousSecret   []byte
	lastRotationTime time.Time
	mutex            sync.RWMutex
}

// Чтение и хэширование секрета
func (j *JWTToken) ReadAndHashSecret() error {
	if j.secret == "" {
		if err := cleanenv.ReadEnv(j); err != nil {
			return fmt.Errorf("failed to read env: %w", err)
		}
	}

	// Хэшируем секрет
	hashedSecret := hashSecret(j.secret)

	j.mutex.Lock()
	defer j.mutex.Unlock()

	// Если это первый запуск или секрет изменился
	if j.currentSecret == nil || !compareSecrets(j.currentSecret, hashedSecret) {
		// Сохраняем предыдущий секрет для grace period
		if j.currentSecret != nil {
			j.previousSecret = j.currentSecret
		}

		j.currentSecret = hashedSecret
		j.lastRotationTime = time.Now()
	}

	return nil
}

// Хэширование секрета
func hashSecret(secret string) []byte {
	hash := sha256.Sum256([]byte(secret))
	return hash[:] // Преобразуем [32]byte в []byte
}

// Сравнение секретов
func compareSecrets(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Проверка необходимости ротации (раз в день)
func (j *JWTToken) checkAndRotate() error {
	j.mutex.RLock()
	needsRotation := time.Since(j.lastRotationTime) >= 24*time.Hour
	j.mutex.RUnlock()

	if needsRotation {
		return j.ReadAndHashSecret()
	}
	return nil
}

// Получение текущего секрета с ротацией
func (j *JWTToken) getCurrentSecret() ([]byte, error) {
	// Проверяем, нужно ли обновить секрет
	if err := j.checkAndRotate(); err != nil {
		return nil, err
	}

	j.mutex.RLock()
	defer j.mutex.RUnlock()

	return j.currentSecret, nil
}

// Получение всех активных секретов (текущий + предыдущий для grace period)
func (j *JWTToken) getAllActiveSecrets() ([][]byte, error) {
	if err := j.checkAndRotate(); err != nil {
		return nil, err
	}

	j.mutex.RLock()
	defer j.mutex.RUnlock()

	secrets := [][]byte{j.currentSecret}
	if j.previousSecret != nil {
		secrets = append(secrets, j.previousSecret)
	}

	return secrets, nil
}

// Генерация токена с временем жизни 3 суток
func (j *JWTToken) GenerateToken(u *Users) (string, error) {
	// Читаем и хэшируем секрет при необходимости
	if err := j.ReadAndHashSecret(); err != nil {
		return "", fmt.Errorf("failed to read secret: %w", err)
	}

	secret, err := j.getCurrentSecret()
	if err != nil {
		return "", fmt.Errorf("failed to get secret: %w", err)
	}

	// Время жизни токена - 3 суток
	expirationTime := time.Now().Add(24 * time.Hour)

	claims := jwt.MapClaims{
		"uid":        u.Id,
		"username":   u.Username,
		"role":       u.UserRole,
		"exp":        expirationTime.Unix(),
		"iat":        time.Now().Unix(),
		"secret_gen": j.lastRotationTime.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("can't create token: %w", err)
	}

	j.Token = signedToken

	return signedToken, nil
}

// Верификация токена с поддержкой старого секрета (grace period)
func (j *JWTToken) VerifyToken(tokenString string) (jwt.MapClaims, error) {
	secrets, err := j.getAllActiveSecrets()
	if err != nil {
		return nil, fmt.Errorf("failed to get secrets: %w", err)
	}

	var lastErr error

	// Пробуем верифицировать токен всеми активными секретами
	for _, secret := range secrets {
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Проверяем алгоритм
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return secret, nil
		})

		if err != nil {
			lastErr = err
			continue
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			return claims, nil
		}
	}

	return nil, fmt.Errorf("token verification failed: %w", lastErr)
}

// Проверка валидности токена (для middleware)
func (j *JWTToken) IsValid(tokenString string) bool {
	_, err := j.VerifyToken(tokenString)
	return err == nil
}

// Получение информации о пользователе из токена
func (j *JWTToken) GetUserFromToken(tokenString string) (*Users, error) {
	claims, err := j.VerifyToken(tokenString)
	if err != nil {
		return nil, err
	}

	user := &Users{
		Username: claims["username"].(string),
		UserRole: claims["user_role"].(string),
		Id:       claims["id"].(int),
	}

	return user, nil
}
