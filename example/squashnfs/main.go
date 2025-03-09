package main

import (
	"fmt"
	"net"
	"os"

	"github.com/go-git/go-billy/v5"
	squashfs "github.com/willscott/go-nfs/helpers/squashfs"

	nfs "github.com/willscott/go-nfs"
	nfshelper "github.com/willscott/go-nfs/helpers"
)

// ROFS is an intercepter for the filesystem indicating it should
// be read only. The undelrying billy.Memfs indicates it supports
// writing, but does not in implement billy.Change to support
// modification of permissions / modTimes, and as such cannot be
// used as RW system.
type ROFS struct {
	billy.Filesystem
}

// Capabilities exports the filesystem as readonly
func (ROFS) Capabilities() billy.Capability {
	return billy.ReadCapability | billy.SeekCapability
}

func main() {
	port := ""
	if len(os.Args) < 2 {
		fmt.Printf("Usage: squashnfs </path/to/squashfs> [port]\n")
		return
	} else if len(os.Args) == 3 {
		port = os.Args[2]
	}

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Printf("Failed to listen: %v\n", err)
		return
	}
	fmt.Printf("squashnfs server running at %s\n", listener.Addr())

	f, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Printf("Failed to open: %v\n", err)
		return
	}
	bfs := squashfs.New(f)

	handler := nfshelper.NewNullAuthHandler(ROFS{bfs})
	cacheHelper := nfshelper.NewCachingHandler(handler, 1024)
	fmt.Printf("%v", nfs.Serve(listener, cacheHelper))
}
