package rest

import (
	"context"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"bequant-tt/cryptocompare/service"

	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
)

type Config struct {
	Address string
}

type Rest struct {
	mux     *mux.Router
	s       *service.Service
	config  *Config
	decoder *schema.Decoder

	mu       sync.Mutex
	stopChan chan error
	listen   bool
}

const (
	ContentType     = "Content-Type"
	ApplicationJSON = "application/json"
)

func New(service *service.Service, config *Config) *Rest {
	rest := &Rest{
		s:       service,
		config:  config,
		mu:      sync.Mutex{},
		decoder: schema.NewDecoder(),
	}

	rest.decoder.IgnoreUnknownKeys(true)

	api := mux.NewRouter()

	// Websocket
	api.HandleFunc("/ws", rest.websocketHandler)
	// Simple handler
	api.HandleFunc("/price", rest.websocketHandler)

	rest.mux = api

	rest.setupMiddleware()

	return rest
}

// Run is function need to implement runner.Runner interface
// same with Listen() method
func (rest *Rest) Run() error {
	return rest.Listen()
}

func (rest *Rest) Listen() (err error) {
	listener, err := net.Listen("tcp", rest.config.Address)
	if err != nil {
		return err
	}

	r := http.NewServeMux()
	r.Handle("/", rest.mux)

	server := &http.Server{Handler: r}
	rest.stopChan = make(chan error)

	rest.mu.Lock()
	rest.listen = true
	rest.mu.Unlock()

	go func() {
		defer close(rest.stopChan)
		e := <-rest.stopChan

		rest.mu.Lock()
		rest.listen = false
		rest.mu.Unlock()

		if e != nil {
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()

		rest.stopChan <- server.Shutdown(ctx)
	}()

	err = server.Serve(listener)

	// If err == http.ErrServerClosed "shutdown" goroutine is executed and will be freed
	// in other case we must send error to it to prevent goroutine leak
	if err != http.ErrServerClosed {
		rest.stopChan <- err
	}

	return err
}

func (rest *Rest) Stop() error {
	rest.mu.Lock()
	listen := rest.listen
	rest.mu.Unlock()

	if !listen {
		return nil
	}

	rest.stopChan <- nil

	return <-rest.stopChan
}

func (rest *Rest) setupMiddleware() {
	rest.mux.Use(func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := r.ParseForm()
			if err != nil {
				log.Println(err)
			}

			vars := mux.Vars(r)
			for key, value := range vars {
				r.Form.Add(key, value)
				r.PostForm.Add(key, value)
			}

			handler.ServeHTTP(w, r)
		})
	})
}
