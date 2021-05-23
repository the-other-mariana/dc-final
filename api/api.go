package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strings"
	"encoding/base64"
	"net/http"
	"time"
	"path/filepath"
	"path"
	"strconv"

	"github.com/the-other-mariana/dc-final/controller"
	"github.com/the-other-mariana/dc-final/scheduler"
)

// user object (struct)
type User struct {
	username string
	password string
	token string
}

// runtime database where all users are stored by token
var Users = make(map[string]User)


// function to create a unique token that expires in 3 hours
func CreateToken(user string) (string, error) {
	tokenizer := jwt.New(jwt.SigningMethodHS256)
	claims := tokenizer.Claims.(jwt.MapClaims)
	claims["user"] = user
	claims["exp"] = time.Now().Add(time.Hour * 3).Unix()
	t, err := tokenizer.SignedString([]byte("our-secret"))
	claims["token"] = t
	return t, err
}

// function that converts file size (int64) into kb, mb, gb
func ByteCount(b int64) string {
    const unit = 1000
	var prefix string = "kMGTPE"
    if b < unit {
        return fmt.Sprintf("%d b", b)
    }
    div, exp := int64(unit), 0
    for n := b / unit; n >= unit; n /= unit {
        div *= unit
        exp++
    }
    return fmt.Sprintf("%.1f%cb", float64(b)/float64(div), prefix[exp])
}

// generic error function
func ErrorResponse(msg string) (gin.H) {
	resp := gin.H{
		"status": "error",
		"message": msg,
	}
	return resp
}

// login json response
func LoginResponse(username string, token string) (gin.H) {
	msg := fmt.Sprintf("Hi %v, welcome to the DPIP system", username)
	resp := gin.H{
		"message": msg,
		"token": token,
	}
	return resp
}

// login handler
func Login(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	fmt.Println("Response Type:", c.Writer.Header().Get("Content-Type"))
	// get curl params
	params := strings.Split(c.Request.Header.Get("Authorization"), " ")

	// they get hashed, so lets decode them to get user:password string
	auth, _ := base64.StdEncoding.DecodeString(params[1])
	fmt.Printf("Decoded User: %v\n", string(auth))

	userInfo := strings.Split(string(auth), ":")
	exists := false

	for _,u := range Users {
		if u.username == userInfo[0] {
			exists = true
		}
	}

	if !exists {
		newToken, err := CreateToken(userInfo[0] + "." + userInfo[1])
		if err != nil {
			c.JSON(http.StatusConflict, ErrorResponse("Error at token creation"))
			return
		}
		newUser := User{
			username: userInfo[0],
			password: userInfo[1],
			token: newToken,
		}
		Users[newToken] = newUser
		c.JSON(http.StatusOK, LoginResponse(newUser.username, newUser.token))
	}
	if exists {
		c.JSON(http.StatusOK, ErrorResponse("This user is already logged"))
	}
}

// logout json response
func LogoutResponse(username string) (gin.H) {
	msg := fmt.Sprintf("Bye %v, your token has been revoked", username)
	resp := gin.H{
		"message": msg,
	}
	return resp
}

// logout handler
func Logout(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	fmt.Println("Response Type:", c.Writer.Header().Get("Content-Type"))

	params := strings.Split(c.Request.Header.Get("Authorization"), " ")
	token := params[1]

	if _, ok := Users[token]; ok {
		c.JSON(http.StatusOK, LogoutResponse(Users[token].username))
		delete(Users, token)
	} else {
		c.JSON(http.StatusConflict, ErrorResponse("Your token does not exist yet"))
	}
}

// status json response
func StatusResponse(username string) (gin.H) {
	msg := fmt.Sprintf("Hi %v, the DPIP System is Up and Running", username)
	t := time.Now() // in format 2021-04-14T23:20:47.361719701Z
	resp := gin.H{
		"message": msg,
		"time": t.Format("2006-01-02 15:04:05"),
	}
	return resp
}

// status handler
func Status(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	fmt.Println("Response Type:", c.Writer.Header().Get("Content-Type"))

	params := strings.Split(c.Request.Header.Get("Authorization"), " ")
	token := params[1]

	if _, ok := Users[token]; ok {
		c.JSON(http.StatusOK, StatusResponse(Users[token].username))
	} else {
		c.JSON(http.StatusConflict, ErrorResponse("Your token does not exist yet"))
	}
}

// upload json response
func UploadResponse(filename string, size int64) (gin.H) {
	resp := gin.H{
		"message": "An image has been successfully uploaded",
		"filename": filename,
		"size": ByteCount(size),
	}
	return resp
}

// upload handler
func Upload(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "application/json")
	fmt.Println("Response Type:", c.Writer.Header().Get("Content-Type"))

	params := strings.Split(c.Request.Header.Get("Authorization"), " ")
	token := params[1]

	if _, ok := Users[token]; ok {
		file, err := c.FormFile("data")
		if err != nil {
			c.JSON(http.StatusConflict, ErrorResponse("Error at retrieving form data"))
			return
		}
		
		imgSize := file.Size
		imgName := filepath.Base(file.Filename)
		// gives the uploaded image a unique name
		id := primitive.NewObjectID().Hex()
		newFilename := id + "_" + imgName
		newPath := path.Join("./", newFilename)

		if err := c.SaveUploadedFile(file, newPath); err != nil {
			c.JSON(http.StatusConflict, ErrorResponse("Could not save the file"))
			return
		}

		c.JSON(http.StatusOK, UploadResponse(newFilename, imgSize))
	} else {
		c.JSON(http.StatusConflict, ErrorResponse("Your token does not exist yet"))
	}
}

var Jobs = make(chan scheduler.Job)
var NumTests int

func Workloads(c *gin.Context) {
	params := strings.Split(c.Request.Header.Get("Authorization"), " ")
	token := params[1]

	if _, ok := Users[token]; ok {
		sampleJob := scheduler.Job{Address: "localhost:50051", RPCName: "test"}
		Jobs <- sampleJob
		time.Sleep(time.Second * 5)
		name := controller.GetWorker(NumTests)
		c.JSON(http.StatusOK, map[string]interface{}{
			"Workload": "test",
			"Job ID":   NumTests,
			"Status":   "Scheduling",
			"Result":   "Done by " + name,
		})
		NumTests += 1
	} else {
		c.JSON(http.StatusOK, ErrorResponse("Your token does not exist yet"))
	}
}

func WorkerStatus(c *gin.Context) {
	params := strings.Split(c.Request.Header.Get("Authorization"), " ")
	token := params[1]

	workerName := c.Param("worker")
	if _, ok := Users[token]; ok {
		c.JSON(http.StatusOK, map[string]interface{}{
			"Worker": controller.Workers[workerName].Name,
			"Tags":   controller.Workers[workerName].Tags,
			"Status": controller.Workers[workerName].Status,
			"Usage":  strconv.Itoa(controller.Workers[workerName].Usage) + "%",
		})
	} else {
		c.JSON(http.StatusOK, ErrorResponse("Your token does not exist yet"))
	}
}