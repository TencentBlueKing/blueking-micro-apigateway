/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 微网关 (BlueKing - Micro APIGateway) available.
 * Copyright (C) 2025 Tencent. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 *     http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 * to the current version of the project delivered to anyone in the future.
 */

package importflow

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"

	resourcebiz "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz/resource"
	schemabiz "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz/schema"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/dto"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/repo"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/cryptography"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/utils/ginx"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/tests/util"
)

func init() {
	if err := cryptography.Init("jxi18GX5w2qgHwfZCFpn07q8FScXJOd3", "k2dbCGetyusW"); err != nil {
		panic(err)
	}
	util.InitEmbedDb()
}

func TestBuildImportIndex(t *testing.T) {
	gatewayCtx, gateway := setupImportGatewayContext(t, "build-index")
	createPluginConfigForImportTest(
		t,
		gatewayCtx,
		gateway.ID,
		"pc-existing",
		`{"id":"pc-existing","name":"pc-existing","plugins":{"limit-count":{"count":1,"time_window":60,"key":"remote_addr","policy":"local"}}}`,
	)
	createCustomSchemaForImportTest(t, gatewayCtx, gateway.ID, "existing-plugin")

	got, err := BuildImportIndex(
		gatewayCtx,
		map[constant.APISIXResource][]*dto.ImportResourceInfo{
			constant.PluginConfig: {
				{
					ResourceType: constant.PluginConfig,
					ResourceID:   "pc-new",
					Name:         "pc-new",
					Config: json.RawMessage(
						`{"id":"pc-new","name":"pc-new","plugins":{"limit-count":{"count":10,"time_window":60,"key":"remote_addr","policy":"local"}}}`,
					),
				},
			},
			constant.Schema: {
				{
					ResourceType: constant.Schema,
					Name:         "existing-plugin",
					Config: json.RawMessage(
						`{"schema":{"type":"object"},"example":{"count":1}}`,
					),
				},
				{
					ResourceType: constant.Schema,
					Name:         "new-plugin",
					Config: json.RawMessage(
						`{"schema":{"type":"object"},"example":{"count":2}}`,
					),
				},
			},
		},
	)
	if !assert.NoError(t, err) {
		return
	}
	if !assert.Len(t, got.ResourceTypeMap[constant.PluginConfig], 1) {
		return
	}
	assert.Contains(
		t,
		got.ExistingResourceIDs,
		fmt.Sprintf(constant.ResourceKeyFormat, constant.PluginConfig, "pc-existing"),
	)
	assert.Contains(
		t,
		got.AllResourceIDs,
		fmt.Sprintf(constant.ResourceKeyFormat, constant.PluginConfig, "pc-existing"),
	)
	assert.Contains(
		t,
		got.AllResourceIDs,
		fmt.Sprintf(constant.ResourceKeyFormat, constant.PluginConfig, "pc-new"),
	)
	assert.Equal(t, "pc-new", got.ResourceTypeMap[constant.PluginConfig][0].ID)
	assert.Contains(t, got.AddedSchemaMap, "new-plugin")
	assert.Contains(t, got.UpdatedSchemaMap, "existing-plugin")
	assert.Contains(t, got.AllSchemaMap, "existing-plugin")
	assert.Contains(t, got.AllSchemaMap, "new-plugin")
}

func TestClassifyImportResources(t *testing.T) {
	got, err := ClassifyImportResources(
		map[constant.APISIXResource][]*dto.ImportResourceInfo{
			constant.PluginConfig: {
				{
					ResourceType: constant.PluginConfig,
					ResourceID:   "pc-existing",
					Config: json.RawMessage(
						`{"id":"pc-existing","name":"pc-existing","plugins":{"limit-count":{"count":1,"time_window":60,"key":"remote_addr","policy":"local"}}}`,
					),
				},
				{
					ResourceType: constant.PluginConfig,
					ResourceID:   "pc-new",
					Config: json.RawMessage(
						`{"id":"pc-new","name":"pc-new","plugins":{"limit-count":{"count":10,"time_window":60,"key":"remote_addr","policy":"local"}}}`,
					),
				},
			},
			constant.Schema: {
				{
					ResourceType: constant.Schema,
					Name:         "existing-plugin",
					Config: json.RawMessage(
						`{"schema":{"type":"object"},"example":{"count":1}}`,
					),
				},
				{
					ResourceType: constant.Schema,
					Name:         "new-plugin",
					Config: json.RawMessage(
						`{"schema":{"type":"object"},"example":{"count":2}}`,
					),
				},
			},
		},
		map[string]struct{}{
			fmt.Sprintf(constant.ResourceKeyFormat, constant.PluginConfig, "pc-existing"): {},
		},
		map[string]*model.GatewayCustomPluginSchema{
			"new-plugin": {Name: "new-plugin"},
		},
	)
	if !assert.NoError(t, err) {
		return
	}
	if !assert.Len(t, got.Add[constant.PluginConfig], 1) {
		return
	}
	if !assert.Len(t, got.Update[constant.PluginConfig], 1) {
		return
	}
	assert.Equal(t, constant.UploadStatusAdd, got.Add[constant.PluginConfig][0].Status)
	assert.Equal(t, "pc-new", got.Add[constant.PluginConfig][0].Name)
	assert.Equal(t, constant.UploadStatusUpdate, got.Update[constant.PluginConfig][0].Status)
	assert.Len(t, got.Add[constant.Schema], 1)
	assert.Len(t, got.Update[constant.Schema], 1)
	assert.Equal(t, "new-plugin", got.Add[constant.Schema][0].Name)
	assert.Equal(t, "existing-plugin", got.Update[constant.Schema][0].Name)
}

func TestPrepareImportUpload(t *testing.T) {
	t.Run("ignore_fields overlays existing config field", func(t *testing.T) {
		gatewayCtx, gateway := setupImportGatewayContext(t, "import-overlay")
		createPluginConfigForImportTest(
			t,
			gatewayCtx,
			gateway.ID,
			"pc-1",
			`{"id":"pc-1","name":"pc-demo","desc":"old-desc","plugins":{"limit-count":{"count":1,"time_window":60,"key":"remote_addr","policy":"local"}}}`,
		)

		got, err := PrepareImportUpload(
			gatewayCtx,
			&dto.ImportUploadInfo{
				Add: map[constant.APISIXResource][]*dto.ImportResourceInfo{},
				Update: map[constant.APISIXResource][]*dto.ImportResourceInfo{
					constant.PluginConfig: {
						{
							ResourceType: constant.PluginConfig,
							ResourceID:   "pc-1",
							Name:         "pc-demo",
							Config: json.RawMessage(
								`{"id":"pc-1","name":"pc-demo","desc":"new-desc","plugins":{"limit-count":{"count":10,"time_window":60,"key":"remote_addr","policy":"local"}}}`,
							),
						},
					},
				},
			},
			map[string]any{},
			map[constant.APISIXResource][]string{
				constant.PluginConfig: {"desc"},
			},
		)
		if !assert.NoError(t, err) {
			return
		}
		if !assert.Len(t, got.UpdateResourceTypeMap[constant.PluginConfig], 1) {
			return
		}
		assert.JSONEq(
			t,
			`{"id":"pc-1","name":"pc-demo","desc":"old-desc","plugins":{"limit-count":{"count":10,"time_window":60,"key":"remote_addr","policy":"local"}}}`,
			string(got.UpdateResourceTypeMap[constant.PluginConfig][0].Config),
		)
	})

	t.Run("missing ignore_fields source keeps imported config", func(t *testing.T) {
		gatewayCtx, gateway := setupImportGatewayContext(t, "import-ignore-missing")
		createPluginConfigForImportTest(
			t,
			gatewayCtx,
			gateway.ID,
			"pc-2",
			`{"id":"pc-2","name":"pc-demo","plugins":{"limit-count":{"count":1,"time_window":60,"key":"remote_addr","policy":"local"}}}`,
		)

		got, err := PrepareImportUpload(
			gatewayCtx,
			&dto.ImportUploadInfo{
				Add: map[constant.APISIXResource][]*dto.ImportResourceInfo{},
				Update: map[constant.APISIXResource][]*dto.ImportResourceInfo{
					constant.PluginConfig: {
						{
							ResourceType: constant.PluginConfig,
							ResourceID:   "pc-2",
							Name:         "pc-demo",
							Config: json.RawMessage(
								`{"id":"pc-2","name":"pc-demo","plugins":{"limit-count":{"count":10,"time_window":60,"key":"remote_addr","rejected_code":503,"policy":"local"}}}`,
							),
						},
					},
				},
			},
			map[string]any{},
			map[constant.APISIXResource][]string{
				constant.PluginConfig: {"plugins.limit-count.rejected_code"},
			},
		)
		if !assert.NoError(t, err) {
			return
		}
		if !assert.Len(t, got.UpdateResourceTypeMap[constant.PluginConfig], 1) {
			return
		}
		assert.JSONEq(
			t,
			`{"id":"pc-2","name":"pc-demo","plugins":{"limit-count":{"count":10,"time_window":60,"key":"remote_addr","rejected_code":503,"policy":"local"}}}`,
			string(got.UpdateResourceTypeMap[constant.PluginConfig][0].Config),
		)
	})

	t.Run("keeps provided resource ids and add update counts", func(t *testing.T) {
		gatewayCtx, gateway := setupImportGatewayContext(t, "import-id-preserve")
		createPluginConfigForImportTest(
			t,
			gatewayCtx,
			gateway.ID,
			"pc-existing",
			`{"id":"pc-existing","name":"pc-existing","plugins":{"limit-count":{"count":1,"time_window":60,"key":"remote_addr","policy":"local"}}}`,
		)

		got, err := PrepareImportUpload(
			gatewayCtx,
			&dto.ImportUploadInfo{
				Add: map[constant.APISIXResource][]*dto.ImportResourceInfo{
					constant.PluginConfig: {
						{
							ResourceType: constant.PluginConfig,
							ResourceID:   "pc-new",
							Name:         "pc-new",
							Config: json.RawMessage(
								`{"id":"pc-new","name":"pc-new","plugins":{"limit-count":{"count":10,"time_window":60,"key":"remote_addr","policy":"local"}}}`,
							),
						},
					},
				},
				Update: map[constant.APISIXResource][]*dto.ImportResourceInfo{
					constant.PluginConfig: {
						{
							ResourceType: constant.PluginConfig,
							ResourceID:   "pc-existing",
							Name:         "pc-existing",
							Config: json.RawMessage(
								`{"id":"pc-existing","name":"pc-existing-updated","plugins":{"limit-count":{"count":20,"time_window":60,"key":"remote_addr","policy":"local"}}}`,
							),
						},
					},
				},
			},
			map[string]any{},
			nil,
		)
		if !assert.NoError(t, err) {
			return
		}
		if !assert.Len(t, got.AddResourceTypeMap[constant.PluginConfig], 1) {
			return
		}
		if !assert.Len(t, got.UpdateResourceTypeMap[constant.PluginConfig], 1) {
			return
		}
		assert.Equal(t, "pc-new", got.AddResourceTypeMap[constant.PluginConfig][0].ID)
		assert.Equal(t, "pc-existing", got.UpdateResourceTypeMap[constant.PluginConfig][0].ID)
	})

	t.Run("empty resource id fails before upload", func(t *testing.T) {
		gatewayCtx, _ := setupImportGatewayContext(t, "import-empty-id")

		got, err := PrepareImportUpload(
			gatewayCtx,
			&dto.ImportUploadInfo{
				Add: map[constant.APISIXResource][]*dto.ImportResourceInfo{
					constant.PluginConfig: {
						{
							ResourceType: constant.PluginConfig,
							Name:         "pc-empty-id",
							Config:       json.RawMessage(`{"name":"pc-empty-id"}`),
						},
					},
				},
				Update: map[constant.APISIXResource][]*dto.ImportResourceInfo{},
			},
			map[string]any{},
			nil,
		)
		assert.Nil(t, got)
		assert.ErrorContains(t, err, "resource id is empty")
	})

	t.Run("missing associated resource fails during upload handling", func(t *testing.T) {
		gatewayCtx, _ := setupImportGatewayContext(t, "import-missing-association")

		got, err := PrepareImportUpload(
			gatewayCtx,
			&dto.ImportUploadInfo{
				Add: map[constant.APISIXResource][]*dto.ImportResourceInfo{
					constant.Route: {
						{
							ResourceType: constant.Route,
							ResourceID:   "route-1",
							Name:         "route-demo",
							Config: json.RawMessage(
								`{"id":"route-1","name":"route-demo","uris":["/demo"],"upstream_id":"up-missing"}`,
							),
						},
					},
				},
				Update: map[constant.APISIXResource][]*dto.ImportResourceInfo{},
			},
			map[string]any{},
			nil,
		)
		assert.Nil(t, got)
		assert.ErrorContains(t, err, "associated upstream [id:up-missing] not found")
	})
}

func TestValidateImportedResources(t *testing.T) {
	ctx := ginx.SetGatewayInfoToContext(
		context.Background(),
		&model.Gateway{APISIXVersion: string(constant.APISIXVersion313)},
	)

	err := ValidateImportedResources(
		ctx,
		map[constant.APISIXResource][]*model.GatewaySyncData{
			constant.Route: {
				{
					Type: constant.Route,
					ID:   "route-1",
					Config: datatypes.JSON(
						[]byte(
							`{"id":"route-1","name":"route-demo","uris":["/demo"],"upstream_id":"up-missing"}`,
						),
					),
				},
			},
		},
		map[string]struct{}{
			fmt.Sprintf(constant.ResourceKeyFormat, constant.Route, "route-1"): {},
		},
		nil,
	)
	assert.ErrorContains(t, err, "associated upstream [id:up-missing] not found")
}

func TestApplyImportIgnoreFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		imported     string
		existing     string
		ignoreFields []string
		want         string
	}{
		{
			name:         "overlay top level field from existing config",
			imported:     `{"name":"route-a","desc":"new-desc","plugins":{}}`,
			existing:     `{"name":"route-a","desc":"old-desc","plugins":{"limit-count":{"count":1}}}`,
			ignoreFields: []string{"desc"},
			want:         `{"name":"route-a","desc":"old-desc","plugins":{}}`,
		},
		{
			name:         "overlay nested field from existing config",
			imported:     `{"plugins":{"limit-count":{"count":10,"time_window":60}}}`,
			existing:     `{"plugins":{"limit-count":{"count":1,"time_window":120}}}`,
			ignoreFields: []string{"plugins.limit-count.count"},
			want:         `{"plugins":{"limit-count":{"count":1,"time_window":60}}}`,
		},
		{
			name:         "ignore missing field keeps imported config",
			imported:     `{"plugins":{}}`,
			existing:     `{"name":"route-a"}`,
			ignoreFields: []string{"plugins.limit-count"},
			want:         `{"plugins":{}}`,
		},
		{
			name:         "partial missing fields only overlays existing fields",
			imported:     `{"desc":"new-desc","plugins":{"limit-count":{"count":10}}}`,
			existing:     `{"desc":"old-desc"}`,
			ignoreFields: []string{"desc", "plugins.limit-count.count"},
			want:         `{"desc":"old-desc","plugins":{"limit-count":{"count":10}}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := applyImportIgnoreFields(
				json.RawMessage(tt.imported),
				datatypes.JSON([]byte(tt.existing)),
				tt.ignoreFields,
			)
			assert.NoError(t, err)
			assert.JSONEq(t, tt.want, string(got))
		})
	}
}

func TestLoadExistingImportResources(t *testing.T) {
	gatewayCtx, gateway := setupImportGatewayContext(t, "load-existing")
	createPluginConfigForImportTest(
		t,
		gatewayCtx,
		gateway.ID,
		"pc-1",
		`{"id":"pc-1","name":"pc-demo","plugins":{"limit-count":{"count":1,"time_window":60,"key":"remote_addr","policy":"local"}}}`,
	)

	got, allResourceIDs, err := loadExistingImportResources(gatewayCtx, constant.PluginConfig)
	assert.NoError(t, err)
	assert.Contains(t, got, fmt.Sprintf(constant.ResourceKeyFormat, constant.PluginConfig, "pc-1"))
	assert.Contains(t, allResourceIDs, fmt.Sprintf(constant.ResourceKeyFormat, constant.PluginConfig, "pc-1"))

	t.Run("empty DB returns empty map", func(t *testing.T) {
		got, empty, err := loadExistingImportResources(gatewayCtx, constant.Upstream)
		assert.NoError(t, err)
		assert.Empty(t, got)
		assert.Empty(t, empty)
	})
}

func TestBuildImportSyncData(t *testing.T) {
	t.Parallel()

	ctx := ginx.SetGatewayInfoToContext(context.Background(), &model.Gateway{ID: 23})
	info := &dto.ImportResourceInfo{
		ResourceType: constant.Route,
		ResourceID:   "route-1",
		Name:         "route-demo",
		Config:       json.RawMessage(`{"id":"route-1","name":"route-demo","uri":"/demo"}`),
	}

	got := buildImportSyncData(ctx, constant.Route, info)
	assert.Equal(t, constant.Route, got.Type)
	assert.Equal(t, "route-1", got.ID)
	assert.Equal(t, 23, got.GatewayID)
	assert.JSONEq(t, `{"id":"route-1","name":"route-demo","uri":"/demo"}`, string(got.Config))
}

func TestPrepareImportResources(t *testing.T) {
	gatewayCtx, gateway := setupImportGatewayContext(t, "prepare-resources")
	createPluginConfigForImportTest(
		t,
		gatewayCtx,
		gateway.ID,
		"pc-1",
		`{"id":"pc-1","name":"pc-demo","desc":"old-desc","plugins":{"limit-count":{"count":1,"time_window":60,"key":"remote_addr","policy":"local"}}}`,
	)

	resources, allResourceIDs, err := prepareImportResources(
		gatewayCtx,
		map[constant.APISIXResource][]*dto.ImportResourceInfo{
			constant.PluginConfig: {
				{
					ResourceType: constant.PluginConfig,
					ResourceID:   "pc-1",
					Name:         "pc-demo",
					Config: json.RawMessage(
						`{"id":"pc-1","name":"pc-demo","desc":"new-desc","plugins":{"limit-count":{"count":10,"time_window":60,"key":"remote_addr","policy":"local"}}}`,
					),
				},
			},
		},
		map[constant.APISIXResource][]string{
			constant.PluginConfig: {"desc"},
		},
	)
	assert.NoError(t, err)
	assert.Contains(t, allResourceIDs, fmt.Sprintf(constant.ResourceKeyFormat, constant.PluginConfig, "pc-1"))
	if !assert.Len(t, resources[constant.PluginConfig], 1) {
		return
	}
	assert.JSONEq(
		t,
		`{"id":"pc-1","name":"pc-demo","desc":"old-desc","plugins":{"limit-count":{"count":10,"time_window":60,"key":"remote_addr","policy":"local"}}}`,
		string(resources[constant.PluginConfig][0].Config),
	)

	t.Run("schema resources are skipped", func(t *testing.T) {
		got, allResourceIDs, err := prepareImportResources(
			gatewayCtx,
			map[constant.APISIXResource][]*dto.ImportResourceInfo{
				constant.Schema: {
					{
						ResourceType: constant.Schema,
						ResourceID:   "schema-1",
						Name:         "demo-plugin",
						Config:       json.RawMessage(`{"name":"demo-plugin"}`),
					},
				},
				constant.Route: {
					{
						ResourceType: constant.Route,
						ResourceID:   "route-1",
						Name:         "route-demo",
						Config: json.RawMessage(
							`{"id":"route-1","name":"route-demo","uris":["/demo"]}`,
						),
					},
				},
			},
			nil,
		)
		assert.NoError(t, err)
		assert.Contains(t, allResourceIDs, fmt.Sprintf(constant.ResourceKeyFormat, constant.Route, "route-1"))
		assert.NotContains(t, got, constant.Schema)
		assert.Len(t, got[constant.Route], 1)
	})
}

func TestPrepareImportValidationInput(t *testing.T) {
	ctx := ginx.SetGatewayInfoToContext(context.Background(), &model.Gateway{ID: 31})

	t.Run("add only", func(t *testing.T) {
		input, err := prepareImportValidationInput(
			ctx,
			&dto.ImportUploadInfo{
				Add: map[constant.APISIXResource][]*dto.ImportResourceInfo{
					constant.Route: {
						{
							ResourceType: constant.Route,
							ResourceID:   "route-1",
							Name:         "route-demo",
							Config: json.RawMessage(
								`{"id":"route-1","name":"route-demo","uri":"/demo"}`,
							),
						},
					},
				},
				Update: map[constant.APISIXResource][]*dto.ImportResourceInfo{},
			},
			nil,
		)
		assert.NoError(t, err)
		assert.Contains(
			t,
			input.AllResourceIDs,
			fmt.Sprintf(constant.ResourceKeyFormat, constant.Route, "route-1"),
		)
		assert.Len(t, input.Add, 1)
		assert.Len(t, input.Add[constant.Route], 1)
		assert.Empty(t, input.Update)
	})

	t.Run("add and update accumulate all resource ids", func(t *testing.T) {
		input, err := prepareImportValidationInput(
			ctx,
			&dto.ImportUploadInfo{
				Add: map[constant.APISIXResource][]*dto.ImportResourceInfo{
					constant.Route: {
						{
							ResourceType: constant.Route,
							ResourceID:   "route-new",
							Name:         "route-new",
							Config:       json.RawMessage(`{"id":"route-new","uri":"/a"}`),
						},
					},
				},
				Update: map[constant.APISIXResource][]*dto.ImportResourceInfo{
					constant.Route: {
						{
							ResourceType: constant.Route,
							ResourceID:   "route-upd",
							Name:         "route-upd",
							Config:       json.RawMessage(`{"id":"route-upd","uri":"/b"}`),
						},
					},
				},
			},
			nil,
		)
		assert.NoError(t, err)
		assert.Contains(
			t,
			input.AllResourceIDs,
			fmt.Sprintf(constant.ResourceKeyFormat, constant.Route, "route-new"),
		)
		assert.Contains(
			t,
			input.AllResourceIDs,
			fmt.Sprintf(constant.ResourceKeyFormat, constant.Route, "route-upd"),
		)
		assert.Len(t, input.Add[constant.Route], 1)
		assert.Len(t, input.Update[constant.Route], 1)
	})
}

func setupImportGatewayContext(t *testing.T, suffix string) (context.Context, *model.Gateway) {
	t.Helper()

	ctx := context.Background()
	gateway := &model.Gateway{
		Name:          publishTestName(t, suffix),
		APISIXVersion: string(constant.APISIXVersion313),
	}
	if err := repo.Gateway.WithContext(ctx).Create(gateway); err != nil {
		t.Fatal(err)
	}

	ctx = ginx.SetGatewayInfoToContext(ctx, gateway)
	ctx = context.WithValue(ctx, constant.UserIDKey, "import-tester")
	return ctx, gateway
}

func createPluginConfigForImportTest(
	t *testing.T,
	ctx context.Context,
	gatewayID int,
	id string,
	config string,
) {
	t.Helper()

	err := resourcebiz.CreatePluginConfig(ctx, model.PluginConfig{
		Name: gjsonGetName(config),
		ResourceCommonModel: model.ResourceCommonModel{
			ID:        id,
			GatewayID: gatewayID,
			Config:    datatypes.JSON([]byte(config)),
			Status:    constant.ResourceStatusSuccess,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func createCustomSchemaForImportTest(t *testing.T, ctx context.Context, gatewayID int, name string) {
	t.Helper()

	err := schemabiz.BatchCreateSchema(ctx, []*model.GatewayCustomPluginSchema{
		{
			GatewayID: gatewayID,
			Name:      name,
			Schema:    datatypes.JSON([]byte(`{"type":"object"}`)),
			Example:   datatypes.JSON([]byte(`{"count":1}`)),
			BaseModel: model.BaseModel{
				Creator: ginx.GetUserIDFromContext(ctx),
				Updater: ginx.GetUserIDFromContext(ctx),
			},
			OperationType: constant.OperationImport,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func gjsonGetName(config string) string {
	var info struct {
		Name string `json:"name"`
	}
	_ = json.Unmarshal([]byte(config), &info)
	return info.Name
}

func publishTestName(t *testing.T, suffix string) string {
	t.Helper()
	name := strings.ToLower(t.Name() + "-" + suffix)
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, "_", "-")
	if len(name) > 48 {
		name = name[:24] + "-" + name[len(name)-23:]
	}
	return name
}
