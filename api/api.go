package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"github.com/hustcat/docker-graph-driver/driver"
)

const (
	defaultContentTypeV1          = "appplication/vnd.docker.plugins.v1+json"
	defaultImplementationManifest = `{"Implements": ["GraphDriver"]}`
	pluginSpecDir                 = "/etc/docker/plugins"
	pluginSockDir                 = "/run/docker/plugins"

	activatePath = "/Plugin.Activate"
	createPath   = "/GraphDriver.Create"
	removePath   = "/GraphDriver.Remove"
	getPath      = "/GraphDriver.Get"
	putPath      = "/GraphDriver.Put"
	existsPath   = "/GraphDriver.Exists"
	statusPath   = "/GraphDriver.Status"
	cleanupPath  = "/GraphDriver.Cleanup"
)

// Request is the structure that docker's requests are deserialized to.
type graphDriverRequest struct {
	ID         string `json:",omitempty"`
	Parent     string `json:",omitempty"`
	MountLabel string `json:",omitempty"`
}

// Response is the strucutre that the plugin's responses are serialized to.
type graphDriverResponse struct {
	Err    error       `json:",omitempty"`
	Dir    string      `json:",omitempty"`
	Exists bool        `json:",omitempty"`
	Status [][2]string `json:",omitempty"`
}

type graphEventsCounter struct {
	activations int
	creations   int
	removals    int
	gets        int
	puts        int
	stats       int
	cleanups    int
	exists      int
}

// Handler forwards requests and responses between the docker daemon and the plugin.
type Handler struct {
	driver graphdriver.Driver
	ec     *graphEventsCounter
	mux    *http.ServeMux
}

// NewHandler initializes the request handler with a driver implementation.
func NewHandler(driver graphdriver.Driver) *Handler {
	h := &Handler{driver, &graphEventsCounter{}, http.NewServeMux()}
	h.initMux()
	return h
}

func (h *Handler) initMux() {
	h.mux.HandleFunc(activatePath, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", defaultContentTypeV1)
		fmt.Fprintln(w, defaultImplementationManifest)
	})

	h.mux.HandleFunc(createPath, func(w http.ResponseWriter, r *http.Request) {
		h.ec.creations++

		var req graphDriverRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), 500)
		}

		if err := h.driver.Create(req.ID, req.Parent); err != nil {
			http.Error(w, err.Error(), 500)
		}

		w.Header().Set("Content-Type", "appplication/vnd.docker.plugins.v1+json")
		fmt.Fprintln(w, `{}`)
	})

	h.mux.HandleFunc("/GraphDriver.Remove", func(w http.ResponseWriter, r *http.Request) {
		h.ec.removals++

		var req graphDriverRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), 500)
		}

		if err := h.driver.Remove(req.ID); err != nil {
			http.Error(w, err.Error(), 500)
		}

		w.Header().Set("Content-Type", "appplication/vnd.docker.plugins.v1+json")
		fmt.Fprintln(w, `{}`)
	})

	h.mux.HandleFunc("/GraphDriver.Get", func(w http.ResponseWriter, r *http.Request) {
		h.ec.gets++

		var req graphDriverRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), 500)
		}

		dir, err := h.driver.Get(req.ID, req.MountLabel)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}

		w.Header().Set("Content-Type", "appplication/vnd.docker.plugins.v1+json")
		fmt.Fprintln(w, fmt.Sprintf(`{"Dir": "%s"}`, dir))
	})

	h.mux.HandleFunc("/GraphDriver.Put", func(w http.ResponseWriter, r *http.Request) {
		h.ec.puts++

		var req graphDriverRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), 500)
		}

		if err := h.driver.Put(req.ID); err != nil {
			http.Error(w, err.Error(), 500)
		}

		w.Header().Set("Content-Type", "appplication/vnd.docker.plugins.v1+json")
		fmt.Fprintln(w, `{}`)
	})

	h.mux.HandleFunc("/GraphDriver.Exists", func(w http.ResponseWriter, r *http.Request) {
		h.ec.exists++

		var req graphDriverRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), 500)
		}

		w.Header().Set("Content-Type", "appplication/vnd.docker.plugins.v1+json")
		fmt.Fprintln(w, fmt.Sprintf(`{"Exists": %b}`, h.driver.Exists(req.ID)))
	})

	h.mux.HandleFunc("/GraphDriver.Status", func(w http.ResponseWriter, r *http.Request) {
		h.ec.stats++

		w.Header().Set("Content-Type", "appplication/vnd.docker.plugins.v1+json")
		fmt.Fprintln(w, `{}`)
	})

	h.mux.HandleFunc("/GraphDriver.Cleanup", func(w http.ResponseWriter, r *http.Request) {
		h.ec.cleanups++

		if err := h.driver.Cleanup(); err != nil {
			http.Error(w, err.Error(), 500)
		}

		w.Header().Set("Content-Type", "appplication/vnd.docker.plugins.v1+json")
		fmt.Fprintln(w, `{}`)
	})
}

// ServeTCP makes the handler to listen for request in a given TCP address.
// It also writes the spec file on the right directory for docker to read.
func (h *Handler) ServeTCP(pluginName, addr string) error {
	return h.listenAndServe("tcp", addr, pluginName)
}

// ServeUnix makes the handler to listen for requests in a unix socket.
// It also creates the socket file on the right directory for docker to read.
func (h *Handler) ServeUnix(systemGroup, addr string) error {
	return h.listenAndServe("unix", addr, systemGroup)
}

func (h *Handler) listenAndServe(proto, addr, group string) error {
	server := http.Server{
		Addr:    addr,
		Handler: h.mux,
	}

	start := make(chan struct{})

	var l net.Listener
	var err error
	switch proto {
	case "tcp":
		l, err = newTCPSocket(addr, nil, start)
		if err == nil {
			err = writeSpec(group, l.Addr().String())
		}
	case "unix":
		var s string
		s, err = fullSocketAddr(addr)
		if err == nil {
			l, err = newUnixSocket(s, group, start)
		}
	}
	if err != nil {
		return err
	}

	close(start)
	return server.Serve(l)
}

func writeSpec(name, addr string) error {
	if err := os.MkdirAll(pluginSpecDir, 0755); err != nil {
		return err
	}

	spec := filepath.Join(pluginSpecDir, name+".spec")
	url := "tcp://" + addr
	return ioutil.WriteFile(spec, []byte(url), 0644)
}

func fullSocketAddr(addr string) (string, error) {
	if err := os.MkdirAll(pluginSockDir, 0755); err != nil {
		return "", err
	}

	if filepath.IsAbs(addr) {
		return addr, nil
	}

	return filepath.Join(pluginSockDir, addr+".sock"), nil
}
