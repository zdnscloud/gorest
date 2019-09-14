package main

import (
	"encoding/base64"
	"fmt"
	"net"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zdnscloud/cement/uuid"
	"github.com/zdnscloud/gorest"
	"github.com/zdnscloud/gorest/adaptor"
	goresterr "github.com/zdnscloud/gorest/error"
	"github.com/zdnscloud/gorest/resource"
	"github.com/zdnscloud/gorest/resource/schema"
)

var (
	version = resource.APIVersion{
		Group:   "zdns.cloud.example",
		Version: "example/v1",
	}
	clusterKind = resource.DefaultKindName(Cluster{})
	nodeKind    = resource.DefaultKindName(Node{})
)

type Cluster struct {
	resource.ResourceBase `json:",inline"`
	Name                  string `json:"name" rest:"required=true,minLen=1,maxLen=10"`
	NodeCount             int    `json:"nodeCount" rest:"required=true,min=1,max=1000"`
}

func (c Cluster) CreateActions(name string) *resource.Action {
	switch name {
	case "encode":
		return &resource.Action{
			Name:  "encode",
			Input: &Input{},
		}
	case "decode":
		return &resource.Action{
			Name:  "decode",
			Input: &Input{},
		}
	default:
		return nil
	}
}

type clusterHandler struct {
	clusters []*Cluster
}

func newClusterHandler() *clusterHandler {
	return &clusterHandler{}
}

func (h *clusterHandler) Create(ctx *resource.Context) (interface{}, *goresterr.APIError) {
	cluster := ctx.Resource.(*Cluster)
	for _, c := range h.clusters {
		if c.Name == cluster.Name {
			return nil, goresterr.NewAPIError(goresterr.DuplicateResource, fmt.Sprintf("cluster %s already exist", cluster.Name))
		}
	}
	cluster.SetID(cluster.Name)
	cluster.SetCreationTimestamp(time.Now())
	h.clusters = append(h.clusters, cluster)
	return cluster, nil
}

func (h *clusterHandler) List(ctx *resource.Context) interface{} {
	return h.clusters
}

func (h *clusterHandler) Get(ctx *resource.Context) interface{} {
	return h.getCluster(ctx.Resource.GetID())
}

func (h *clusterHandler) getCluster(name string) *Cluster {
	for _, c := range h.clusters {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func (h *clusterHandler) Action(ctx *resource.Context) (interface{}, *goresterr.APIError) {
	r := ctx.Resource
	input, _ := r.GetAction().Input.(*Input)
	switch r.GetAction().Name {
	case "encode":
		return base64.StdEncoding.EncodeToString([]byte(input.Data)), nil
	case "decode":
		if data, e := base64.StdEncoding.DecodeString(input.Data); e != nil {
			return nil, goresterr.NewAPIError(goresterr.InvalidFormat, e.Error())
		} else {
			return string(data), nil
		}
	default:
		panic("it should never come here")
	}
}

type Node struct {
	resource.ResourceBase `json:",inline"`
	Address               string `json:"address,omitempty" rest:"required=true,minLen=7,maxLen=13"`
	IsWorker              bool   `json:"isWorker"`
}

func (n Node) GetParents() []resource.ResourceKind {
	return []resource.ResourceKind{Cluster{}}
}

func (n Node) CreateDefaultResource() resource.Resource {
	return &Node{
		IsWorker: true,
	}
}

type nodeHandler struct {
	clusters *clusterHandler
	nodes    map[string][]*Node
}

func newNodeHandler(h *clusterHandler) *nodeHandler {
	return &nodeHandler{
		clusters: h,
		nodes:    make(map[string][]*Node),
	}
}

func (h *nodeHandler) Create(ctx *resource.Context) (interface{}, *goresterr.APIError) {
	node := ctx.Resource.(*Node)
	id, _ := uuid.Gen()
	node.SetID(id)
	return h.addNode(node)
}

func (h *nodeHandler) addNode(node *Node) (interface{}, *goresterr.APIError) {
	cn := node.GetParent().GetID()
	nodes, ok := h.getNodesInCluster(cn)
	if !ok {
		return nil, goresterr.NewAPIError(goresterr.NotFound, fmt.Sprintf("unknown cluster %s ", cn))
	}

	if ip := net.ParseIP(node.Address); ip == nil {
		return nil, goresterr.NewAPIError(goresterr.InvalidFormat, "address isn't valid ipv4 address")
	}

	if h.getNode(nodes, node.Address) != nil {
		return nil, goresterr.NewAPIError(goresterr.DuplicateResource, fmt.Sprintf("node %s already exist", node.Address))
	}

	h.nodes[cn] = append(nodes, node)
	return node, nil
}

func (h *nodeHandler) getNodesInCluster(name string) ([]*Node, bool) {
	if h.clusters.getCluster(name) == nil {
		return nil, false
	}
	nodes, ok := h.nodes[name]
	if ok == false {
		return nil, true
	} else {
		return nodes, true
	}
}

func (h *nodeHandler) getNode(nodes []*Node, address string) *Node {
	for _, n := range nodes {
		if n.Address == address {
			return n
		}
	}
	return nil
}

func (h *nodeHandler) Delete(ctx *resource.Context) *goresterr.APIError {
	node := ctx.Resource.(*Node)
	return h.deleteNode(node)
}

func (h *nodeHandler) deleteNode(node *Node) *goresterr.APIError {
	cn := node.GetParent().GetID()
	nodes, ok := h.getNodesInCluster(cn)
	if !ok {
		return goresterr.NewAPIError(goresterr.NotFound, fmt.Sprintf("unknown cluster %s ", cn))
	}

	id := node.GetID()
	for i, n := range nodes {
		if n.GetID() == id {
			h.nodes[cn] = append(nodes[:i], nodes[i+1:]...)
			return nil
		}
	}
	return goresterr.NewAPIError(goresterr.NotFound, fmt.Sprintf("node %s not found", id))
}

func (h *nodeHandler) Update(ctx *resource.Context) (interface{}, *goresterr.APIError) {
	node := ctx.Resource.(*Node)
	cn := node.GetParent().GetID()
	nodes, ok := h.getNodesInCluster(cn)
	if !ok {
		return nil, goresterr.NewAPIError(goresterr.NotFound, fmt.Sprintf("unknown cluster %s ", cn))
	}

	id := node.GetID()
	for i, n := range nodes {
		if n.GetID() == id {
			if n.Address == node.Address {
				nodes[i] = node
			} else {
				if _, err := h.addNode(node); err != nil {
					return nil, err
				}
				nodes = append(nodes[:i], nodes[i+1:]...)
			}
			h.nodes[cn] = nodes
			return node, nil
		}
	}
	return nil, goresterr.NewAPIError(goresterr.NotFound, fmt.Sprintf("node %s not found", id))
}

func (h *nodeHandler) List(ctx *resource.Context) interface{} {
	node := ctx.Resource.(*Node)
	cn := node.GetParent().GetID()
	nodes, _ := h.getNodesInCluster(cn)
	return nodes
}

func (h *nodeHandler) Get(ctx *resource.Context) interface{} {
	node := ctx.Resource.(*Node)
	cn := node.GetParent().GetID()
	nodes, ok := h.getNodesInCluster(cn)
	if !ok {
		return nil
	}

	id := node.GetID()
	for _, n := range nodes {
		if n.GetID() == id {
			return n
		}
	}
	return nil
}

type Input struct {
	Data string `json:"data,omitempty"`
}

func main() {
	schemas := schema.NewSchemaManager()
	ch := newClusterHandler()
	chw, _ := resource.HandlerAdaptor(ch)
	nhw, _ := resource.HandlerAdaptor(newNodeHandler(ch))
	schemas.Import(&version, Cluster{}, chw)
	schemas.Import(&version, Node{}, nhw)
	router := gin.Default()
	adaptor.RegisterHandler(router, gorest.NewAPIServer(schemas), schemas.GenerateResourceRoute())
	router.Run("0.0.0.0:1234")
}
