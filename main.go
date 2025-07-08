package main

import (
	"bookmark-cli/db"
	"bookmark-cli/repository"
	"bookmark-cli/utils"
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, relying on system variables.")
	}

	// connect to the database and close it when done
	database, err := db.Connect()
	if err != nil {
		fmt.Println("could not connect to the database:", err)
		return
	}
	defer database.Close()

	reader := bufio.NewReader(os.Stdin)
	var user *repository.User
	var email string

	// Ask for email first
	for {
		fmt.Print("Enter your email: ")
		email, _ = reader.ReadString('\n')
		email = strings.TrimSpace(email)

		if !utils.ValidateEmail(email) {
			fmt.Println("Invalid email format.")
			continue
		}
		break
	}

	// Check if user exists
	user, err = repository.GetUserByEmail(database, email)
	if err != nil {
		if err == sql.ErrNoRows {
			// User does not exist, proceed to create a new user
			fmt.Println("Welcome! Let's create your account.")
			fmt.Print("Enter your name: ")
			name, _ := reader.ReadString('\n')
			name = strings.TrimSpace(name)

			user, err = repository.CreateUser(database, name, email)
			if err != nil {
				fmt.Printf("Could not create user: %v\n", err)
				return
			}
			fmt.Printf("Account created for %s. Your user ID is %d. Save this ID to fetch your saved bookmarks later.\n", user.Name, user.ID)
		} else {
			// Another database error occurred
			fmt.Printf("Error retrieving user: %v\n", err)
			return
		}
	} else {
		// user exists
		fmt.Printf("Welcome back, %s! Your user ID is %d.\n", user.Name, user.ID)
	}

	// Main application loop
	for {
		fmt.Print("\nWhat would you like to do?\n (1) Add a bookmark\n (2) List my bookmarks\n (3) Exit\n")
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			repository.AddBookmark(database, reader, user.ID)
		case "2":
			repository.ListBookmarks(database, user.ID)
		case "3":
			fmt.Println("Goodbye!")
			return
		default:
			fmt.Println("Invalid choice. Please enter 1, 2, or 3.")
		}
	}

}
