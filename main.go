package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"github.com/joho/godotenv"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Tarea es una estructura para almacenar los detalles de la tarea
type Tarea struct {
	ID          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Titulo      string             `json:"titulo,omitempty" bson:"titulo,omitempty"`
	Descripcion string             `json:"descripcion,omitempty" bson:"descripcion,omitempty"`
}

var client *mongo.Client



func main() {
	fmt.Println("Iniciando aplicación...")
	var err error

	if os.Getenv("APP_ENV") == "local" {
		// Cargar el archivo .env
		envErr := godotenv.Load()
		if envErr != nil {
			log.Fatalf("Error loading .env file: %v", envErr)
		}

	}

	
	connStr := os.Getenv("MONGO_CONNECTION_STRING")
	// Creamos una conexión con MongoDB
	client, err = mongo.NewClient(options.Client().ApplyURI(connStr))
	if err != nil {
		log.Fatal(err)
	}

	// Creamos un contexto con un tiempo límite de 10 segundos
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Conectamos con MongoDB
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Verificar conexión con MongoDB
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Conexión exitosa a MongoDB")

	// Creamos un router con gorilla/mux
	router := mux.NewRouter()

	// Definimos nuestras rutas
	router.HandleFunc("/tareas", getTareas).Methods("GET")
	router.HandleFunc("/tareas/{id}", getTarea).Methods("GET")
	router.HandleFunc("/tareas", createTarea).Methods("POST")
	router.HandleFunc("/tareas/{id}", updateTarea).Methods("PUT")
	router.HandleFunc("/tareas/{id}", deleteTarea).Methods("DELETE")

	port := os.Getenv("PORT")

	fmt.Println("Iniciando servidor en el puerto", port)
	// Iniciamos el servidor HTTP
	log.Fatal(http.ListenAndServe(port, router))
}

func getTareas(w http.ResponseWriter, r *http.Request) {
	// Creamos un contexto con un tiempo límite de 10 segundos
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Obtenemos una colección de tareas
	tareasCollection := client.Database("go-db").Collection("tareas")

	// Buscamos todas las tareas
	cursor, err := tareasCollection.Find(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(ctx)

	// Creamos un slice de tareas vacío
	var tareas []Tarea

	// Iteramos sobre el cursor y añadimos cada tarea al slice
	for cursor.Next(ctx) {
		var tarea Tarea
		if err := cursor.Decode(&tarea); err != nil {
			log.Fatal(err)
		}
		tareas = append(tareas, tarea)
	}

	// Codificamos el slice de tareas como JSON y lo escribimos en la respuesta HTTP
	json.NewEncoder(w).Encode(tareas)
}

func getTarea(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	// Creamos un contexto con un tiempo límite de 10 segundos
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Obtenemos una colección de tareas
	tareasCollection := client.Database("go-db").Collection("tareas")

	// Convertimos el parámetro de la URL en un ObjectID de MongoDB
	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Buscamos la tarea con el ID especificado
	var tarea Tarea
	err = tareasCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&tarea)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Codificamos la tarea como JSON y la escribimos en la respuesta HTTP
	json.NewEncoder(w).Encode(tarea)
}

func createTarea(w http.ResponseWriter, r *http.Request) {
	// Creamos un contexto con un tiempo límite de 10 segundos
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Obtenemos una colección de tareas
	tareasCollection := client.Database("go-db").Collection("tareas")

	// Decodificamos el cuerpo de la solicitud HTTP en una tarea
	var tarea Tarea
	err := json.NewDecoder(r.Body).Decode(&tarea)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Insertamos la tarea en la base de datos
	result, err := tareasCollection.InsertOne(ctx, tarea)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Obtenemos el ID de la tarea insertada y lo añadimos a la tarea
	tarea.ID = result.InsertedID.(primitive.ObjectID)

	// Codificamos la tarea como JSON y la escribimos en la respuesta HTTP
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tarea)
}

func updateTarea(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	// Creamos un contexto con un tiempo límite de 10 segundos
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Obtenemos una colección de tareas
	tareasCollection := client.Database("go-db").Collection("tareas")

	// Convertimos el parámetro de la URL en un ObjectID de MongoDB
	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Decodificamos el cuerpo de la solicitud HTTP en una tarea
	var tarea Tarea
    _ = json.NewDecoder(r.Body).Decode(&tarea)
    tarea.ID = id

	// Crear un filtro para buscar la tarea por ID
    filter := bson.M{"_id": tarea.ID}

	// Crear una actualización utilizando el operador "$set"
    update := bson.M{"$set": bson.M{
        "titulo": tarea.Titulo,
        "descripcion": tarea.Descripcion,
    }}

	_ , updateErr := tareasCollection.UpdateOne(ctx, filter, update)
    if updateErr != nil {
        log.Fatal(updateErr)
    }


	// Codificamos la tarea como JSON y la escribimos en la respuesta HTTP
	json.NewEncoder(w).Encode(tarea)
}

func deleteTarea(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	// Creamos un contexto con un tiempo límite de 10 segundos
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Obtenemos una colección de tareas
	tareasCollection := client.Database("go-db").Collection("tareas")

	// Convertimos el parámetro de la URL en un ObjectID de MongoDB
	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	filter := bson.M{"_id": id}

	// Borramos la tarea de la base de datos
	result, err := tareasCollection.DeleteOne(ctx, filter)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Comprobamos si se borró una tarea
	if result.DeletedCount == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Escribimos una respuesta vacía con un estado 204 No Content
	w.WriteHeader(http.StatusNoContent)
}
