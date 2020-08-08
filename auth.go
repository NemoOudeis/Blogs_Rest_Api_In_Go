package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"firebase.google.com/go/auth"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

// Users is a structure which holds database for CRUD operation in the client and app initialized for admin in the backend
type Users struct {
	db         *firestore.Client
	authClient *auth.Client
}

// User holds basic user info of a current user
type User struct {
	GeneratedID string `json:"id"`
	Email       string `json:"email"`
	Password    string `json:"password"`
}

func initUsers(db *firestore.Client, authClient *auth.Client) *Users {
	return &Users{db: db, authClient: authClient}
}

func createTokenForAuth(email string) (string, error) {
	jwtHashKey := env.JwtHashKey
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_email": email,
		"iss":        "__init__",
		"exp":        time.Now().Add(time.Minute * 60).Unix(),
	})
	tokenString, err := token.SignedString([]byte(jwtHashKey))
	log.Println(tokenString) // <--- security problem
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// Signup registers a new user with given valid email and password
func (users *Users) Signup(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")

	if request.Method != http.MethodPost {
		statusCode := http.StatusMethodNotAllowed
		statusMessage := Error{
			Message: http.StatusText(statusCode),
		}
		ExitWithError(response, statusCode, statusMessage)
		return
	}

	request.ParseForm()
	urlEncodedFormInputMap := request.Form
	email, isEmailFound := urlEncodedFormInputMap["email"]
	password, isPasswordFound := urlEncodedFormInputMap["password"]

	if isEmailFound == false {
		statusCode := http.StatusBadRequest
		statusMessage := Error{
			Message:       http.StatusText(statusCode),
			CustomMessage: "Email is required.",
		}
		ExitWithError(response, statusCode, statusMessage)
		return
	}

	if isPasswordFound == false {
		statusCode := http.StatusBadRequest
		statusMessage := Error{
			Message:       http.StatusText(statusCode),
			CustomMessage: "Password is required.",
		}
		ExitWithError(response, statusCode, statusMessage)
		return
	}

	params := (&auth.UserToCreate{}).
		Email(strings.Join(email, "")).
		Password(strings.Join(password, "")).
		Disabled(false)

	newUser, err := users.authClient.CreateUser(context.Background(), params)
	if err != nil {
		statusCode := http.StatusServiceUnavailable
		statusMessage := Error{
			// err.Error() is a custom error message from client firestore API
			Message:       err.Error(),
			CustomMessage: "Error creating a user.",
		}
		ExitWithError(response, statusCode, statusMessage)
		return
	}
	log.Printf("Successfully created user: %#v\n", newUser.UserInfo)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password[0]), 10)
	if err != nil {
		statusCode := http.StatusServiceUnavailable
		statusMessage := Error{
			Message:       err.Error(),
			CustomMessage: "Error minting a token.",
		}
		ExitWithError(response, statusCode, statusMessage)
		return
	}
	log.Println(password[0])
	log.Println(hashedPassword)
	newUserInfo := User{
		GeneratedID: newUser.UID,
		Email:       email[0],
		Password:    string(hashedPassword),
	}
	log.Println(newUserInfo)
	docRef, _, err := users.db.Collection("users").Add(context.Background(), map[string]interface{}{
		"id":       newUserInfo.GeneratedID,
		"email":    newUserInfo.Email,
		"password": newUserInfo.Password,
	})

	if err != nil {
		statusCode := http.StatusServiceUnavailable
		statusMessage := Error{
			// err.Error() is a custom error message from client firestore API
			Message: err.Error(),
		}
		ExitWithError(response, statusCode, statusMessage)
		return
	}

	docSnapshot, err := users.db.Collection("users").Doc(docRef.ID).Get(context.Background())
	if err != nil {
		statusCode := http.StatusServiceUnavailable
		statusMessage := Error{
			// err.Error() is a custom error message from client firestore API
			Message: err.Error(),
		}
		ExitWithError(response, statusCode, statusMessage)
		return
	}

	docSnapshotDatum := docSnapshot.Data()
	userEmailFromDB := docSnapshotDatum["email"].(string)
	customMessage := fmt.Sprintf("New user was created with this email: %s", userEmailFromDB)

	statusCode := http.StatusCreated
	statusMessage := SuccessJSONGenerator(http.StatusText(statusCode), customMessage)
	ReturnSuccessfulResponse(response, statusCode, statusMessage)
}

/*
 Responsibilities of this function (by reading the code):
 - parse input 	 											<- adapter layer
 - input validation												<- business logic layer
 - lookup user													<- business logic layer
 - query db & handle error cases							<- adapter layer
 - authenticate (check password)								<- business logic
 - create JWT token											<- adapter layer
 - return error as HTTP response (token, token missing)		<- adapter layer
*/
// Login authenticates existing user and mints token to allow exploring other endpoints
func (users *Users) Login(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")

	if request.Method != http.MethodPost {
		statusCode := http.StatusMethodNotAllowed
		statusMessage := Error{
			Message: http.StatusText(statusCode),
		}
		ExitWithError(response, statusCode, statusMessage)
		return
	}

	request.ParseForm()
	urlEncodedFormInputMap := request.Form
	email, isEmailFound := urlEncodedFormInputMap["email"]
	password, isPasswordFound := urlEncodedFormInputMap["password"]

	if isEmailFound == false {
		statusCode := http.StatusBadRequest
		statusMessage := Error{
			Message:       http.StatusText(statusCode),
			CustomMessage: "Email is required.",
		}
		ExitWithError(response, statusCode, statusMessage)
		return
	}

	if isPasswordFound == false {
		statusCode := http.StatusBadRequest
		statusMessage := Error{
			Message:       http.StatusText(statusCode),
			CustomMessage: "Password is required.",
		}
		ExitWithError(response, statusCode, statusMessage)
		return
	}

	// inject dependencies
	func initLoginUseCase(db *firestore.Client) *LoginUseCase { // <-- usually happens at application startup
		return &LoginUseCase{ persistence: &UserPersistenceWithFirebase{db: db} }
	}

	// 2 other objects
	// 1. business logic (struct type, which only depends on persistence obj. Conforms to the obj `UserlookUpper`.)
	// 2. persistence object

	type UserPersistence interface { // <-- like input port or output port
		LookupByEmail(email string) user
	}

	type UserPersistenceWithFirebase struct { // <--- presenter
		db *firestore.Client
	}

	// interactor (domain)
	type LoginUseCase struct {
		persistence *UserPersistence
	}

	func (useCase *LoginUseCase) loginUser(email, password) user, error {
		user, err := useCase.persistence.LookupByEmail(email)
		if err != nil {
			return // return new business domain error UserNotFoundException
		}

		// we check if password is correct

		if err != nil {
			return // return new business domain error InvalidPassword
		}
	}

	token, err := useCase.LoginUser(email, passowrd)
	iter := users.db.Collection("users").Where("email", "==", strings.Join(email, "")).Limit(1).Documents(context.Background())
	doc, err := iter.GetAll()
	if err != nil {
		statusCode := http.StatusServiceUnavailable
		statusMessage := Error{
			// err.Error() is a custom error message from client firestore API
			Message: err.Error(),
		}
		ExitWithError(response, statusCode, statusMessage)
		return
	}

	if len(doc) == 0 {
		statusCode := http.StatusBadRequest
		statusMessage := Error{
			Message:       http.StatusText(statusCode),
			CustomMessage: "Login failed. The user does not exist.",
		}
		ExitWithError(response, statusCode, statusMessage)
		return
	}

	userInfoFromDB := doc[0].Data()
	hashedPassword := userInfoFromDB["password"]
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword.(string)), []byte(password[0]))
	log.Println(hashedPassword) 	// <--- security problem	
	log.Println(password[0])		// <--- security problem
	if err != nil {
		statusCode := http.StatusUnauthorized
		statusMessage := Error{
			Message:       http.StatusText(statusCode),
			CustomMessage: "Login failed. Make sure both of your email and password is correct.",
		}
		ExitWithError(response, statusCode, statusMessage)
		return
	}

	token, err := createTokenForAuth(email[0])
	if err != nil {
		statusCode := http.StatusServiceUnavailable
		statusMessage := Error{
			Message:       http.StatusText(statusCode),
			CustomMessage: "Failed to mint a token",
		}
		ExitWithError(response, statusCode, statusMessage)
		return
	}

	statusCode := http.StatusOK
	statusMessage := SuccessJSONGenerator(http.StatusText(statusCode), token)
	ReturnSuccessfulResponse(response, statusCode, statusMessage)
}

/*
 Responsibilities of this function (by reading the code):
 - input validation											<- adapter layer
 - validate jwt signature 										<- business logic layer
 - read environment file									<- adapter layer
 - return error as HTTP response (token, token missing)		<- adapter layer
 - invoke next handler										<- adapter layer
*/
func (users *Users) verifyToken(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Content-Type", "application/json")
		authHeader := request.Header.Get("Authorization")
		bearerToken := strings.Split(authHeader, " ")

		if len(bearerToken) == 2 && bearerToken[0] == "Bearer" {
			tokenFromClient := bearerToken[1]
			token, err := jwt.Parse(tokenFromClient, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("Error occured")
				}
				jwtHashKey := LoadEnvFileAndReturnEnvVarValueByKey("JWT_HASH_KEY")
				return []byte(jwtHashKey), nil
			})

			if err != nil || !token.Valid {
				statusMessage := Error{
					Message:       http.StatusText(http.StatusUnauthorized),
					CustomMessage: "Auth Failed.",
				}
				ExitWithError(response, statusCode, statusMessage)
				return
			}

			next.ServeHTTP(response, request)
		} else {
			statusMessage := Error{
				Message:       http.StatusText(http.StatusBadRequest),
				CustomMessage: "Invalid Token.",
			}
			ExitWithError(response, statusCode, statusMessage)
			return
		}
	})
}
