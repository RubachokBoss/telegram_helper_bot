package service

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/RubachokBoss/telegram_helper_bot/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type authService struct {
	userRepo  domain.WebUserRepository
	jwtSecret string
}

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func NewAuthService(userRepo domain.WebUserRepository, jwtSecret string) domain.AuthService {
	return &authService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

func (s *authService) Register(reg domain.UserRegistration) (*domain.WebUser, error) {
	log.Printf("🔐 Attempting to register user: %s", reg.Email)

	// Проверяем, что пользователь с таким email не существует
	existingUser, err := s.userRepo.FindByEmail(reg.Email)
	if err != nil {
		log.Printf("❌ Error checking existing user: %v", err)
		return nil, err
	}
	if existingUser != nil {
		log.Printf("⚠️ User with email %s already exists", reg.Email)
		return nil, errors.New("user with this email already exists")
	}

	// Валидация данных
	if reg.Email == "" || reg.Password == "" || reg.FirstName == "" {
		return nil, errors.New("email, password, and first name are required")
	}

	if len(reg.Password) < 6 {
		return nil, errors.New("password must be at least 6 characters long")
	}

	// Хэшируем пароль
	hashedPassword, err := s.HashPassword(reg.Password)
	if err != nil {
		log.Printf("❌ Error hashing password: %v", err)
		return nil, err
	}

	// Создаем пользователя
	user := &domain.WebUser{
		Email:     reg.Email,
		Password:  hashedPassword,
		FirstName: reg.FirstName,
		LastName:  reg.LastName,
	}

	err = s.userRepo.Create(user)
	if err != nil {
		log.Printf("❌ Error creating user: %v", err)
		return nil, err
	}

	log.Printf("✅ User registered successfully: %s (%s)", user.ID, user.Email)
	return user, nil
}

func (s *authService) Login(creds domain.UserCredentials) (string, error) {
	log.Printf("🔐 Attempting login for user: %s", creds.Email)

	// Находим пользователя по email
	user, err := s.userRepo.FindByEmail(creds.Email)
	if err != nil {
		log.Printf("❌ Error finding user: %v", err)
		return "", err
	}
	if user == nil {
		log.Printf("⚠️ User not found: %s", creds.Email)
		return "", errors.New("invalid credentials")
	}

	// Проверяем пароль
	if !s.CheckPasswordHash(creds.Password, user.Password) {
		log.Printf("⚠️ Invalid password for user: %s", creds.Email)
		return "", errors.New("invalid credentials")
	}

	// Генерируем JWT токен
	token, err := s.generateToken(user)
	if err != nil {
		log.Printf("❌ Error generating token: %v", err)
		return "", err
	}

	log.Printf("✅ User logged in successfully: %s", user.Email)
	return token, nil
}

func (s *authService) ValidateToken(tokenString string) (*domain.WebUser, error) {
	log.Printf("🔐 Validating JWT token")

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		log.Printf("❌ Error parsing token: %v", err)
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		// Находим пользователя по ID из токена
		user, err := s.userRepo.FindByID(claims.UserID)
		if err != nil {
			log.Printf("❌ Error finding user by ID: %v", err)
			return nil, err
		}
		if user == nil {
			log.Printf("⚠️ User not found by ID: %s", claims.UserID)
			return nil, errors.New("user not found")
		}

		log.Printf("✅ Token validated for user: %s", user.Email)
		return user, nil
	}

	log.Printf("⚠️ Invalid token")
	return nil, errors.New("invalid token")
}

func (s *authService) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func (s *authService) CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (s *authService) generateToken(user *domain.WebUser) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour) // Токен действителен 24 часа

	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))

	if err != nil {
		return "", err
	}

	return tokenString, nil
}
