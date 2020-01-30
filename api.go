package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	database string
	password string
	username string
	host     string
	//Collection is the MongoDB collection
	Collection *mgo.Collection
	//Session handles the session for MongoDB
	Session *mgo.Session
)

//init defines variables used in func main
func init() {
	database = os.Getenv("AZURE_DATABASE")
	password = os.Getenv("AZURE_DATABASE_PASSWORD")
	username = os.Getenv("AZURE_DATABASE_USERNAME")
	host = os.Getenv("AZURE_DATABASE_HOST")

	// DialInfo holds options for establishing a session with a MongoDB cluster.
	dialInfo := &mgo.DialInfo{
		Addrs:    []string{fmt.Sprintf("%s.documents.azure.com:10255", host)},
		Timeout:  60 * time.Second,
		Database: database,
		Username: username,
		Password: password,
		DialServer: func(addr *mgo.ServerAddr) (net.Conn, error) {
			return tls.Dial("tcp", addr.String(), &tls.Config{})
		},
	}

	// Create a session which maintains a pool of socket connections to our MongoDB.
	session, err := mgo.DialWithInfo(dialInfo)
	if err != nil {
		fmt.Printf("Can't connect to mongo, go error %v\n", err)
		os.Exit(1)
	}

	//defer session.Close()

	// SetSafe changes the session safety mode.
	session.SetSafe(&mgo.Safe{})

	Session = session
}

//Task model
type Task struct {
	ID          bson.ObjectId `bson:"_id,omitempty"`
	Completed   bool          `bson:"completed"`
	Description string        `bson:"description"`
	Duedate     string        `bson:"duedate"`
	Title       string        `bson:"title"`
}

//SetupRouter sets up routes
func SetupRouter() *gin.Engine {
	router := gin.Default()
	router.POST("/tasks", createNewTask)
	router.GET("/tasks", getAllTasks)
	router.GET("/tasks/:id", getTaskbyID)
	router.PUT("/tasks/:id", updateTaskbyID)
	router.DELETE("/tasks/:id", deleteTaskbyID)
	router.GET("/health", health)
	return router
}

//main function
func main() {
	router := SetupRouter()
	router.Run()
}

//getAllTasks gets all Tasks from the collection
func getAllTasks(c *gin.Context) {
	// grab session from connection pool
	sessionCopy := Session.Copy()
	defer sessionCopy.Close()

	// connect to collection
	Collection = sessionCopy.DB(database).C("tasks")

	result := make([]Task, 0, 15)
	err := Collection.Find(nil).All(&result)
	if err != nil {
		log.Print("Error finding records: ", err)
		c.JSON(500, gin.H{"error": err})
		return
	}
	c.JSON(200, result)
}

//getTaskbyID gets a task by ObjectID
func getTaskbyID(c *gin.Context) {
	id := c.Param("id")
	result := Task{}
	if !bson.IsObjectIdHex(id) {
		log.Print("id is not valid")
		c.JSON(400, gin.H{"error": "id is not in a valid format"})
		return
	}
	// grab session from connection pool
	sessionCopy := Session.Copy()
	defer sessionCopy.Close()

	// connect to collection
	Collection = sessionCopy.DB(database).C("tasks")

	err := Collection.FindId(bson.ObjectIdHex(id)).One(&result)
	if err != nil {
		log.Print("Error finding record: ", err)
		c.JSON(404, gin.H{"error": err})
		return
	}
	c.JSON(200, result)
}

//createNewTask creates a new task
func createNewTask(c *gin.Context) {
	if len(c.PostForm("title")) == 0 {
		log.Print("title missing")
		c.JSON(400, gin.H{"error": "title missing"})
		return
	}
	if len(c.PostForm("description")) == 0 {
		log.Print("description missing")
		c.JSON(400, gin.H{"error": "description missing"})
		return
	}
	// grab session from connection pool
	sessionCopy := Session.Copy()
	defer sessionCopy.Close()

	// connect to collection
	Collection = sessionCopy.DB(database).C("tasks")

	err := Collection.Insert(&Task{
		Completed:   false,
		Description: c.PostForm("description"),
		Duedate:     c.PostForm("duedate"),
		Title:       c.PostForm("title"),
	})
	if err != nil {
		log.Print("Error inserting record: ", err)
		c.JSON(500, gin.H{"error": err})
		return
	}
	c.JSON(201, "successful")
}

//updateTaskbyID updates a task by ObjectID
func updateTaskbyID(c *gin.Context) {
	id := c.Param("id")
	if !bson.IsObjectIdHex(id) {
		log.Print("id is not valid")
		c.JSON(400, gin.H{"error": "id is not in a valid format"})
		return
	}
	result := Task{}
	var completed bool
	// grab session from connection pool
	sessionCopy := Session.Copy()
	defer sessionCopy.Close()

	// connect to collection
	Collection = sessionCopy.DB(database).C("tasks")

	err := Collection.FindId(bson.ObjectIdHex(id)).One(&result)
	if err != nil {
		log.Print("Error finding record: ", err)
		c.JSON(500, gin.H{"error": err})
		return
	}
	if len(c.PostForm("completed")) == 0 {
		completed = result.Completed
	} else {
		completed, err = strconv.ParseBool(c.PostForm("completed"))
	}
	description := c.DefaultPostForm("description", result.Description)
	duedate := c.DefaultPostForm("duedate", result.Duedate)
	title := c.DefaultPostForm("title", result.Title)
	updateQuery := bson.M{"_id": bson.ObjectIdHex(id)}
	change := bson.M{"$set": bson.M{
		"completed":   completed,
		"description": description,
		"duedate":     duedate,
		"title":       title,
	}}
	err = Collection.Update(updateQuery, change)
	if err != nil {
		log.Print("Error updating record: ", err)
		c.JSON(500, gin.H{"error": err})
		return
	}
	c.JSON(200, "update successful")
}

//deleteTaskbyID deletes a task by ObjectID
func deleteTaskbyID(c *gin.Context) {
	id := c.Param("id")
	if !bson.IsObjectIdHex(id) {
		log.Print("id is not valid")
		c.JSON(400, gin.H{"error": "id is not in a valid format"})
		return
	}
	updateQuery := bson.M{"_id": bson.ObjectIdHex(id)}
	// grab session from connection pool
	sessionCopy := Session.Copy()
	defer sessionCopy.Close()

	// connect to collection
	Collection = sessionCopy.DB(database).C("tasks")

	err := Collection.Remove(updateQuery)
	if err != nil {
		log.Print("Error deleting record: ", err)
		c.JSON(500, gin.H{"error": err})
		return
	}
	c.JSON(200, "delete successful")
}

//health used for health probes
func health(c *gin.Context) {
	c.JSON(200, "im alive")
}
