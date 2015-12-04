package mfs

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	context "github.com/ipfs/go-ipfs/Godeps/_workspace/src/golang.org/x/net/context"

	dag "github.com/ipfs/go-ipfs/merkledag"
	ft "github.com/ipfs/go-ipfs/unixfs"
	ufspb "github.com/ipfs/go-ipfs/unixfs/pb"
)

var ErrNotYetImplemented = errors.New("not yet implemented")
var ErrInvalidChild = errors.New("invalid child node")
var ErrDirExists = errors.New("directory already has entry by that name")

type Directory struct {
	dserv  dag.DAGService
	parent childCloser

	childDirs map[string]*Directory
	files     map[string]*File

	lock sync.Mutex
	node *dag.Node
	ctx  context.Context

	modTime time.Time

	name string
}

func NewDirectory(ctx context.Context, name string, node *dag.Node, parent childCloser, dserv dag.DAGService) *Directory {
	return &Directory{
		dserv:     dserv,
		ctx:       ctx,
		name:      name,
		node:      node,
		parent:    parent,
		childDirs: make(map[string]*Directory),
		files:     make(map[string]*File),
		modTime:   time.Now(),
	}
}

// closeChild updates the child by the given name to the dag node 'nd'
// and changes its own dag node, then propogates the changes upward
func (d *Directory) closeChild(name string, nd *dag.Node) error {
	_, err := d.dserv.Add(nd)
	if err != nil {
		return err
	}

	d.lock.Lock()
	defer d.lock.Unlock()
	err = d.updateChild(name, nd)
	if err != nil {
		return err
	}

	return d.parent.closeChild(d.name, d.node)
}

func (d *Directory) updateChild(name string, nd *dag.Node) error {
	err := d.node.RemoveNodeLink(name)
	if err != nil && err != dag.ErrNotFound {
		return err
	}

	err = d.node.AddNodeLinkClean(name, nd)
	if err != nil {
		return err
	}

	d.modTime = time.Now()

	return nil
}

func (d *Directory) Type() NodeType {
	return TDir
}

// childFile returns a file under this directory by the given name if it exists
func (d *Directory) childFile(name string) (*File, error) {
	fi, ok := d.files[name]
	if ok {
		return fi, nil
	}

	fsn, err := d.childNode(name)
	if err != nil {
		return nil, err
	}

	if fi, ok := fsn.(*File); ok {
		return fi, nil
	}

	return nil, fmt.Errorf("%s is not a file", name)
}

// childDir returns a directory under this directory by the given name if it
// exists.
func (d *Directory) childDir(name string) (*Directory, error) {
	dir, ok := d.childDirs[name]
	if ok {
		return dir, nil
	}

	fsn, err := d.childNode(name)
	if err != nil {
		return nil, err
	}

	if dir, ok := fsn.(*Directory); ok {
		return dir, nil
	}

	return nil, fmt.Errorf("%s is not a directory", name)
}

// childNode returns a FSNode under this directory by the given name if it exists.
// it does *not* check the cached dirs and files
func (d *Directory) childNode(name string) (FSNode, error) {
	nd, err := d.childFromDag(name)
	if err != nil {
		return nil, err
	}

	i, err := ft.FromBytes(nd.Data)
	if err != nil {
		return nil, err
	}

	switch i.GetType() {
	case ufspb.Data_Directory:
		ndir := NewDirectory(d.ctx, name, nd, d, d.dserv)
		d.childDirs[name] = ndir
		return ndir, nil
	case ufspb.Data_File:
		nfi, err := NewFile(name, nd, d, d.dserv)
		if err != nil {
			return nil, err
		}
		d.files[name] = nfi
		return nfi, nil
	case ufspb.Data_Metadata:
		return nil, ErrNotYetImplemented
	default:
		return nil, ErrInvalidChild
	}
}

// childFromDag searches through this directories dag node for a child link
// with the given name
func (d *Directory) childFromDag(name string) (*dag.Node, error) {
	for _, lnk := range d.node.Links {
		if lnk.Name == name {
			return lnk.GetNode(d.ctx, d.dserv)
		}
	}

	return nil, os.ErrNotExist
}

// Child returns the child of this directory by the given name
func (d *Directory) Child(name string) (FSNode, error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	return d.childUnsync(name)
}

// childUnsync returns the child under this directory by the given name
// without locking, useful for operations which already hold a lock
func (d *Directory) childUnsync(name string) (FSNode, error) {
	cdir, ok := d.childDirs[name]
	if ok {
		return cdir, nil
	}

	cfile, ok := d.files[name]
	if ok {
		return cfile, nil
	}

	return d.childNode(name)
}

type NodeListing struct {
	Name string
	Type int
	Size int64
	Hash string
}

func (d *Directory) List() ([]NodeListing, error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	var out []NodeListing
	for _, l := range d.node.Links {
		child := NodeListing{}
		child.Name = l.Name

		c, err := d.childUnsync(l.Name)
		if err != nil {
			return nil, err
		}

		child.Type = int(c.Type())
		if c, ok := c.(*File); ok {
			size, err := c.Size()
			if err != nil {
				return nil, err
			}
			child.Size = size
		}
		nd, err := c.GetNode()
		if err != nil {
			return nil, err
		}

		k, err := nd.Key()
		if err != nil {
			return nil, err
		}

		child.Hash = k.B58String()

		out = append(out, child)
	}

	return out, nil
}

func (d *Directory) Mkdir(name string) (*Directory, error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	_, err := d.childDir(name)
	if err == nil {
		return nil, os.ErrExist
	}
	_, err = d.childFile(name)
	if err == nil {
		return nil, os.ErrExist
	}

	ndir := &dag.Node{Data: ft.FolderPBData()}

	_, err = d.dserv.Add(ndir)
	if err != nil {
		return nil, err
	}

	err = d.node.AddNodeLinkClean(name, ndir)
	if err != nil {
		return nil, err
	}

	err = d.parent.closeChild(d.name, d.node)
	if err != nil {
		return nil, err
	}

	return d.childDir(name)
}

func (d *Directory) Unlink(name string) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	delete(d.childDirs, name)
	delete(d.files, name)

	err := d.node.RemoveNodeLink(name)
	if err != nil {
		return err
	}

	return d.parent.closeChild(d.name, d.node)
}

// AddChild adds the node 'nd' under this directory giving it the name 'name'
func (d *Directory) AddChild(name string, nd *dag.Node) error {
	d.Lock()
	defer d.Unlock()

	_, err := d.childUnsync(name)
	if err == nil {
		return ErrDirExists
	}

	_, err = d.dserv.Add(nd)
	if err != nil {
		return err
	}

	err = d.node.AddNodeLinkClean(name, nd)
	if err != nil {
		return err
	}

	d.modTime = time.Now()

	return d.parent.closeChild(d.name, d.node)
}

func (d *Directory) sync() error {
	for name, dir := range d.childDirs {
		nd, err := dir.GetNode()
		if err != nil {
			return err
		}

		_, err = d.dserv.Add(nd)
		if err != nil {
			return err
		}

		err = d.updateChild(name, nd)
		if err != nil {
			return err
		}
	}

	for name, file := range d.files {
		nd, err := file.GetNode()
		if err != nil {
			return err
		}

		_, err = d.dserv.Add(nd)
		if err != nil {
			return err
		}

		err = d.updateChild(name, nd)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *Directory) GetNode() (*dag.Node, error) {
	d.Lock()
	defer d.Unlock()

	err := d.sync()
	if err != nil {
		return nil, err
	}

	return d.node, nil
}

func (d *Directory) Lock() {
	d.lock.Lock()
}

func (d *Directory) Unlock() {
	d.lock.Unlock()
}
