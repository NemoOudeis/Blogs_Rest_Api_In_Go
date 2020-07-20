package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"cloud.google.com/go/firestore"
	"firebase.google.com/go/auth"
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

func (users *Users) signup(response http.ResponseWriter, request *http.Request) {
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
			CustomMessage: "error creating a user",
		}
		ExitWithError(response, statusCode, statusMessage)
		return
	}

	newUserInfo := User{
		GeneratedID: newUser.UID,
		Email:       strings.Join(email, ""),
		Password:    strings.Join(password, ""),
	}

	result, _, err := users.db.Collection("users").Add(context.Background(), map[string]interface{}{
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

	docSnapshot, err := users.db.Collection("users").Doc(result.ID).Get(context.Background())
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

func (users *Users) login(response http.ResponseWriter, request *http.Request) {
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
	if userInfoFromDB["password"] != password[0] {
		statusCode := http.StatusBadRequest
		statusMessage := Error{
			Message:       http.StatusText(statusCode),
			CustomMessage: "Login failed. Make sure both of your email and password is correct.",
		}
		ExitWithError(response, statusCode, statusMessage)
		return
	}

	token, err := users.authClient.CustomToken(context.Background(), userInfoFromDB["id"].(string))
	if err != nil {
		statusCode := http.StatusServiceUnavailable
		statusMessage := Error{
			// err.Error() is a custom error message from client firestore API
			Message: err.Error(),
		}
		ExitWithError(response, statusCode, statusMessage)
		return
	}

	statusCode := http.StatusOK
	statusMessage := SuccessJSONGenerator(http.StatusText(statusCode), token)
	ReturnSuccessfulResponse(response, statusCode, statusMessage)
}
