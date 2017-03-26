// Golang HTML5 Server Side Events Example
//
// Run this code like:
//  > go run server.go
//
// Then open up your browser to http://localhost:8000
// Your browser must support HTML5 SSE, of course.

package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// A single Broker will be created in this program. It is responsible
// for keeping a list of which clients (browsers) are currently attached
// and broadcasting events (messages) to those clients.
//
type Broker struct {

	// Create a map of clients, the keys of the map are the channels
	// over which we can push messages to attached clients.  (The values
	// are just booleans and are meaningless.)
	//
	clients map[chan string]bool

	// Channel into which new clients can be pushed
	//
	newClients chan chan string

	// Channel into which disconnected clients should be pushed
	//
	defunctClients chan chan string

	// Channel into which messages are pushed to be broadcast out
	// to attahed clients.
	//
	messages chan string
}

// Reflector listens on a port for POST requests, and reflects
// everything it gets into the broker.
//
type Reflector struct {
	broker *Broker
}

// This Broker method starts a new goroutine.  It handles the addition
// & removal of clients, as well as the broadcasting of messages out
// to clients that are currently attached.
//
func (b *Broker) Start() {
	go func() {
		for {

			// Block until we receive from one of the
			// three following channels.
			select {

			case s := <-b.newClients:

				// There is a new client attached and we
				// want to start sending them messages.
				b.clients[s] = true
				log.Println("Added new client")

			case s := <-b.defunctClients:

				// A client has dettached and we want to
				// stop sending them messages.
				delete(b.clients, s)
				close(s)

				log.Println("Removed client")

			case msg := <-b.messages:

				// There is a new message to send.  For each
				// attached client, push the new message
				// into the client's message channel.
				for s, _ := range b.clients {
					s <- msg
				}
				log.Printf("Broadcast message to %d clients", len(b.clients))
			}
		}
	}()
}

// This Broker method handles and HTTP request at the "/events/" URL.
//
func (b *Broker) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// Make sure that the writer supports flushing.
	//
	f, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	// Create a new channel, over which the broker can
	// send this client messages.
	messageChan := make(chan string)

	// Add this client to the map of those that should
	// receive updates
	b.newClients <- messageChan

	// Listen to the closing of the http connection via the CloseNotifier
	notify := w.(http.CloseNotifier).CloseNotify()
	go func() {
		<-notify
		// Remove this client from the map of attached clients
		// when `EventHandler` exits.
		b.defunctClients <- messageChan
		log.Println("HTTP connection just closed.")
	}()

	// Set the headers related to event streaming.
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Don't close the connection, instead loop 10 times,
	// sending messages and flushing the response each time
	// there is a new message to send along.
	//
	// NOTE: we could loop endlessly; however, then you
	// could not easily detect clients that dettach and the
	// server would continue to send them messages long after
	// they're gone due to the "keep-alive" header.  One of
	// the nifty aspects of SSE is that clients automatically
	// reconnect when they lose their connection.
	//
	// A better way to do this is to use the CloseNotifier
	// interface that will appear in future releases of
	// Go (this is written as of 1.0.3):
	// https://code.google.com/p/go/source/detail?name=3292433291b2
	//
	for {

		// Read from our messageChan.
		msg, open := <-messageChan

		if !open {
			// If our messageChan was closed, this means that the client has
			// disconnected.
			break
		}

		// Write to the ResponseWriter, `w`.  The "data:" and "\n\n" are part of the protocol.
		fmt.Fprintf(w, "data: %s\n\n", msg)

		// Flush the response.  This is only possible if
		// the repsonse supports streaming.
		f.Flush()
	}

	// Done.
	log.Println("Finished HTTP request at ", r.URL.Path)
}

// Handler for the main page, which we wire up to the
// route at "/" below in `main`.
//
func MainPageHandler(w http.ResponseWriter, r *http.Request) {

	// Did you know Golang's ServeMux matches only the
	// prefix of the request URL?  It's true.  Here we
	// insist the path is just "/".
	if r.URL.Path != "/" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Read in the template with our SSE JavaScript code.
	t, err := template.ParseFiles("templates/index.html")
	if err != nil {
		log.Fatal("error parsing template.")
	}

	// Render the template, writing to `w`.
	t.Execute(w, "")

	// Done.
	log.Println("Finished HTTP request at ", r.URL.Path)
}

func startSSEServer(b *Broker) {
	// Start processing events
	b.Start()

	// static images
	fs := http.FileServer(http.Dir("/images"))
	http.Handle("/images", http.StripPrefix("/images", fs))

	// Make b the HTTP handler for "/events/".  It can do
	// this because it has a ServeHTTP method.  That method
	// is called in a separate goroutine for each
	// request to "/events/".
	http.Handle("/events/", b)

	// When we get a request at "/", call `MainPageHandler`
	// in a new goroutine.
	http.Handle("/", http.HandlerFunc(MainPageHandler))

	// Start the SSE server and listen forever on port 8000.
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatalf("Failed to start broker: %v", err)
	}
}

// Listen for POST requests, read the body, and reflect it into the
// broker.
//
func (reflector *Reflector) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Error(w, "You can only POST here.", 400)
		return
	}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "Error reading body", 500)
		return
	}
	reflector.broker.messages <- string(body)
	w.WriteHeader(http.StatusOK)
}

// Start the broker and reflector
func main() {
	broker := &Broker{
		make(map[chan string]bool),
		make(chan (chan string)),
		make(chan (chan string)),
		make(chan string),
	}
	reflector := &Reflector{broker}
	go func() {
		err := http.ListenAndServe(":8001", reflector)
		if err != nil {
			log.Fatalf("Failed to start reflector: %v", err)
		}
	}()
	go startSSEServer(broker)

	for {
		time.Sleep(1 * time.Second)
	}
}
