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

const (
	indexTemplatePath string = "views/index.html"
	rosterFilePath    string = "latest_grades.json"
)

var (
	currentRoster     Roster
	idMap             map[string]string
	templates         *template.Template
	templatesModified os.FileInfo
)

func serveError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
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
			return
		}
	} else {
		// Use for API call:
		fmt.Fprintf(w, "{\"error\":\"Fields missing from submission\"}")

		// Use for form post:
		// http.Redirect(w, r, "/", 302)
	}
}

func ReloadTemplates() {
	// Has the index file changed?
	templatesCurrent, err := os.Lstat(indexTemplatePath)
	if err != nil {
		panic(err)
	}
	if templatesCurrent.ModTime() != templatesModified.ModTime() {
		fmt.Println("hot dang, index's differn't")
		// Reload template(s)
		if templates, err = template.ParseFiles(
			indexTemplatePath,
		); err != nil {
			panic(err)
		}
		templatesModified = templatesCurrent
	}
}

func ReloadRoster() {
	// Has the roster file changed?
	rosterCurrent, err := os.Lstat(rosterFilePath)
	if err != nil {
		panic(err)
	}
	if currentRoster.LastModified == nil || rosterCurrent.ModTime() != currentRoster.LastModified.ModTime() {
		fmt.Println("hot dang, roster's differn't")

		LoadRoster()

		currentRoster.LastModified = rosterCurrent
	}
}

func LoadRoster() {
	file, err := ioutil.ReadFile(rosterFilePath)
	if err != nil {
		fmt.Printf("File error: %v\n", err)
		panic(err)
	}
	json.Unmarshal(file, &currentRoster)

	SeedIdMap()
}

func SeedIdMap() {
	idMap = make(map[string]string, len(currentRoster.Students))
	for _, student := range currentRoster.Students {
		idMap[student.Affiliate] = strings.ToLower(student.LastName)
	}
}

func refresh(w http.ResponseWriter, r *http.Request) {
	ReloadTemplates()
	ReloadRoster()
	http.Redirect(w, r, "/", 302)
}

func index(w http.ResponseWriter, r *http.Request) {
	if err := templates.ExecuteTemplate(w, "index.html", r.URL.Path[1:]); err != nil {
		serveError(w, err)
	}
}

func init() {
	var err error

	// Parse template(s)
	templates, err = template.New("").ParseFiles(indexTemplatePath)
	if err != nil {
		panic(err)
	}

	// Remember it's last modified time/size
	templatesModified, err = os.Lstat(indexTemplatePath)
	if err != nil {
		panic(err)
	}

	// Import class information
	LoadRoster()

	// Remember it's last modified time/size
	currentRoster.LastModified, err = os.Lstat(rosterFilePath)
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
	http.HandleFunc("/refresh", refresh)
	http.HandleFunc("/", index)
	http.ListenAndServe(":8080", nil)
}
