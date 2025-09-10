package approvalserver

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	accessv1alpha1 "antware.xyz/jitaccess/api/v1alpha1"

	"github.com/coreos/go-oidc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
)

var jitGVR = schema.GroupVersionResource{
	Group:    "access.antware.xyz",
	Version:  "v1alpha1",
	Resource: "jitaccessrequests",
}
var clusterjitGVR = schema.GroupVersionResource{
	Group:    "access.antware.xyz",
	Version:  "v1alpha1",
	Resource: "clusterjitaccessrequests",
}

type Server struct {
	client dynamic.Interface
	ns     string
	config ServerConfig
}

type ServerConfig struct {
	clientId string
	issuer   string
}

func NewServer(client dynamic.Interface, namespace string) *Server {
	return &Server{
		client: client,
		ns:     namespace,
		config: ServerConfig{
			clientId: "your-client-id",
			issuer:   "https://accounts.google.com",
		},
	}
}

func (s *Server) Start(addr string) error {
	http.Handle("/cluster/requests", s.authMiddleware(
		http.HandlerFunc(s.listClusterRequests),
	))

	http.Handle("/cluster/approve", s.authMiddleware(
		http.HandlerFunc(s.approveClusterRequest),
	))

	http.Handle("/cluster/deny", s.authMiddleware(
		http.HandlerFunc(s.denyClusterRequest),
	))

	http.Handle("/requests", s.authMiddleware(
		http.HandlerFunc(s.listRequests),
	))

	http.Handle("/approve", s.authMiddleware(
		http.HandlerFunc(s.approveRequest),
	))

	http.Handle("/deny", s.authMiddleware(
		http.HandlerFunc(s.denyRequest),
	))

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := fmt.Fprint(w, "ok")

		if err != nil {
			log.Printf("Error: %s\n", err)
		}
	})

	http.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := fmt.Fprint(w, "ready")

		if err != nil {
			log.Printf("Error: %s\n", err)
		}
	})

	log.Printf("Approval server listening on %s\n", addr)
	return http.ListenAndServe(addr, nil)
}

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractBearerToken(r)
		if token == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		provider, err := oidc.NewProvider(context.Background(), s.config.issuer)
		if err != nil {
			http.Error(w, "OIDC error", http.StatusInternalServerError)
			return
		}

		verifier := provider.Verifier(&oidc.Config{ClientID: s.config.clientId})
		_, err = verifier.Verify(context.Background(), token)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func extractBearerToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}
	return parts[1]
}

func (s *Server) listRequests(w http.ResponseWriter, r *http.Request) {
	reqs, err := s.client.Resource(jitGVR).Namespace(metav1.NamespaceAll).
		List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	typedReqs := &accessv1alpha1.JITAccessRequestList{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(
		reqs.UnstructuredContent(), typedReqs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	{
		_, err := fmt.Fprintf(w, "<h1>Pending Requests</h1>")
		if err != nil {
			log.Printf("Error: %s\n", err)
		}
	}

	for _, item := range typedReqs.Items {
		if item.Status.State == "Pending" {
			name := item.GetName()
			justification := item.Spec.Justification

			_, err := fmt.Fprintf(w,
				`<p>
					%s - <a href="/approve?name=%s">Approve</a> | <a href="/deny?name=%s">Deny</a><br />
					Justification provided: %s
				</p>`,
				name, name, name, justification)

			if err != nil {
				log.Printf("Error: %s\n", err)
			}
		}
	}
}

func (s *Server) listClusterRequests(w http.ResponseWriter, r *http.Request) {
	reqs, err := s.client.Resource(clusterjitGVR).Namespace(metav1.NamespaceNone).
		List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	typedReqs := &accessv1alpha1.ClusterJITAccessRequestList{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(
		reqs.UnstructuredContent(), typedReqs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	{
		_, err := fmt.Fprintf(w, "<h1>Pending Requests</h1>")
		if err != nil {
			log.Printf("Error: %s\n", err)
		}
	}

	for _, item := range typedReqs.Items {
		if item.Status.State == "Pending" {
			name := item.GetName()
			justification := item.Spec.Justification

			_, err := fmt.Fprintf(w,
				`<p>
					%s - <a href="/cluster/approve?name=%s">Approve</a> | <a href="/cluster/deny?name=%s">Deny</a><br />
					Justification provided: %s
				</p>`,
				name, name, name, justification)

			if err != nil {
				log.Printf("Error: %s\n", err)
			}
		}
	}
}

func (s *Server) approveRequest(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")
	patch := []byte(`{"status":{"state":"Approved"}}`)

	_, err := s.client.Resource(jitGVR).Namespace(namespace).
		Patch(context.TODO(), name, types.MergePatchType, patch, metav1.PatchOptions{}, "status")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/requests", http.StatusSeeOther)
}

func (s *Server) denyRequest(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")
	patch := []byte(`{"status":{"state":"Denied"}}`)

	_, err := s.client.Resource(jitGVR).Namespace(namespace).
		Patch(context.TODO(), name, types.MergePatchType, patch, metav1.PatchOptions{}, "status")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/requests", http.StatusSeeOther)
}

func (s *Server) approveClusterRequest(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	patch := []byte(`{"status":{"state":"Approved"}}`)

	_, err := s.client.Resource(clusterjitGVR).Namespace(metav1.NamespaceNone).
		Patch(context.TODO(), name, types.MergePatchType, patch, metav1.PatchOptions{}, "status")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/requests", http.StatusSeeOther)
}

func (s *Server) denyClusterRequest(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	patch := []byte(`{"status":{"state":"Denied"}}`)

	_, err := s.client.Resource(clusterjitGVR).Namespace(metav1.NamespaceNone).
		Patch(context.TODO(), name, types.MergePatchType, patch, metav1.PatchOptions{}, "status")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/requests", http.StatusSeeOther)
}
