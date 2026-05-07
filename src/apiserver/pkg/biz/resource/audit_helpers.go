// Package resource contains APISIX resource CRUD helpers shared by higher-level biz domains.
package resource

import (
	"context"

	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz/auditlog"
	schemabiz "github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/biz/schema"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/constant"
	"github.com/TencentBlueKing/blueking-micro-apigateway/apiserver/pkg/entity/model"
)

type FuncUpdateResourceStatusByID func(
	ctx context.Context,
	resourceType constant.APISIXResource,
	id string,
	status constant.ResourceStatus,
) error

type FuncBatchUpdateResourceStatus func(
	ctx context.Context,
	resourceType constant.APISIXResource,
	ids []string,
	status constant.ResourceStatus,
) error

func wrapUpdateResourceStatusByIDAddAuditLog(
	ctx context.Context,
	resourceType constant.APISIXResource,
	id string,
	status constant.ResourceStatus,
	fn FuncUpdateResourceStatusByID,
) error {
	resourceInfo, err := GetResourceByID(ctx, resourceType, id)
	if err != nil {
		return err
	}
	if err := fn(ctx, resourceType, id, status); err != nil {
		return err
	}
	return auditlog.AddBatchAuditLog(
		ctx,
		constant.OperationTypeUpdate,
		resourceType,
		[]*model.ResourceCommonModel{&resourceInfo},
		map[string]constant.ResourceStatus{id: status},
	)
}

func wrapBatchUpdateResourceStatusAddAuditLog(
	ctx context.Context,
	resourceType constant.APISIXResource,
	resourceIDs []string,
	status constant.ResourceStatus,
	fn FuncBatchUpdateResourceStatus,
) error {
	resourceList, err := BatchGetResources(ctx, resourceType, resourceIDs)
	if err != nil {
		return err
	}
	if err := fn(ctx, resourceType, resourceIDs, status); err != nil {
		return err
	}
	resourceStatusMap := make(map[string]constant.ResourceStatus, len(resourceList))
	for _, resource := range resourceList {
		resourceStatusMap[resource.ID] = status
	}
	return auditlog.AddBatchAuditLog(
		ctx,
		constant.OperationTypeUpdate,
		resourceType,
		resourceList,
		resourceStatusMap,
	)
}

func addDeleteResourceByIDAuditLog(
	ctx context.Context,
	resourceType constant.APISIXResource,
	resourceIDs []string,
) error {
	resourceList, err := BatchGetResources(ctx, resourceType, resourceIDs)
	if err != nil {
		return err
	}
	resourceStatusMap := make(map[string]constant.ResourceStatus, len(resourceList))
	for _, resource := range resourceList {
		if resource.Status == constant.ResourceStatusCreateDraft {
			resourceStatusMap[resource.ID] = constant.ResourceStatusDeleted
			continue
		}
		resourceStatusMap[resource.ID] = constant.ResourceStatusDeleteDraft
	}
	return auditlog.AddBatchAuditLog(
		ctx,
		constant.OperationTypeDelete,
		resourceType,
		resourceList,
		resourceStatusMap,
	)
}

func wrapBatchRevertResourceAddAuditLog(
	ctx context.Context,
	resourceType constant.APISIXResource,
	resourceIDs []string,
	afterResources []*model.ResourceCommonModel,
) error {
	resourceList, err := BatchGetResources(ctx, resourceType, resourceIDs)
	if err != nil {
		return err
	}
	return auditlog.AddRevertAuditLog(ctx, resourceType, resourceIDs, resourceList, afterResources)
}

func batchDeleteResourceSchemaAssociation(
	ctx context.Context,
	resourceIDs []string,
	resourceType constant.APISIXResource,
) error {
	return schemabiz.BatchDeleteResourceSchemaAssociation(ctx, resourceIDs, resourceType)
}
