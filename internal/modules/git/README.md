# Git Module

Repository management and Smart HTTP protocol layer for the federated Git forge.

## Architecture

```
internal/modules/git/
├── domain.go      # Types, errors, request/response structs
├── repository.go  # RepositoryManager + PackService interfaces
├── service.go     # Service implementation (filesystem-backed)
├── handler.go     # SmartHTTPHandler (http.Handler for Git Smart HTTP)
└── service_test.go
```

### Layers

| Layer | File | Responsibility |
|---|---|---|
| Domain | `domain.go` | `Repository`, `CreateInput`, `RefInfo`, `RepoStats`, sentinel errors |
| Interface | `repository.go` | `RepositoryManager` (CRUD + inspection), `PackService` (upload/receive packs) |
| Service | `service.go` | Filesystem-backed implementation using go-git |
| Handler | `handler.go` | HTTP routing, Smart HTTP protocol, path sanitization |

## Interfaces

### RepositoryManager

Repository lifecycle and inspection:

```go
type RepositoryManager interface {
    Create(ctx context.Context, input CreateInput) (Repository, error)
    GetByID(ctx context.Context, id uuid.UUID) (Repository, error)
    GetByName(ctx context.Context, ownerID uuid.UUID, name string) (Repository, error)
    ListByOwner(ctx context.Context, ownerID uuid.UUID) ([]Repository, error)
    Update(ctx context.Context, id uuid.UUID, input UpdateInput) (Repository, error)
    Delete(ctx context.Context, id uuid.UUID) error
    GetRefs(ctx context.Context, repo Repository) ([]RefInfo, error)
    GetStats(ctx context.Context, repo Repository) (RepoStats, error)
    Exists(repo Repository) bool
    RepoPath(repo Repository) string
}
```

### PackService

Smart HTTP pack operations (fetch/push):

```go
type PackService interface {
    UploadPack(ctx context.Context, req UploadPackRequest, w io.Writer, r io.Reader) error
    ReceivePack(ctx context.Context, req ReceivePackRequest, w io.Writer, r io.Reader) error
}
```

The current `Service` implements both interfaces. The split exists so future implementations can swap the storage layer (e.g., database-backed metadata) while keeping pack operations filesystem-bound.

## Setup

### 1. Create the service

```go
import "github.com/qoppa-tech/toy-gitfed/internal/modules/git"

svc := git.NewService("/var/lib/toy-gitfed/repos")
```

The `reposDir` argument is the absolute path where bare git repositories will be stored.

### 2. Create a Smart HTTP handler

```go
handler := git.NewSmartHTTPHandler(svc)
```

### 3. Mount into your router

```go
mux := http.NewServeMux()
handler.Mount(mux)

// Or use the handler directly:
http.ListenAndServe(":8080", handler)
```

## Usage

### Creating a repository

```go
repo, err := svc.Create(ctx, git.CreateInput{
    Name:        "my-project",
    Description: "A federated project",
    IsPrivate:   false,
    OwnerID:     ownerUUID,
    DefaultRef:  "refs/heads/main",
})
```

Repository names must match `^[a-zA-Z0-9][a-zA-Z0-9._-]*$`.

### Inspecting a repository

```go
// List all refs
refs, err := svc.GetRefs(ctx, repo)
// refs: [{Name: "HEAD", Hash: "..."}, {Name: "refs/heads/main", Hash: "..."}]

// Get statistics
stats, err := svc.GetStats(ctx, repo)
// stats: {BranchCount: 1, TagCount: 0, CommitCount: 42, LastCommitTime: ...}

// Check existence
if svc.Exists(repo) { ... }

// Resolve filesystem path
path := svc.RepoPath(repo)
```

### Smart HTTP endpoints

Once the handler is mounted, standard git clients interact with these endpoints:

| Method | Path | Purpose |
|---|---|---|
| `GET` | `/{repo}/info/refs?service=git-upload-pack` | Ref advertisement (fetch/clone) |
| `GET` | `/{repo}/info/refs?service=git-receive-pack` | Ref advertisement (push) |
| `POST` | `/{repo}/git-upload-pack` | Fetch data transfer |
| `POST` | `/{repo}/git-receive-pack` | Push data transfer |

Clients use them transparently:

```sh
# Clone
git clone http://localhost:8080/my-project

# Push
cd my-project
git remote add origin http://localhost:8080/my-project
git push origin main

# Fetch
git fetch origin
```

### Path utilities

```go
// Validate a repo name before creating
if err := git.ValidateRepoName("my-repo"); err != nil { ... }

// Sanitize user-supplied paths (removes "..", leading/trailing slashes)
safe := git.SanitizeRepoPath("../../../etc/passwd")  // "etc/passwd"

// Build a composed path with sanitization
path := git.BuildRepoPath("/repos", "org", "my-repo")  // "/repos/org/my-repo"
```

## Smart HTTP Protocol Flow

### Clone / Fetch

```
Client                              Server
  |                                   |
  | GET /repo/info/refs?service=      |
  |   git-upload-pack                 |
  |---------------------------------->|
  |                                   | 1. Validate repo exists
  |                                   | 2. UploadPack(Adverts=true)
  |  200 application/x-git-           |
  |    upload-pack-advertisement      |
  |<----------------------------------|
  |                                   |
  | POST /repo/git-upload-pack        |
  |   Content-Type: application/      |
  |     x-git-upload-pack-request     |
  |---------------------------------->|
  |                                   | 3. UploadPack(StatelessRPC=true)
  |  200 application/x-git-           |
  |    upload-pack-result             |
  |  (pkt-line stream + packfile)     |
  |<----------------------------------|
```

### Push

```
Client                              Server
  |                                   |
  | GET /repo/info/refs?service=      |
  |   git-receive-pack                |
  |---------------------------------->|
  |  200 application/x-git-           |
  |    receive-pack-advertisement     |
  |<----------------------------------|
  |                                   |
  | POST /repo/git-receive-pack       |
  |   Content-Type: application/      |
  |     x-git-receive-pack-request    |
  |---------------------------------->|
  |                                   | 4. ReceivePack(StatelessRPC=true)
  |  200 application/x-git-           |
  |    receive-pack-result            |
  |  (pkt-line status report)         |
  |<----------------------------------|
```

All communication uses the [Git pkt-line framing](../pkg/pktline/) protocol. The handler delegates pack negotiation to `github.com/go-git/go-git/v6/plumbing/transport.UploadPack` and `ReceivePack`.

## Current Limitations

| Method | Status | Reason |
|---|---|---|
| `Create` | Implemented | Bare `git init` on filesystem |
| `GetByName` | Implemented | Checks `config` file on disk |
| `GetRefs` | Implemented | Reads via go-git storage |
| `GetStats` | Implemented | Iterates objects and refs |
| `Exists` | Implemented | Stat check for `config` |
| `RepoPath` | Implemented | `filepath.Join(reposDir, name)` |
| `GetByID` | Stub | Requires database |
| `ListByOwner` | Stub | Requires database |
| `Update` | Stub | Requires database |
| `Delete` | Stub | Requires database |

The stub methods return descriptive errors. Once a persistence layer is added, implement these by wiring a `RepositoryManager` that combines the database store with the filesystem `PackService`.

## Dependencies

| Package | Purpose |
|---|---|
| `github.com/go-git/go-git/v6` | Repository init, pack upload/receive, storage |
| `github.com/go-git/go-billy/v6` | Filesystem abstraction (osfs) |
| `github.com/qoppa-tech/toy-gitfed/pkg/pktline` | Git pkt-line framing (encode/decode) |
