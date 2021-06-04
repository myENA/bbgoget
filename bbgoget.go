package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

func main() {
	fs := flag.NewFlagSet("bbgoget", flag.ExitOnError)

	var (
		listenAddr string
	)
	bbh := &BBHandler{}

	fs.StringVar(&listenAddr, "listen-address", ":8800", "specify listen address")
	fs.IntVar(&bbh.sshPort, "ssh-port", 7999, "specify git server ssh port")
	fs.StringVar(&bbh.serverNameOverride, "servername-override", "", "override the server name.  "+
		"if not specified, uses the host value from the X-Forwarded-Host header")
	fs.IntVar(&bbh.depth, "depth", 3, "specify depth of repositories")
	fs.BoolVar(&bbh.rpMode, "reverse-proxy-mode", false, "if true, reverse proxy for reverse-proxy-url")
	fs.StringVar(&bbh.rpURLString, "reverse-proxy-url", "", "if reverse-proxy-mode is true, this must be "+
		"set to the upstream url for http(s) traffic")
	fs.BoolVar(&bbh.rpIgnoreSSLErrors, "revers-proxy-ignore-ssl-errors", false, "if true, ignore "+
		"upstream tls errors in the case of a self signed certificate, for instance")

	_ = fs.Parse(os.Args[1:])

	err := bbh.Initialize()
	if err != nil {
		log.Printf("Error: %s", err)
		os.Exit(1)
	}
	server := &http.Server{
		Addr:              listenAddr,
		Handler:           bbh,
		ReadHeaderTimeout: 2 * time.Second,
		WriteTimeout:      10 * time.Second,
	}

	err = server.ListenAndServe()
	if err != nil {
		log.Printf("Error: %s", err)
		os.Exit(1)
	}
}

type BBHandler struct {
	serverNameOverride string
	sshPort            int
	depth              int
	rpMode             bool
	rpURLString        string
	rpIgnoreSSLErrors  bool
	rpURL              *url.URL
	rp                 *httputil.ReverseProxy
}

func (bbh *BBHandler) Initialize() error {
	if bbh.rpMode == true {
		if bbh.rpURLString == "" {
			return fmt.Errorf("if reverse proxy mode is enabled, the URL must be set")
		}
		var err error
		bbh.rpURL, err = url.Parse(bbh.rpURLString)
		if err != nil {
			return fmt.Errorf("invalid reverse proxy URL: %w", err)
		}
		trans := &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: bbh.rpIgnoreSSLErrors,
			},
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}
		bbh.rp = &httputil.ReverseProxy{
			Director:       bbh.Director,
			Transport:      trans,
		}
	}
	return nil
}

func (bbh *BBHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {

	if req.URL.Query().Get("go-get") != "1" {
		// This is not a go-get request. Check to see if we're in proxy mode.
		switch bbh.rpMode {
		case false:
			resp.WriteHeader(http.StatusBadRequest)
			resp.Write([]byte("not a go-get request, investigate proxy config\n"))
		case true:
			// if we get here, we need to proxy the request
			bbh.rp.ServeHTTP(resp, req)
		}
		return
	}

	pathParts := strings.Split(req.URL.Path, "/")

	if len(pathParts) < bbh.depth {
		resp.WriteHeader(http.StatusBadRequest)
		resp.Write([]byte("invalid path\n"))
		return
	}

	// Strip everything after the repository portion of the URL
	prefix := strings.Join(pathParts[:bbh.depth], "/")

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

	hostSuffix := ""
	if bbh.sshPort != 22 {
		hostSuffix = fmt.Sprintf(":%d", bbh.sshPort)
	}
	repoURL := fmt.Sprintf("ssh://git@%s%s%s.git", host, hostSuffix, prefix)
	importPrefix := host + prefix

	resp.WriteHeader(http.StatusOK)
	w := bufio.NewWriter(resp)
	_, err := w.WriteString(fmt.Sprintf("<html><head><meta name=\"go-import\" content=\"%s %s %s\"></head><body>go get</body></html>", importPrefix, "git", repoURL))
	if err != nil {
		fmt.Printf("got error writing body: %s\n", err)
	}
	_ = w.Flush()
}

func (bbh BBHandler) Director(req *http.Request) {
	req.URL.Host = bbh.rpURL.Host
	req.URL.Scheme = bbh.rpURL.Scheme
	if len(bbh.rpURL.Path) > 0 {
		req.URL.Path = path.Join(bbh.rpURL.Path, req.URL.Path)
	}
}

func splitHostPort(hp string) (host string, port string) {
	vals := strings.Split(hp, ":")
	if len(vals) == 1 {
		return vals[0], ""
	}
	return vals[0], vals[1]
}
