package main

import(
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
	"context"
	"os"
	"os/signal"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/thedevsaddam/renderer"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var rnd *renderer.Render
var db *mgo.Database

const (
	hostName 		string = "localhost:27017"
	dbName 			string = "demo_todo"
	collectionName 	string = "todo"
	port 			string = ":9000"
)

type (
	todoModel struct{
		ID		bson.ObjectId`bson:"_id,omitempty"`
		Title 	string = `bson:"title"`
		Completed bool `bson:"completed"`
		CreatedAt time.Time `bson:"createdAt"`
	}
	todo struct{
		ID		string `json:"id"`
		Title	string `json:"title"`
		Completed string `json:"completed"`
		CreatedAt time.Time `json:"created_at"`
	}	
)

func init(){
	rnd = renderer.New()
	sess, err:=mgo.Dial(hostName)
	checkErr(err)
	sess.SetMode(mgo.Monotonic, true)
	db = sess.DB(dbName)
}

func main(){
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", homeHandler)
	r.Mount("/todo", todoHandlers())
}

func todoHandlers() http.Handler{
	rg:= chi.NewRouter()
	rg.Group(func(r chi.Router){
		r.Get("/", fetchTodos)
		r.Post("/", creatTodo)
		r.Put("/{id}", updateTodo)
		r.Delete("/{id}") deleteTodo
	})
	return rg
}

func checkErr(err error){
	if err!=nil{
		log.Fatal(err)
	}
}