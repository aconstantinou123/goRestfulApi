package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/fatih/structs"
	"github.com/gorilla/mux"
	"github.com/mitchellh/mapstructure"
	"github.com/neo4j/neo4j-go-driver/neo4j"
)

var (
	uri      string = "bolt://neo4j:7687"
	username string = "neo4j"
	password string = "test"
	query    string
)

var (
	driver  neo4j.Driver
	session neo4j.Session
	result  neo4j.Result
	err     error
)

//Event - struct
type Event struct {
	ID          string `json:"ID"`
	Title       string `json:"Title"`
	Description string `json:"Description"`
}

//Person - struct
type Person struct {
	Name string `json:"Name"`
	Age  int64  `json:"Age"`
}

var event1 = Event{
	ID:          "1",
	Title:       "Introduction to Golang",
	Description: "Come join us for a chance to learn how golang works and get to eventually try it out",
}
var event2 = Event{
	ID:          "2",
	Title:       "Advanced Golang",
	Description: "Come join us for a chance to learn advanced golang",
}

var person1 = Person{
	Name: "Vladimir Putin",
	Age:  67,
}

var person2 = Person{
	Name: "Xi Jinping",
	Age:  66,
}

func establishConnection() error {
	if driver, err = neo4j.NewDriver(uri, neo4j.BasicAuth(username, password, ""), func(config *neo4j.Config) {
		config.Log = neo4j.ConsoleLogger(neo4j.ERROR)
	}); err != nil {
		return err
	}
	if session, err = driver.Session(neo4j.AccessModeWrite); err != nil {
		return err
	}
	return nil
}

func runSeeds() error {
	_, err = session.Run("MATCH (n) DETACH DELETE n", map[string]interface{}{})
	if err != nil {
		return err
	}
	result, err = session.Run("CREATE (n:Person { Name: $Name, Age: $Age }) - [:ATTENDED]"+
		"-> (m:Event { ID: $ID, Title: $Title, Description: $Description })",
		map[string]interface{}{
			"Name":        person1.Name,
			"Age":         person1.Age,
			"ID":          event1.ID,
			"Title":       event1.Title,
			"Description": event1.Description,
		})
	if err != nil {
		return err
	}
	result, err = session.Run("CREATE (n:Person { Name: $Name, Age: $Age }) - [:ATTENDED]"+
		"-> (m:Event { ID: $ID, Title: $Title, Description: $Description })",
		map[string]interface{}{
			"Name":        person2.Name,
			"Age":         person2.Age,
			"ID":          event2.ID,
			"Title":       event2.Title,
			"Description": event2.Description,
		})
	if err != nil {
		return err
	}
	return nil
}

func getAll(w http.ResponseWriter, r *http.Request) {
	persons := []Person{}
	eventsToReturn := []Event{}
	result, err := session.Run("MATCH (persons:Person) - [:ATTENDED] -> (events:Event) RETURN persons, events",
		map[string]interface{}{})
	if err != nil {
		fmt.Fprintf(w, "db err")
	}
	for result.Next() {
		record := result.Record()
		personToAdd := Person{}
		if value, ok := record.Get("persons"); ok {
			node := value.(neo4j.Node)
			props := node.Props()
			err := mapstructure.Decode(props, &personToAdd)
			if err != nil {
				fmt.Fprintf(w, "parse err")
			}
		}
		eventToAdd := Event{}
		if value, ok := record.Get("events"); ok {
			node := value.(neo4j.Node)
			props := node.Props()
			err := mapstructure.Decode(props, &eventToAdd)
			if err != nil {
				fmt.Fprintf(w, "parse err")
			}
		}
		persons = append(persons, personToAdd)
		eventsToReturn = append(eventsToReturn, eventToAdd)
	}
	w.WriteHeader(http.StatusAccepted)
	data := map[string]interface{}{
		"events":  eventsToReturn,
		"persons": persons,
	}
	json.NewEncoder(w).Encode(data)
}

func seeds(w http.ResponseWriter, r *http.Request) {
	if err := runSeeds(); err != nil {
		log.Fatal(err)
		fmt.Fprintf(w, "Error running seeds")
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode("Done")
}

func createEvent(w http.ResponseWriter, r *http.Request) {
	newEvent := Event{}
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "Incorrect format")
	}
	json.Unmarshal(reqBody, &newEvent)
	cypher := `CREATE(event:Event) SET event = $prop RETURN event`
	result, err := session.Run(cypher, map[string]interface{}{
		"prop": structs.Map(newEvent),
	})
	createdEvent := Event{}
	for result.Next() {
		record := result.Record()
		if value, ok := record.Get("event"); ok {
			node := value.(neo4j.Node)
			props := node.Props()
			err := mapstructure.Decode(props, &createdEvent)
			if err != nil {
				fmt.Fprintf(w, "parse err")
			}
		}
	}
	json.NewEncoder(w).Encode(createdEvent)
}

func createPerson(w http.ResponseWriter, r *http.Request) {
	newPerson := Person{}
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "Incorrect format")
	}
	json.Unmarshal(reqBody, &newPerson)
	cypher := `CREATE(person:Person) SET person = $prop RETURN person`
	result, err := session.Run(cypher, map[string]interface{}{
		"prop": structs.Map(newPerson),
	})
	createdPerson := Person{}
	for result.Next() {
		record := result.Record()
		if value, ok := record.Get("person"); ok {
			node := value.(neo4j.Node)
			props := node.Props()
			err := mapstructure.Decode(props, &createdPerson)
			if err != nil {
				fmt.Fprintf(w, "parse err")
			}
		}
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdPerson)
}

// func getOneEvent(w http.ResponseWriter, r *http.Request) {
// 	eventID := mux.Vars(r)["id"]
// 	for _, singleEvent := range events {
// 		if singleEvent.ID == eventID {
// 			json.NewEncoder(w).Encode(singleEvent)
// 		}
// 	}
// }

// func updateEvent(w http.ResponseWriter, r *http.Request) {
// 	eventID := mux.Vars(r)["id"]
// 	var updatedEvent event

// 	reqBody, err := ioutil.ReadAll(r.Body)
// 	if err != nil {
// 		fmt.Fprintf(w, "Please enter data with the event title and description only in order to update")
// 	}
// 	json.Unmarshal(reqBody, &updatedEvent)

// 	for i, singleEvent := range events {
// 		if singleEvent.ID == eventID {
// 			singleEvent.Title = updatedEvent.Title
// 			singleEvent.Description = updatedEvent.Description
// 			events = append(events[:i], singleEvent)
// 			json.NewEncoder(w).Encode(singleEvent)
// 		}
// 	}
// }

// func deleteEvent(w http.ResponseWriter, r *http.Request) {
// 	eventID := mux.Vars(r)["id"]

// 	for i, singleEvent := range events {
// 		if singleEvent.ID == eventID {
// 			events = append(events[:i], events[i+1:]...)
// 			fmt.Fprintf(w, "The event with ID %v has been deleted successfully", eventID)
// 		}
// 	}
// }

func homeLink(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome home!")
}

func commonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func main() {
	if err := establishConnection(); err != nil {
		log.Fatal(err)
		fmt.Println("Cannot connect to db")
	}
	port := ":8080"
	router := mux.NewRouter().StrictSlash(true)
	router.Use(commonMiddleware)
	router.HandleFunc("/", homeLink)
	router.HandleFunc("/seeds", seeds).Methods("POST")
	router.HandleFunc("/events", createEvent).Methods("POST")
	router.HandleFunc("/person", createPerson).Methods("POST")
	router.HandleFunc("/events", getAll).Methods("GET")
	// router.HandleFunc("/events/{id}", getOneEvent).Methods("GET")
	// router.HandleFunc("/events/{id}", updateEvent).Methods("PATCH")
	// router.HandleFunc("/events/{id}", deleteEvent).Methods("DELETE")
	fmt.Println("Server listening on port", port)
	log.Fatal(http.ListenAndServe(port, router))
	defer driver.Close()
	defer session.Close()
}
