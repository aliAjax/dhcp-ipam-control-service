package infrastructure

import "github.com/example/dhcp-ipam-control/internal/platform/storage"

type Repository struct{ Store *storage.Memory }

func NewRepository() *Repository { return &Repository{Store: storage.NewMemory()} }
