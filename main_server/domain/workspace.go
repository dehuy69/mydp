package domain

import (
	"github.com/dehuy69/mydp/main_server/models"
	service "github.com/dehuy69/mydp/main_server/service"
)

type WorkspaceWrapper struct {
	SQLiteCatalogService *service.SQLiteCatalogService
	SQLiteIndexService   *service.SQLiteIndexService
	Workspace            *models.Workspace
	BadgerService        *service.BadgerService
}

func NewWorkspaceWrapper(workspace *models.Workspace, SQLiteCatalogService *service.SQLiteCatalogService, SQLiteIndexService *service.SQLiteIndexService, BadgerService *service.BadgerService) *WorkspaceWrapper {
	return &WorkspaceWrapper{
		SQLiteCatalogService: SQLiteCatalogService,
		SQLiteIndexService:   SQLiteIndexService,
		Workspace:            workspace,
		BadgerService:        BadgerService,
	}
}

func (cw *WorkspaceWrapper) CreateWorkspace() error {
	err := cw.SQLiteCatalogService.CreateWorkspace(cw.Workspace)
	if err != nil {
		return err
	}
	return nil
}
