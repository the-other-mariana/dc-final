package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/static"
	"github.com/dgrijalva/jwt-go"
	//"go.mongodb.org/mongo-driver/bson/primitive"
	"strings"
	"encoding/base64"
	"net/http"
	"time"
	"path/filepath"
	"path"
	"os"
	//"io"

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
	resp := gin.H{
		"user": username,
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
		"logout_message": msg,
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
func StatusResponse() (gin.H) {
	t := time.Now() // in format 2021-04-14T23:20:47.361719701Z

	resp := gin.H{
		"system_name": "Distributed Parallel Image Processing (DPIP) System",
		"server_time": t.Format("2006-01-02 15:04:05"),
		"active_workloads": len(controller.Workloads),
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
		c.JSON(http.StatusOK, StatusResponse())
	} else {
		c.JSON(http.StatusConflict, ErrorResponse("Your token does not exist yet"))
	}
}

// upload json response
func UploadResponse(workloadId string, id string, imgType string) (gin.H) {
	resp := gin.H{
		"workload_id": workloadId,
		"image_id": id,
		"type": imgType,
	}
	return resp
}

var Jobs = make(chan scheduler.Job)
var NumTests int

func CreateWorkload(c *gin.Context){
	params := strings.Split(c.Request.Header.Get("Authorization"), " ")
	token := params[1]

	if _, ok := Users[token]; ok {
		workloadName := c.PostForm("workload_name")
		filter := c.PostForm("filter")

		if strings.Contains(workloadName, "_"){
			c.JSON(http.StatusConflict, ErrorResponse("Workload Names cannot include _ character"))
		}
		if strings.Contains(workloadName, "="){
			c.JSON(http.StatusConflict, ErrorResponse("Workload Names cannot include = character"))
		}

		taken := false
		for _, v := range controller.Workloads {
			if v.Name == workloadName {
				taken = true
				break
			}
		}
		
		if (!taken){
			workloadStatus := "scheduling"
			if len(controller.Workers) > 0 {
				workloadStatus = "running"
			}
			// making directory for processed images
			uploadsFolder := "public/results/" + workloadName + "/"
			_ = os.MkdirAll(uploadsFolder, 0755)

			// making directory for not yet processed images
			downloadFolder := "download/" + workloadName + "/"
			_ = os.MkdirAll("public/" + downloadFolder, 0755)

			newWL := controller.Workload{
				Id: fmt.Sprintf("%v", len(controller.Workloads)),
				Filter: filter,
				Name: workloadName,
				Status: workloadStatus,
				Jobs: 0,
				Imgs: []string{},
				Filtered: []string{},
			}

			controller.Workloads[fmt.Sprintf("%v", newWL.Id)] = newWL

			c.JSON(http.StatusOK, map[string]interface{}{
				"workload_id": newWL.Id,
				"filter":   filter,
				"workload_name": workloadName,
				"status": newWL.Status,
				"running_jobs": newWL.Jobs,
				"filtered_images": newWL.Filtered,
			})
		} else {
			c.JSON(http.StatusOK, ErrorResponse("This workload already exists"))
		}

		
	} else {
		c.JSON(http.StatusOK, ErrorResponse("Your token does not exist yet"))
	}
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

		workloadId := c.PostForm("workload_id")
		
		id := 0
		myWorkload := controller.Workload{}
		updatedWL := controller.Workload{}

		if _, ok := controller.Workloads[workloadId]; ok {

			updatedWL = controller.Workload{
				Id: controller.Workloads[workloadId].Id,
				Filter: controller.Workloads[workloadId].Filter,
				Name: controller.Workloads[workloadId].Name,
				Status: "scheduling",
				Jobs: controller.Workloads[workloadId].Jobs + 1,
				Imgs: controller.Workloads[workloadId].Imgs,
			}

			myWorkload = updatedWL
			id = len(controller.Workloads[workloadId].Imgs) + 1
			
		} else {
			c.JSON(http.StatusConflict, ErrorResponse("Given workload does not exist"))
		}

		fileId := fmt.Sprintf("o%v_%v", id, updatedWL.Name) 
		newFilename := fileId + filepath.Ext(file.Filename)
		downloadFolder := "public/download/" + myWorkload.Name + "/"
		newPath := path.Join(downloadFolder, newFilename)

		updatedWL.Imgs = append(controller.Workloads[workloadId].Imgs, newFilename)
		controller.Workloads[workloadId] = updatedWL

		if err := c.SaveUploadedFile(file, newPath); err != nil {
			c.JSON(http.StatusConflict, ErrorResponse("Could not save the file"))
			return
		}
		registeredImage := controller.Image{Id:id, Name: fileId, Ext: filepath.Ext(file.Filename)}
		controller.Uploads[fileId] = registeredImage

		details := [4]string{newPath, filepath.Ext(file.Filename), workloadId, controller.Workloads[workloadId].Filter}
		sampleJob := scheduler.Job{Address: "localhost:50051", RPCName: "image", Info: details}
		Jobs <- sampleJob
		NumTests += 1
		time.Sleep(time.Second * 1)

		c.JSON(http.StatusOK, UploadResponse(workloadId, fileId, "original"))
	} else {
		c.JSON(http.StatusConflict, ErrorResponse("Your token does not exist yet"))
	}
}

func WorkloadDetails(c *gin.Context) {
	params := strings.Split(c.Request.Header.Get("Authorization"), " ")
	token := params[1]

	workloadId := c.Param("workload_id")
	if _, ok := Users[token]; ok {
		reqWorkload := controller.Workloads[workloadId]
		// status and jobs running calculation
		reqStatus := "running"
		reqJobs := len(reqWorkload.Imgs) - len(controller.Workloads[workloadId].Filtered)
		if len(reqWorkload.Imgs) == len(reqWorkload.Filtered){
			reqStatus = "completed"
		}

		updatedWL := controller.Workload{
			Id: reqWorkload.Id,
			Filter: reqWorkload.Filter,
			Name: reqWorkload.Name,
			Status: reqStatus, 
			Jobs: reqWorkload.Jobs,
			Imgs: reqWorkload.Imgs,
			Filtered: reqWorkload.Filtered,
		}
		controller.Workloads[workloadId] = updatedWL

		c.JSON(http.StatusOK, map[string]interface{}{
			"workload_id": updatedWL.Id,
			"filter":   updatedWL.Filter,
			"workload_name": updatedWL.Name,
			"status": updatedWL.Status,
			"running_jobs": reqJobs,
			"filtered_images": controller.Workloads[workloadId].Filtered,
		})
	} else {
		c.JSON(http.StatusConflict, ErrorResponse("Your token does not exist yet"))
	}
}

func DownloadImage(c *gin.Context) {
	params := strings.Split(c.Request.Header.Get("Authorization"), " ")
	token := params[1]

	image_id := c.Param("image_id")
	if _, ok := Users[token]; ok {
		
		// filtered files
		if string(image_id[0]) == "f" {
			imgInfo := strings.Split(image_id, "_")
			downloadPath := "./public/results/" + imgInfo[1] + "/" + image_id + ".png"
			c.File(downloadPath)
		} else {
			// original files
			ext := controller.Uploads[image_id].Ext
			imgInfo := strings.Split(image_id, "_")
			downloadPath := "./public/download/" + imgInfo[1] + "/" + image_id + ext
			c.File(downloadPath)
		}

	} else {
		c.JSON(http.StatusConflict, ErrorResponse("Your token does not exist yet"))
	}

	
}

func Start(){
	router := gin.Default()
	router.Use(static.Serve("/", static.LocalFile("./public", true)))
	router.POST("/login", Login)
	router.DELETE("/logout", Logout)
	router.GET("/status", Status)
	router.POST("/images", Upload)

	router.POST("/workloads", CreateWorkload)
	router.GET("/workloads/:workload_id", WorkloadDetails)
	router.GET("/images/:image_id", DownloadImage)

	router.Run(":8080")
}