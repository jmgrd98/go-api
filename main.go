package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Todo struct {
	ID        primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	Completed bool               `json:"completed"`
	Body      string             `json:"body"`
}

var collection *mongo.Collection

func main() {
	app := fiber.New()

	// Load environment variables
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error loading .env file")
	}

	// MongoDB connection
	MONGODB_URI := os.Getenv("MONGODB_URI")
	fmt.Println(MONGODB_URI)
	clientOptions := options.Client().ApplyURI(MONGODB_URI)

	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	// Verify connection
	if err = client.Ping(context.Background(), nil); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to MongoDB!")

	collection = client.Database("go-api").Collection("todos")

	// Routes
	app.Get("/api/todos", getTodos)
	app.Post("/api/todos", createTodo)
	app.Patch("/api/todos/:id", updateTodo)
	app.Delete("/api/todos/:id", deleteTodo)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}
	log.Fatal(app.Listen(":" + port))
}

func getTodos(c *fiber.Ctx) error {
	var todos []Todo
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer cursor.Close(context.Background())

	if err = cursor.All(context.Background(), &todos); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(todos)
}

func createTodo(c *fiber.Ctx) error {
	todo := new(Todo)
	if err := c.BodyParser(todo); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if todo.Body == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Todo body is required"})
	}

	result, err := collection.InsertOne(context.Background(), todo)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	todo.ID = result.InsertedID.(primitive.ObjectID)
	return c.Status(201).JSON(todo)
}

func updateTodo(c *fiber.Ctx) error {
	id, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID format"})
	}

	update := bson.M{"$set": bson.M{"completed": true}}
	result, err := collection.UpdateOne(
		context.Background(),
		bson.M{"_id": id},
		update,
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	if result.MatchedCount == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Todo not found"})
	}

	return c.JSON(fiber.Map{"success": true})
}

func deleteTodo(c *fiber.Ctx) error {
	id, err := primitive.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID format"})
	}

	result, err := collection.DeleteOne(context.Background(), bson.M{"_id": id})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	if result.DeletedCount == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Todo not found"})
	}

	return c.JSON(fiber.Map{"success": true})
}
