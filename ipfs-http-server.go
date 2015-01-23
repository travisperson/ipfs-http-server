package main

import (
	"fmt"
	"mime"
	"strings"
	"code.google.com/p/go.net/context"
	core "github.com/jbenet/go-ipfs/core"
	coreunix "github.com/jbenet/go-ipfs/core/coreunix"
	fsrepo "github.com/jbenet/go-ipfs/repo/fsrepo"
	"io"
	"net/url"
	"net/http"
	"net/http/httputil"
	"os/user"
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

	io.Copy(w, reader)
}

func main() {

	// Check to see if the daemon is running
	// Hopefully there will be a better way to do this with
	// HEAD at some point
	resp, err := http.Get("http://127.0.0.1:5001/ipfs/")

	if err != nil {
		panic(err)
	}

	if resp.StatusCode == 400 {
		remote, err := url.Parse("http://127.0.0.1:5001")

		if err != nil {
			panic(err)
		}
		proxy := httputil.NewSingleHostReverseProxy(remote)
		http.Handle("/ipfs/", proxy)
	} else {
		ipfs := IPFSHandler{}

		usr, err := user.Current()
		if err != nil {
			panic(err)
		}

		ipfs.Init(usr.HomeDir + "/.go-ipfs")

		http.HandleFunc("/ipfs/", ipfs.Get)
	}

	fmt.Println(resp)


	http.Handle("/", http.FileServer(http.Dir(".")))

	http.ListenAndServe(":8080", nil)
}

func doStuff(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hello")
}
