package main

import (
	"fmt"
	"mime"
	"strings"
	"code.google.com/p/go.net/context"
	core "github.com/jbenet/go-ipfs/core"
	coreunix "github.com/jbenet/go-ipfs/core/coreunix"
	fsrepo "github.com/jbenet/go-ipfs/repo/fsrepo"
	"net"
	"io"
	"os"
	"net/url"
	"net/http"
	"net/http/httputil"
	"os/user"
	"flag"
)
type IPFSHandler struct {
	repo *fsrepo.FSRepo
	node *core.IpfsNode
}


func (p *IPFSHandler) Init(repo string) {
	p.repo = fsrepo.At(repo)
	err := p.repo.Open()
	if err != nil {
		panic(err)
	}

	p.node, err = core.NewIPFSNode(context.Background(), core.Online(p.repo))
	if err != nil {
		panic(err)
	}
}

func (p *IPFSHandler) Get(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[6:]
	fmt.Println(path)

	if len(path) < 3 {
		w.WriteHeader(403)
		io.WriteString(w, "Hash must be longer than 3 bytes")
		return
	}

	reader, err := coreunix.Cat(p.node, path)
	if err != nil {
		w.WriteHeader(404)
		io.WriteString(w, "Failed to retrieve: " + path)
		return
	}

	extensionIndex := strings.LastIndex(path, ".")
	if extensionIndex != -1 {
		extension := path[extensionIndex:]
		mimeType := mime.TypeByExtension(extension)
		if len(mimeType) > 0 {
			w.Header().Add("Content-Type", mimeType)
		}
	}

	io.Copy(os.Stdout, reader)

}

func main() {

	flag_port  := flag.Int("port", 8080, "Port to attempt to listen on")
	flag_proxy := flag.String("proxy", "", "Reverse proxy / to a different server instead of serving the current directory")
	
	flag.Parse()

	ipfs := IPFSHandler{}

	usr, err := user.Current()
	if err != nil {
		panic(err)
	}

	ipfs_repo := usr.HomeDir + "/.go-ipfs"

	// Check to see if the repo is locked
	repoLocked := fsrepo.LockedByOtherProcess(ipfs_repo)

	if repoLocked {
		// Most likely the daemon is running
		remote, err := url.Parse("http://127.0.0.1:8080")

		if err != nil {
			panic(err)
		}

		proxy := httputil.NewSingleHostReverseProxy(remote)
		http.Handle("/ipfs/", proxy)
	} else {
		ipfs.Init(ipfs_repo)
		http.HandleFunc("/ipfs/", ipfs.Get)
	}

	if len(*flag_proxy) > 0 {
		host, port, err := net.SplitHostPort(*flag_proxy)
	
		var remote *url.URL
		if err != nil {
			remote, err = url.Parse(*flag_proxy)
			if err != nil {
				panic(err)
			}
		} else {
			remote = new(url.URL)
			remote.Host = host+":"+port
			remote.Scheme = "http"
		}

		proxy := httputil.NewSingleHostReverseProxy(remote)
		http.Handle("/", proxy)
	} else {
		http.Handle("/", http.FileServer(http.Dir(".")))
	}

	addr := &net.TCPAddr{net.IPv4(127,0,0,1), *flag_port,""}

	for  {
		_, err := net.Dial("tcp", addr.String())
		if err == nil {
			addr.Port++
		} else {
			fmt.Printf("Starting ipfs-http-server on http://%s\n", addr.String())
			err = http.ListenAndServe(addr.String(), nil)
			
			if err != nil {
				fmt.Printf("Error: ", err)
			}
		}
	}
}

func doStuff(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hello")
}
