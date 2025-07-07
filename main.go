package main

import (
	"bookmark-cli/db"
	"bookmark-cli/repository"
	"bookmark-cli/utils"
	"bufio"
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

	// prompt for user name
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your name: ")
	name, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading input:", err)
		return
	}
	name = strings.TrimSpace(name)

	// prompt for email until a valid one is entered
	var email string
	for {
		fmt.Print("Enter your email: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			return
		}
		input = strings.TrimSpace(input)

		if !utils.ValidateEmail(input) {
			fmt.Println("Invalid email format. Try again.")
			continue
		}
		email = input
		break
	}

	// create a new user in the database
	user, err := repository.CreateUser(database, name, email)
	if err == repository.ErrEmailTaken {
		fmt.Println("That email is already taken. Please try another one.")
		return
	} else if err != nil {
		fmt.Println("could not create user.", err)
		return
	}
	fmt.Printf("Welcome %s! Your user ID is %d. Save this ID to fetch your saved bookmarks later.\n", user.Name, user.ID)

	// keep asking for bookmarks until the user decides to stop
	for {
		fmt.Printf("Add a bookmark? (y/n): ")
		ans, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			return
		}
		ans = strings.ToLower(strings.TrimSpace(ans))
		if ans != "y" {
			break
		}

		fmt.Print("Title: ")
		title, _ := reader.ReadString('\n')
		title = strings.TrimSpace(title)

		fmt.Print("URL: ")
		url, _ := reader.ReadString('\n')
		url = strings.TrimSpace(url)

		bm, err := repository.CreateBookmark(database, user.ID, title, url)
		if err != nil {
			fmt.Println("could not save bookmark: ", err)
			continue
		}
		fmt.Printf("Saved: %s\n", bm.Title)
	}

	for {
		fmt.Print("Do you want to fetch your bookmarks? (y/n): ")
		ans, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			return
		}
		ans = strings.ToLower(strings.TrimSpace(ans))
		if ans != "y" {
			break
		}
		fmt.Print("Enter your user ID to fetch your bookmarks: ")
		var userID int
		_, err = fmt.Scanf("%d", &userID)
		if err != nil {
			fmt.Println("Invalid user ID. Please enter a valid number.")
			continue
		}
		fmt.Println("\nYour bookmarks:")
		bms, err := repository.ListBookmarks(database, userID)
		if err != nil {
			fmt.Println("could not retrieve bookmarks:", err)
			continue
		}
		if len(bms) == 0 {
			fmt.Println("Empty. You haven't added any bookmarks yet.")
		}
		for _, bm := range bms {
			fmt.Printf("Title: %s, URL: %s, Created At: %s\n", bm.Title, bm.URL, bm.CreatedAt)
		}
		break
	}
}
