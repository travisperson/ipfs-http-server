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
	"net/http"
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

	ipfs := IPFSHandler{}
	
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}

	ipfs.Init(usr.HomeDir + "/.go-ipfs")

	http.HandleFunc("/ipfs/", ipfs.Get)
	http.Handle("/", http.FileServer(http.Dir(".")))

	http.ListenAndServe(":8080", nil)
}

func doStuff(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hello")
}
