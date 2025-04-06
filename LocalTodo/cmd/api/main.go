package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
)

var dbFile *os.File

const dbPath = "./db.json"

func main() {
	err := openFile()
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}

	defer dbFile.Close()

	switch os.Args[1] {
	case "createTodo":
		if len(os.Args) < 3 {
			log.Fatal("Usage: gorun createTodo <title>")
		}
		err := createTodo(os.Args[2])
		if err != nil {
			log.Printf("Error creating todo: %v", err)
			os.Exit(1)
		}
		fmt.Println("Todo created successfully")

	case "listTodos":
		todos, err := listTodos()
		if err != nil {
			log.Printf("Error listing todos: %v", err)
			os.Exit(1)
		}
		fmt.Println("Todos:")
		for _, todo := range todos {
			fmt.Printf(
				"{ID: %d, Title: %s, Completed: %t}\n",
				todo.ID,
				todo.Title,
				todo.Completed,
			)
		}
	case "updateTodos":
		if len(os.Args) < 3 {
			log.Fatal("Usage: gorun updateTodos <id> <title> <completed>")
		}
		id := os.Args[2]
		title := ""
		completed := ""
		if len(os.Args) > 3 {
			title = os.Args[3]
		}
		if len(os.Args) > 4 {
			completed = os.Args[4]
		}

		err := updateTodos(id, title, completed)
		if err != nil {
			log.Printf("Error updating todo: %v", err)
			os.Exit(1)
		}
	case "clearTodos":
		err := clearTodos()
		if err != nil {
			log.Printf("Error clearing todos: %v", err)
			os.Exit(1)
		}

	default:
		log.Fatalf("Unknown command: %s", os.Args[1])
	}
}

func openFile() error {
	jsonFile, err := os.OpenFile(dbPath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}

	dbFile = jsonFile
	info, err := dbFile.Stat()
	if err != nil {
		return err
	}

	fmt.Printf("File size: %d bytes\n", info.Size())

	return nil
}

func readFile() (*Todos, error) {
	var todos Todos

	if dbFile == nil {
		return nil, fmt.Errorf("file is not opened")
	}

	_, err := dbFile.Seek(0, io.SeekStart)
	if err != nil {
		return nil, fmt.Errorf("failed to seek to start of file: %w", err)
	}

	err = json.NewDecoder(dbFile).Decode(&todos)
	if err != nil {
		if errors.Is(err, io.EOF) {
			encoder := json.NewEncoder(dbFile)
			encoder.Encode(&Todos{Todos: []Todo{}})
			return &todos, nil
		}
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	return &todos, nil
}

func createTodo(title string) error {
	todos, err := readFile()
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	maxId := uint8(1)
	for _, todo := range todos.Todos {
		if todo.ID >= maxId {
			maxId = todo.ID + 1
		}
	}

	newTodo := Todo{
		ID:        maxId,
		Title:     title,
		Completed: false,
	}

	_, err = dbFile.Seek(0, io.SeekStart)
	if err != nil {
		return fmt.Errorf("failed to seek to start of file: %w", err)
	}
	err = dbFile.Truncate(0)
	if err != nil {
		return fmt.Errorf("failed to truncate file: %w", err)
	}

	encoder := json.NewEncoder(dbFile)
	todos.Todos = append(todos.Todos, newTodo)
	err = encoder.Encode(todos)
	if err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

func listTodos() ([]Todo, error) {
	todos, err := readFile()
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	return todos.Todos, nil
}

func updateTodos(idStr, title, completedStr string) error {
	todos, err := readFile()
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}
	id, err := strconv.ParseInt(idStr, 10, 8)
	if err != nil {
		return fmt.Errorf("invalid ID: %w", err)
	}

	found := false
	for i, todo := range todos.Todos {
		if todo.ID == uint8(id) {
			found = true

			if title != "" {
				todos.Todos[i].Title = title
			}
			if completedStr != "" {
				completedBool, err := strconv.ParseBool(completedStr)
				if err != nil {
					return fmt.Errorf("invalid completed value: %w", err)
				}
				todos.Todos[i].Completed = completedBool
			}
		}
	}
	if !found {
		return fmt.Errorf("todo with ID %d not found", id)
	}

	_, err = dbFile.Seek(0, io.SeekStart)
	if err != nil {
		return fmt.Errorf("failed to seek to start of file: %w", err)
	}
	err = dbFile.Truncate(0)
	if err != nil {
		return fmt.Errorf("failed to truncate file: %w", err)
	}
	err = json.NewEncoder(dbFile).Encode(todos)
	if err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

func clearTodos() error {
	_, err := dbFile.Seek(0, io.SeekStart)
	if err != nil {
		return fmt.Errorf("failed to seek to start of file: %w", err)
	}
	err = dbFile.Truncate(0)
	if err != nil {
		return fmt.Errorf("failed to truncate file: %w", err)
	}

	encoder := json.NewEncoder(dbFile)
	err = encoder.Encode(&Todos{Todos: []Todo{}})
	if err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

type Todos struct {
	Todos []Todo `json:"Todos"`
}

type Todo struct {
	ID        uint8  `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}
