package main

import (
	"fmt"
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
	blk := r.URL.Path[6:]

	if len(blk) < 3 {
		w.WriteHeader(403)
		io.WriteString(w, "Hash must be longer than 3 bytes")
		return
	}

	fmt.Println(blk)
	
	reader, err := coreunix.Cat(p.node, blk)

	if err != nil {
		w.WriteHeader(404)
		io.WriteString(w, "Failed to retrieve: " + blk)
		return
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
