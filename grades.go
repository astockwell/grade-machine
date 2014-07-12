package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

// Struct to JSON
// TODO: Add functions to structs
type Assignment struct {
	Name     string
	Total    string
	Score    string
	Comments string
	Category string
}

type Student struct {
	Affiliate    string
	FirstName    string
	LastName     string
	CurrentGrade string
	Assignments  []Assignment
}

type Roster struct {
	Students     []Student
	LastModified os.FileInfo
}

var (
	currentRoster     Roster
	idMap             map[string]string
	templates         *template.Template
	templatesModified os.FileInfo
)

func serveError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func handler(w http.ResponseWriter, r *http.Request) {
	// Has the index file changed?
	templatesCurrent, err := os.Lstat("views/index.html")
	if err != nil {
		panic(err)
	}
	if templatesCurrent.ModTime() != templatesModified.ModTime() {
		fmt.Println("hot dang, index's differnt")
		// Reload template(s)
		if templates, err = template.ParseFiles(
			"views/index.html",
		); err != nil {
			panic(err)
		}
		templatesModified = templatesCurrent
	}

	// Has the roster file changed?
	rosterCurrent, err := os.Lstat("./latest_grades.json")
	if err != nil {
		panic(err)
	}
	if rosterCurrent.ModTime() != currentRoster.LastModified.ModTime() {
		fmt.Println("hot dang, roster's differnt")
		// Reload roster
		file, err := ioutil.ReadFile("./latest_grades.json")
		if err != nil {
			fmt.Printf("File error: %v\n", err)
			panic(err)
		}
		json.Unmarshal(file, &currentRoster)
		idMap = make(map[string]string, len(currentRoster.Students))
		for _, student := range currentRoster.Students {
			idMap[student.Affiliate] = strings.ToLower(student.LastName)
		}
		currentRoster.LastModified = rosterCurrent
	}

	// Serve it!
	if err := templates.ExecuteTemplate(w, "index.html", r.URL.Path[1:]); err != nil {
		serveError(w, err)
	}
}

func grades(w http.ResponseWriter, r *http.Request) {
	// Handle JSON post request
	var s Student
	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&s); err == io.EOF {
		// It's not JSON
		fmt.Fprintf(w, "{\"error\":\"Fields missing from submission\"}")
		return
	} else if err != nil {
		// Don't know what it is (possibly form post)
		fmt.Println("Error:", err)
		panic(err)
	} else {
		// It's JSON alright
		w.Header().Set("Content-Type", "application/json")
	}

	if r.ContentLength > 0 && len(s.LastName) > 0 {
		if strings.ToLower(s.LastName) == idMap[s.Affiliate] {
			var index int
			for i, b := range currentRoster.Students {
				if b.Affiliate == s.Affiliate {
					index = i
					break
				}
			}

			j, err := json.Marshal(currentRoster.Students[index])
			if err != nil {
				panic(err)
			}
			fmt.Fprintf(w, string(j))
		} else {
			fmt.Fprintf(w, "{\"error\":\"No match for ID and last name\"}")
		}
	} else {
		// Use for API call:
		fmt.Fprintf(w, "{\"error\":\"Fields missing from submission\"}")

		// Use for form post:
		// http.Redirect(w, r, "/", 302)
	}
}

func init() {
	var err error

	// Parse template(s)
	if templates, err = template.New("").ParseFiles(
		"views/index.html",
	); err != nil {
		panic(err)
	}

	// Remember it's last modified time/size
	templatesModified, err = os.Lstat("views/index.html")
	if err != nil {
		panic(err)
	}

	// Import class information
	file, err := ioutil.ReadFile("./latest_grades.json")
	if err != nil {
		fmt.Printf("File error: %v\n", err)
		panic(err)
	}
	json.Unmarshal(file, &currentRoster)

	// Create authenticaion lookup
	idMap = make(map[string]string, len(currentRoster.Students))
	for _, student := range currentRoster.Students {
		idMap[student.Affiliate] = strings.ToLower(student.LastName)
	}

	// Remember it's last modified time/size
	currentRoster.LastModified, err = os.Lstat("./latest_grades.json")
	if err != nil {
		panic(err)
	}

	// All good?
	fmt.Println(":)")
}

func main() {
	// serve CSS static assets first
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	http.HandleFunc("/grades", grades)
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
