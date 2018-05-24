package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	fs := flag.NewFlagSet("bbgoget", flag.ExitOnError)

	var (
		sshPort            int
		listenAddr         string
		serverNameOverride string
	)

	fs.StringVar(&listenAddr, "listen-address", ":8800", "specify listen address")
	fs.IntVar(&sshPort, "ssh-port", 7999, "specify bitbucket ssh port")
	fs.StringVar(&serverNameOverride, "servername-override", "", "override the server name.  if not specified, uses the host value from the X-Forwarded-Host header")

	_ = fs.Parse(os.Args[1:])

	bbh := &BBHandler{
		sshPort:            sshPort,
		serverNameOverride: serverNameOverride,
	}
	server := &http.Server{
		Addr:              listenAddr,
		Handler:           bbh,
		ReadHeaderTimeout: 2 * time.Second,
		WriteTimeout:      10 * time.Second,
	}

	err := server.ListenAndServe()
	if err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}

type BBHandler struct {
	serverNameOverride string
	sshPort            int
}

func (bbh *BBHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {

	if req.URL.Query().Get("go-get") != "1" {
		resp.WriteHeader(http.StatusBadRequest)
		resp.Write([]byte("not a go-get request, investigate proxy config\n"))
		return
	}

	// Since path should begin with '/', strings.split
	// will return the first element as blank, second
	// as project, third as our actual repository.
	pathParts := strings.Split(req.URL.Path, "/")

	if len(pathParts) < 3 {
		resp.WriteHeader(http.StatusBadRequest)
		resp.Write([]byte("invalid path\n"))
		return
	}

	// Strip everything after the repository portion of the URL
	prefix := strings.Join(pathParts[:3], "/")

	// Try to find the host
	host := bbh.serverNameOverride
	if len(host) == 0 {
		host = req.Header.Get("X-Forwarded-Host")
		if len(host) == 0 {
			host, _ = splitHostPort(req.URL.Host)
		}
	}

	if len(host) == 0 {
		fmt.Printf("error with blank host")
		resp.WriteHeader(http.StatusBadRequest)
		resp.Write([]byte("blank host"))
	}

	repoURL := fmt.Sprintf("ssh://git@%s:%d%s.git", host, bbh.sshPort, prefix)
	importPrefix := host + prefix

	resp.WriteHeader(http.StatusOK)
	w := bufio.NewWriter(resp)
	_, err := w.WriteString(fmt.Sprintf("<html><head><meta name=\"go-import\" content=\"%s %s %s\"></head><body>go get</body></html>", importPrefix, "git", repoURL))
	if err != nil {
		fmt.Printf("got error writing body: %s\n", err)
	}
	_ = w.Flush()
}

func splitHostPort(hp string) (host string, port string) {
	vals := strings.Split(hp, ":")
	if len(vals) == 1 {
		return vals[0], ""
	}
	return vals[0], vals[1]
}
