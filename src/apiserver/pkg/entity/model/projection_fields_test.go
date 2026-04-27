package model_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
)

type sqliteColumnInfo struct {
	Name string
}

func openProjectionTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "projection-fields.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	return db
}

func tableColumnNames(t *testing.T, db *gorm.DB, table string) []string {
	t.Helper()

	var columns []sqliteColumnInfo
	err := db.Raw("PRAGMA table_info(" + table + ")").Scan(&columns).Error
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	names := make([]string, 0, len(columns))
	for _, column := range columns {
		names = append(names, column.Name)
	}
	return names
}

func TestProjectionFieldsSkippedDuringMigration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		model any
		table string
	}{
		{
			name:  "route",
			model: &model.Route{},
			table: "route",
		},
		{
			name:  "consumer",
			model: &model.Consumer{},
			table: "consumer",
		},
		{
			name:  "gateway_sync_data",
			model: &model.GatewaySyncData{},
			table: "gateway_sync_data",
		},
	}

	unwantedColumns := []string{
		"name_value",
		"service_id_value",
		"upstream_id_value",
		"plugin_config_id_value",
		"group_id_value",
		"ssl_id_value",
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := openProjectionTestDB(t)
			if !assert.NoError(t, db.AutoMigrate(tt.model)) {
				return
			}

			columns := tableColumnNames(t, db, tt.table)
			for _, column := range unwantedColumns {
				assert.NotContains(t, columns, column)
			}
		})
	}
}

func TestAliasProjectionQueryWorksWithoutPhysicalProjectionColumns(t *testing.T) {
	t.Parallel()

	db := openProjectionTestDB(t)
	if !assert.NoError(t, db.AutoMigrate(&model.Route{})) {
		return
	}

	if !assert.NoError(t, db.Table("route").Create(map[string]any{
		"id":               "route-id",
		"gateway_id":       1001,
		"name":             "route-a",
		"service_id":       "svc-a",
		"upstream_id":      "ups-a",
		"plugin_config_id": "pc-a",
		"config":           datatypes.JSON(`{"uris":["/test"]}`),
		"status":           constant.ResourceStatusSuccess,
	}).Error) {
		return
	}

	var resource model.ResourceCommonModel
	err := db.Table(model.Route{}.TableName()).
		Select(
			"route.*, name AS name_value, service_id AS service_id_value, upstream_id AS upstream_id_value, plugin_config_id AS plugin_config_id_value",
		).
		Where("id = ?", "route-id").
		Take(&resource).Error
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "route-a", resource.NameValue)
	assert.Equal(t, "svc-a", resource.ServiceIDValue)
	assert.Equal(t, "ups-a", resource.UpstreamIDValue)
	assert.Equal(t, "pc-a", resource.PluginConfigIDValue)

	if !assert.NoError(t, resource.RestoreConfigForRead(constant.Route)) {
		return
	}
	assert.JSONEq(
		t,
		`{"id":"route-id","name":"route-a","service_id":"svc-a","upstream_id":"ups-a","plugin_config_id":"pc-a","uris":["/test"]}`,
		string(resource.Config),
	)
}

func TestGatewaySyncDataQueryFallsBackToConfigWithoutProjectionColumns(t *testing.T) {
	t.Parallel()

	db := openProjectionTestDB(t)
	if !assert.NoError(t, db.AutoMigrate(&model.GatewaySyncData{})) {
		return
	}

	synced := model.GatewaySyncData{
		ID:          "route-id",
		GatewayID:   1001,
		Type:        constant.Route,
		Config:      datatypes.JSON(`{"name":"route-a","service_id":"svc-a","upstream_id":"ups-a"}`),
		ModRevision: 1,
	}
	if !assert.NoError(t, db.Create(&synced).Error) {
		return
	}

	var stored model.GatewaySyncData
	err := db.Where("gateway_id = ? AND id = ?", synced.GatewayID, synced.ID).Take(&stored).Error
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, "route-a", stored.GetName())
	assert.Equal(t, "svc-a", stored.GetServiceID())
	assert.Equal(t, "ups-a", stored.GetUpstreamID())
}
