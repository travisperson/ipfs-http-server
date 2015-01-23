package main

import (
	"code.google.com/p/go.net/context"
	core "github.com/jbenet/go-ipfs/core"
	coreunix "github.com/jbenet/go-ipfs/core/coreunix"
	fsrepo "github.com/jbenet/go-ipfs/repo/fsrepo"
	u "github.com/jbenet/go-ipfs/util"
	"io"
	"net/http"
)

type IPFSHandler struct {
	repo *fsrepo.FSRepo
	node *core.IpfsNode
}


func (p *IPFSHandler) Init(repo string) {
	p.repo = fsrepo.At("/home/travis/.go-ipfs")
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
	reader, err := coreunix.Cat(p.node, u.B58KeyDecode(blk))

	if err != nil {
		panic(err)
	}

	io.Copy(w, reader)
}
func main() {

	ipfs := IPFSHandler{}

	ipfs.Init("/home/travis/.go-ipfs")

	//r := mux.NewRouter()
	//r.PathPrefix("/ipfs/{hash}").HandlerFunc(ipfs.Get)
	//r.PathPrefix("/").Handler(http.FileServer(http.Dir(".")))
	http.HandleFunc("/ipfs/", ipfs.Get)
	http.Handle("/", http.FileServer(http.Dir(".")))


	http.ListenAndServe(":8080", nil)
}

func doStuff(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hello")
}
