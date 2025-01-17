package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"testing/utils"
    "time"
)

var addFriend = `
<!DOCTYPE html>
<html>
<head>
    <title>Kyle Connect - Add a friend / Send Friend Request</title>
</head>
<body>
    <h1>Add a friend</h1>
	<p>Enter a name of a friend that you would like to add</p>
    <form method="POST" action="/form">
        <label for="username">Username:</label>
        <input type="username" id="username" name="username" placeholder="Enter username to add" required><br><br>

        <h1>Accept or Decline Friend Request</h1>
        <p>Enter accept or decline for your answer</p>
        <label for="answer">Answer:</label>
        <input type="answer" id="answer" name="answer" placeholder="Enter your answer" required><br><br>

        <input type="submit" value="Send friend request">
    </form>

    {{if .Username}}
    <p>Your friend request has been sent to {{.Username}}!</p>
    <p>Your answer is: {{.Answer}}!</p>
    {{end}}
</body>
</html>
`

type User struct {
    Username string
    Answer string
	UI template.HTML
}

func convertStringToInt(number string) int {
    i, err := strconv.Atoi(number)
    if err != nil {
        panic(err)
    }
    return i
}

func convertIntToString(number int) string {
	str := strconv.Itoa(number)
    return str
}

// Handler function to serve the form and process submissions
func formHandler(w http.ResponseWriter, r *http.Request) {
    // Parse the HTML template
    tmpl, err := template.New("form").Parse(addFriend)
    if err != nil {
        log.Fatal(err)
    }

    // Initialize form data
    userData := User{}

    if r.Method == http.MethodPost {
        // Parse form data
        err := r.ParseForm()
        if err != nil {
            http.Error(w, "Error parsing form data", http.StatusBadRequest)
            return
        }

        userData.Username = r.FormValue("username")
		enteredUsername := userData.Username

        userData.Answer = r.FormValue("answer")
        ans := strings.TrimSpace(strings.ToLower(userData.Answer))

        // db, err := sql.Open("mysql", "root@tcp(127.0.0.1)/kyleconnect") // this line of code works for localhost but not docker!
        db, err := sql.Open("mysql", "root@tcp(host.docker.internal:3306)/kyleconnect?parseTime=true")

		username, er := utils.RetrieveUsernameFromDb(db, enteredUsername)
		if err != nil {
            fmt.Println("this is what is driving me insane...................")
			log.Fatal(er)
		}
    
		if er != nil {
			log.Fatal("error retrieving username from database", er)
		}

        // this should get a userid for a given inputted username
        id, err := utils.GetUserId(db, username)

        // convert the id variable above from a string to an int 
        idConverted := convertStringToInt(id)
        if err != nil {
            fmt.Println("error finding userid for", id)
        } 
        
        // I have made the logged-in table 
        // I have altered my login.go file to insert some data into this new table, but the name is hard-coded (so i need to write a query to get a username from an email)
        // now the data is in the logged-in table, I now need to get that data

        lastUser, e := utils.GetLastUserLoggedIn(db)
        if e != nil {
            fmt.Println("error getting last user in logged-in table", e)
        }
        utils.CatchError(e)
        fmt.Println("value variable = ", lastUser) // value is 5 because that is the last element in the db table

        lastUserFromIntToString := convertIntToString(lastUser)

        emailId, e := utils.RetrieveEmailFromId(db, lastUserFromIntToString)
        if e != nil {
            fmt.Println("error getting email from id", e)
        }
        fmt.Println("emailId is:", emailId)
        
        loggedInUser, e := utils.RetrieveEmail(db, emailId)
        if e != nil {
            fmt.Println("error finding name for", username)
        }
        fmt.Println("loggedInUser is:", loggedInUser)

        loggedInUserId, er := utils.GetUserId(db, loggedInUser)
        convertedValue := convertStringToInt(loggedInUserId)

        status := "pending"
        // if we have any pending values we can either accept them or not
        // if we accept the friend request then I, add my friend to a new db / table called "friends" --- done
        // if I say decline then I, delete that friend out of the table --- done 
        
		if strings.TrimSpace(strings.ToLower(username)) == strings.TrimSpace(strings.ToLower(enteredUsername)) {

            if loggedInUser == username {
                fmt.Println("you cannot send a friend reuqest to yourself!")
                return
            }

            // below I put the values for the friend request into a friend request table 
            utils.PutDataToFriendRequestTable(db, convertedValue, loggedInUser, idConverted, username, status)
            time.Sleep(time.Second*5)

            if status == "pending" {

                if ans == "accept" {
                    status = "accept"
                    utils.UpdateFriendRequestStatus(db, status, username)
                    
                    // here I put user1 and user2 into the friends table
                    utils.PutFriendsToFriendsTable(db, loggedInUser, username) // need to fix the hard-coded Kyle!
                } 
    
                if ans == "decline" {
                    time.Sleep(time.Second*5)
                    status = "decline"
                    utils.DeclineFriendRequest(db, username)
                }
            } else {
                fmt.Println("no status is:", status)
            }
		} else {
			fmt.Println("noooo", er)
		}
        utils.CatchError(err)
        defer db.Close()
    }
    // Render the HTML template with the form data
    tmpl.Execute(w, userData)
}

func main() {
    // Set up the route and handler for the form
    http.HandleFunc("/", formHandler)

    // Start the HTTP server
    fmt.Println("Server started at http://localhost:8082")
    log.Fatal(http.ListenAndServe(":8082", nil))
}
