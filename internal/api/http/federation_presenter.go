package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/qoppa-tech/toy-gitfed/internal/modules/federation"
	apperrors "github.com/qoppa-tech/toy-gitfed/pkg/errors"
)

type FederationInstanceConfig struct {
	InstanceURL  string
	InstanceName string
	AdminEmail   string
	Capabilities []string
}

type FederationPresenter struct {
	service  *federation.Service
	instance FederationInstanceConfig
}

func NewFederationPresenter(service *federation.Service, instance FederationInstanceConfig) *FederationPresenter {
	if len(instance.Capabilities) == 0 {
		instance.Capabilities = []string{"discover", "verify", "peers"}
	}
	return &FederationPresenter{service: service, instance: instance}
}

func (p *FederationPresenter) RegisterRoutes(mux *http.ServeMux, authMw func(http.Handler) http.Handler) {
	if authMw == nil {
		authMw = func(next http.Handler) http.Handler { return next }
	}

	mux.HandleFunc("GET /.well-known/gitfed", p.GetMetadata)
	mux.Handle("GET /federation/peers", authMw(http.HandlerFunc(p.ListPeers)))
	mux.Handle("POST /federation/peers", authMw(http.HandlerFunc(p.AddPeer)))
	mux.Handle("POST /federation/verify", authMw(http.HandlerFunc(p.VerifyRequest)))
	mux.Handle("GET /federation/resolve/{did}", authMw(http.HandlerFunc(p.ResolveDID)))
}

func (p *FederationPresenter) GetMetadata(w http.ResponseWriter, r *http.Request) {
	instanceURL := strings.TrimSpace(p.instance.InstanceURL)
	if instanceURL == "" {
		instanceURL = "http://localhost:8080"
	}
	instanceName := strings.TrimSpace(p.instance.InstanceName)
	if instanceName == "" {
		instanceName = "toy-gitfed"
	}

	writeJSON(r.Context(), w, http.StatusOK, federation.InstanceMetadata{
		Version:      federation.ProtocolVersion,
		InstanceURL:  instanceURL,
		InstanceName: instanceName,
		AdminEmail:   strings.TrimSpace(p.instance.AdminEmail),
		Capabilities: p.instance.Capabilities,
	})
}

func (p *FederationPresenter) ListPeers(w http.ResponseWriter, r *http.Request) {
	if p.service == nil {
		writeError(r.Context(), w, apperrors.NotImplemented("federation service not configured"), "")
		return
	}
	peers, err := p.service.ListPeers(r.Context())
	if err != nil {
		writeError(r.Context(), w, err, "")
		return
	}
	writeJSON(r.Context(), w, http.StatusOK, peers)
}

type addPeerRequest struct {
	InstanceURL string `json:"instance_url"`
	Name        string `json:"name"`
}

func (p *FederationPresenter) AddPeer(w http.ResponseWriter, r *http.Request) {
	if p.service == nil {
		writeError(r.Context(), w, apperrors.NotImplemented("federation service not configured"), "")
		return
	}

	var req addPeerRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<12)).Decode(&req); err != nil {
		writeError(r.Context(), w, apperrors.New(apperrors.CodeInvalidArgument, "invalid request body"), "")
		return
	}

	peer, err := p.service.AddPeer(r.Context(), federation.Peer{
		InstanceURL: strings.TrimSpace(req.InstanceURL),
		Name:        strings.TrimSpace(req.Name),
	})
	if err != nil {
		writeError(r.Context(), w, err, "")
		return
	}
	writeJSON(r.Context(), w, http.StatusCreated, peer)
}

func (p *FederationPresenter) VerifyRequest(w http.ResponseWriter, r *http.Request) {
	actor, ok := FederationActorDIDFromContext(r.Context())
	if !ok || actor == "" {
		writeError(r.Context(), w, apperrors.New(apperrors.CodeUnauthenticated, "federation principal missing from context"), "")
		return
	}
	writeJSON(r.Context(), w, http.StatusOK, map[string]string{
		"status": "verified",
		"actor":  actor,
	})
}

func (p *FederationPresenter) ResolveDID(w http.ResponseWriter, r *http.Request) {
	if p.service == nil {
		writeError(r.Context(), w, apperrors.NotImplemented("federation service not configured"), "")
		return
	}

	resolution, err := p.service.ResolveDID(r.Context(), strings.TrimSpace(r.PathValue("did")))
	if err != nil {
		writeError(r.Context(), w, err, "")
		return
	}
	writeJSON(r.Context(), w, http.StatusOK, resolution)
}
