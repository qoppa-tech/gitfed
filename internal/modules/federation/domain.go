package federation

import "time"

type ActivityType string

const (
	ActivityDiscover ActivityType = "discover"
	ActivityClone    ActivityType = "clone"
	ActivityFetch    ActivityType = "fetch"
)

type Envelope struct {
	Type      ActivityType `json:"type"`
	Actor     string       `json:"actor"`
	Object    string       `json:"object,omitempty"`
	Published time.Time    `json:"published"`
}

type Peer struct {
	InstanceURL string    `json:"instance_url"`
	Name        string    `json:"name"`
	AddedAt     time.Time `json:"added_at"`
}

type InstanceMetadata struct {
	Version      string   `json:"version"`
	InstanceURL  string   `json:"instance_url"`
	InstanceName string   `json:"instance_name"`
	AdminEmail   string   `json:"admin_email,omitempty"`
	Capabilities []string `json:"capabilities"`
}

type DIDResolution struct {
	DID         string `json:"did"`
	Principal   string `json:"principal"`
	InstanceURL string `json:"instance_url"`
}

type VerifyRequestInput struct {
	Method    string
	Path      string
	ActorDID  string
	Timestamp time.Time
	Nonce     string
	Digest    string
	Version   string
	Signature string
}

type VerifiedRequest struct {
	ActorDID string
}
