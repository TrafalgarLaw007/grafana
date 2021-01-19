package librarypanels

import (
	"errors"
	"fmt"

	"github.com/grafana/grafana/pkg/api/routing"
	"github.com/grafana/grafana/pkg/components/simplejson"
	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/registry"
	"github.com/grafana/grafana/pkg/services/sqlstore"
	"github.com/grafana/grafana/pkg/services/sqlstore/migrator"
	"github.com/grafana/grafana/pkg/setting"
)

// LibraryPanelService is the service for the Panel Library feature.
type LibraryPanelService struct {
	Cfg           *setting.Cfg          `inject:""`
	SQLStore      *sqlstore.SQLStore    `inject:""`
	RouteRegister routing.RouteRegister `inject:""`
	log           log.Logger
}

func init() {
	registry.RegisterService(&LibraryPanelService{})
}

// Init initializes the LibraryPanel service
func (lps *LibraryPanelService) Init() error {
	lps.log = log.New("librarypanels")

	lps.registerAPIEndpoints()

	return nil
}

// IsEnabled returns true if the Panel Library feature is enabled for this instance.
func (lps *LibraryPanelService) IsEnabled() bool {
	if lps.Cfg == nil {
		return false
	}

	return lps.Cfg.IsPanelLibraryEnabled()
}

// LoadLibraryPanelsForDashboard loads library panels JSON from the database for a dashboard.
func (lps *LibraryPanelService) LoadLibraryPanelsForDashboard(dash *models.Dashboard) error {
	if !lps.IsEnabled() {
		return nil
	}

	libraryPanels, err := lps.getLibraryPanelsForDashboardID(dash.Id)
	if err != nil {
		return err
	}

	panels := dash.Data.Get("panels").MustArray()
	for i, panel := range panels {
		panelAsJSON := simplejson.NewFromAny(panel)
		libraryPanel := panelAsJSON.Get("libraryPanel")
		if libraryPanel.Interface() == nil {
			continue
		}

		// we have a library panel
		uid := libraryPanel.Get("uid").MustString()
		if len(uid) == 0 {
			return errors.New("found a library panel without uid")
		}

		libraryPanelInDB, ok := libraryPanels[uid]
		if !ok {
			return errors.New("found a library panel that does not exists as a connection")
		}

		// we have a match between what is stored in db and in dashboard json
		libraryPanelModel, err := libraryPanelInDB.Model.MarshalJSON()
		if err != nil {
			return fmt.Errorf("could not marshal library panel JSON: %w", err)
		}

		libraryPanelModelAsJSON, err := simplejson.NewJson(libraryPanelModel)
		if err != nil {
			return fmt.Errorf("could not convert library panel to JSON: %w", err)
		}

		// set the library panel json as new panel json in dashboard json
		dash.Data.Get("panels").SetIndex(i, libraryPanelModelAsJSON.Interface())

		// set dashboard specific props
		elem := dash.Data.Get("panels").GetIndex(i)
		elem.Set("gridPos", panelAsJSON.Get("gridPos").MustMap())
		elem.Set("id", panelAsJSON.Get("id").MustInt64())
		elem.Set("libraryPanel", map[string]interface{}{
			"uid":  libraryPanelInDB.UID,
			"name": libraryPanelInDB.Name,
		})
	}

	return nil
}

// CleanLibraryPanelsForDashboard cleans library panels JSON before storing a dashboard to the database.
func (lps *LibraryPanelService) CleanLibraryPanelsForDashboard(dash *models.Dashboard) error {
	if !lps.IsEnabled() {
		return nil
	}

	panels := dash.Data.Get("panels").MustArray()
	for i, panel := range panels {
		panelAsJSON := simplejson.NewFromAny(panel)
		libraryPanel := panelAsJSON.Get("libraryPanel")
		if libraryPanel.Interface() == nil {
			continue
		}

		// we have a library panel
		uid := libraryPanel.Get("uid").MustString()
		if len(uid) == 0 {
			return errors.New("found a library panel without uid")
		}
		name := libraryPanel.Get("name").MustString()
		if len(name) == 0 {
			return errors.New("found a library panel without name")
		}

		gridPos := panelAsJSON.Get("gridPos").MustMap()
		id := panelAsJSON.Get("id").MustInt64(int64(i))
		dash.Data.Get("panels").SetIndex(i, map[string]interface{}{
			"id":      id,
			"gridPos": gridPos,
			"libraryPanel": map[string]interface{}{
				"uid":  uid,
				"name": name,
			},
		})
	}

	return nil
}

// ConnectLibraryPanelsForDashboard connects library panels to a new dashboard.
func (lps *LibraryPanelService) ConnectLibraryPanelsForDashboard(c *models.ReqContext, dash *models.Dashboard) error {
	if !lps.IsEnabled() {
		return nil
	}

	if dash.Id == 0 || dash.Uid == "" {
		return errors.New("dashboard is missing an ID or uid")
	}

	panels := dash.Data.Get("panels").MustArray()
	for _, panel := range panels {
		panelAsJSON := simplejson.NewFromAny(panel)
		libraryPanel := panelAsJSON.Get("libraryPanel")
		if libraryPanel.Interface() == nil {
			continue
		}

		// we have a library panel
		uid := libraryPanel.Get("uid").MustString()
		if len(uid) == 0 {
			return errors.New("found a library panel without uid")
		}
		err := lps.connectDashboard(c, uid, dash.Id)
		if err != nil {
			return err
		}
	}

	return nil
}

// AddMigration defines database migrations.
// If Panel Library is not enabled does nothing.
func (lps *LibraryPanelService) AddMigration(mg *migrator.Migrator) {
	if !lps.IsEnabled() {
		return
	}

	libraryPanelV1 := migrator.Table{
		Name: "library_panel",
		Columns: []*migrator.Column{
			{Name: "id", Type: migrator.DB_BigInt, IsPrimaryKey: true, IsAutoIncrement: true},
			{Name: "org_id", Type: migrator.DB_BigInt, Nullable: false},
			{Name: "folder_id", Type: migrator.DB_BigInt, Nullable: false},
			{Name: "uid", Type: migrator.DB_NVarchar, Length: 40, Nullable: false},
			{Name: "name", Type: migrator.DB_NVarchar, Length: 255, Nullable: false},
			{Name: "model", Type: migrator.DB_Text, Nullable: false},
			{Name: "created", Type: migrator.DB_DateTime, Nullable: false},
			{Name: "created_by", Type: migrator.DB_BigInt, Nullable: false},
			{Name: "updated", Type: migrator.DB_DateTime, Nullable: false},
			{Name: "updated_by", Type: migrator.DB_BigInt, Nullable: false},
		},
		Indices: []*migrator.Index{
			{Cols: []string{"org_id", "folder_id", "name"}, Type: migrator.UniqueIndex},
		},
	}

	mg.AddMigration("create library_panel table v1", migrator.NewAddTableMigration(libraryPanelV1))
	mg.AddMigration("add index library_panel org_id & folder_id & name", migrator.NewAddIndexMigration(libraryPanelV1, libraryPanelV1.Indices[0]))

	libraryPanelDashboardV1 := migrator.Table{
		Name: "library_panel_dashboard",
		Columns: []*migrator.Column{
			{Name: "id", Type: migrator.DB_BigInt, IsPrimaryKey: true, IsAutoIncrement: true},
			{Name: "librarypanel_id", Type: migrator.DB_BigInt, Nullable: false},
			{Name: "dashboard_id", Type: migrator.DB_BigInt, Nullable: false},
			{Name: "created", Type: migrator.DB_DateTime, Nullable: false},
			{Name: "created_by", Type: migrator.DB_BigInt, Nullable: false},
		},
		Indices: []*migrator.Index{
			{Cols: []string{"librarypanel_id", "dashboard_id"}, Type: migrator.UniqueIndex},
		},
	}

	mg.AddMigration("create library_panel_dashboard table v1", migrator.NewAddTableMigration(libraryPanelDashboardV1))
	mg.AddMigration("add index library_panel_dashboard librarypanel_id & dashboard_id", migrator.NewAddIndexMigration(libraryPanelDashboardV1, libraryPanelDashboardV1.Indices[0]))
}
