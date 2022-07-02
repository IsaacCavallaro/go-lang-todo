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

func homeHandler(w http.ResponseWriter, r *http.Request){
	err := rnd.Template(w, http.StatusOK, []string{"static/home.tpl", nil})
	checkErr(err)
}

func fetchTodos(w http.ResponseWriter, r *http.Request){
	todos := []todoModel{}

	if err := db.c(collectionName).Find(bson.M{}).All(&todos); err!=nil {
		rnd.JSON(w, http.StatusProcessing, renderer.M{
			"message":"Failed to fetch todo",
			"error":err,
		})
		return
	}
	todoList := []todo{}

	for _,t := range todos{
		todoList = append(todoList, todo{
			ID: t.ID.Hex(), 
			Title: t.Title,
			Completed: t.Completed,
			CreatedAt: t.CreatedAt,
		})
	}
	rnd.JSON(w, http.StatusOK, renderer.M{
		"data": todoList,
	})
}

func createTodo(w http.ResponseWriter, r *http.Request){
	var t todo

	if err := json.NewDecoder(r.Body).Decode(&t); err!=nil {
		rnd.JSON(w, http.StatusProcessing, err)
		return
	}

	if t.Title == "" {
		rnd.JSON(w, http.StatusBadRequest, renderer.M(
			"Message: The title is required"
		))
		return
	}

	tm := todoModel{
		ID: bson.NewObjectid(), 
		Title: t.Title,
		Completed: false,
		CreatedAt: time.now(),
	}

	if err := db.c(collectionName).Insert(&tm); err !=nil {
		rnd.JSON(w, http.StatusProcessing,renderer.M(
			"Message: Failed to save todo"
		))
		return
	}

	rnd.JSON(w, http.statusCreated, renderer.M(
		"Message: todo created successfully"
		"todo_id:" tm.ID.Hex()
	))
}

func main(){
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt)
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", homeHandler)
	r.Mount("/todo", todoHandlers())

	srv := &http.Server{
		Addr: port,
		Handler: r,
		ReadTimeout: 60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IndleTimeout: 60 * time.Second
	}
	go func(){
		log.Println("listening on port", port)
		if err:=srv.ListenAndServe(); err!=nil {
			log.Printf("listen:%s\n", err)
		}
	}
	log.Println("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	srv.Shutdown(ctx)
	defer cancel(
		log.Println("server gracefully stoped.")
	)
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