package approvalserver

import (
	"context"
	"fmt"
	"log"
	"net/http"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
)

var jitGVR = schema.GroupVersionResource{
	Group:    "access.antware.xyz",
	Version:  "v1alpha1",
	Resource: "jitaccessrequests",
}

type Server struct {
	client dynamic.Interface
	ns     string
}

func NewServer(client dynamic.Interface, namespace string) *Server {
	return &Server{
		client: client,
		ns:     namespace,
	}
}

func (s *Server) Start(addr string) error {
	http.HandleFunc("/requests", s.listRequests)
	http.HandleFunc("/approve", s.approveRequest)
	http.HandleFunc("/deny", s.denyRequest)

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
	})

	http.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ready")
	})

	log.Printf("Approval server listening on %s\n", addr)
	return http.ListenAndServe(addr, nil)
}
func (s *Server) listRequests(w http.ResponseWriter, r *http.Request) {
	reqs, err := s.client.Resource(jitGVR).Namespace(s.ns).
		List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "<h1>Pending Requests</h1>")
	for _, item := range reqs.Items {
		state, _, _ := unstructured.NestedString(item.Object, "status", "state")
		if state == "Pending" {
			name := item.GetName()
			fmt.Fprintf(w,
				`<p>%s - <a href="/approve?name=%s">Approve</a> | <a href="/deny?name=%s">Deny</a></p>`,
				name, name, name)
		}
	}
}

func (s *Server) approveRequest(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	patch := []byte(`{"status":{"state":"Approved"}}`)

	_, err := s.client.Resource(jitGVR).Namespace(s.ns).
		Patch(context.TODO(), name, types.MergePatchType, patch, metav1.PatchOptions{}, "status")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/requests", http.StatusSeeOther)
}

func (s *Server) denyRequest(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	patch := []byte(`{"status":{"state":"Denied"}}`)

	_, err := s.client.Resource(jitGVR).Namespace(s.ns).
		Patch(context.TODO(), name, types.MergePatchType, patch, metav1.PatchOptions{}, "status")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/requests", http.StatusSeeOther)
}
