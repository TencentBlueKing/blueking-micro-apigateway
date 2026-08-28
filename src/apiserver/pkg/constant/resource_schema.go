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

// Package constant ...
package constant

// ResourceSchemaCapability defines what fields a resource type supports in its APISIX schema
type ResourceSchemaCapability struct {
	// SupportsNameField indicates if the resource type has "name" property in APISIX schema
	// This varies by APISIX version:
	// - 3.11: route, service, upstream, plugin_config
	// - 3.13/3.17: adds consumer_group, stream_route, proto
	// - 3.18: removes name from plugin_config, consumer_group, stream_route, proto
	SupportsNameField bool

	// RequiresIDInSchema indicates if the resource type requires "id" in the schema
	// Note: This is different from whether the resource uses ID - all resources use ID internally,
	// but only some require it in the APISIX schema validation
	RequiresIDInSchema bool

	// UsesIDField indicates if the resource uses "id" as identifier
	// Consumer is special - it uses "username" instead of "id"
	UsesIDField bool
}

// ResourceSupportsNameField checks if a resource type has "name" property in its APISIX 3.11 schema.
// This legacy helper reflects the APISIX 3.11 baseline. New code should use
// ResourceSupportsNameFieldForVersion.
//
// For APISIX 3.11 (current baseline):
//   - route, service, upstream, plugin_config: have "name" field
//   - consumer_group, global_rule, ssl, stream_route, proto: NO "name" field
//
// For version-aware checking, use ResourceSupportsNameFieldForVersion instead.
func ResourceSupportsNameField(resourceType APISIXResource) bool {
	switch resourceType {
	case Route, Service, Upstream, PluginConfig:
		return true
	default:
		return false
	}
}

// ResourceSupportsNameFieldForVersion checks if a resource type has "name" property in the specified APISIX version's
// schema.
// This function is version-aware and returns the correct result for each APISIX version.
//
// APISIX 3.11:
//   - route, service, upstream, plugin_config: have "name"
//   - consumer_group, stream_route, proto: NO "name"
//   - global_rule, ssl: NO "name"
//
// APISIX 3.13/3.17:
//   - route, service, upstream, plugin_config: have "name"
//   - consumer_group, stream_route, proto: have "name" (ADDED in 3.13)
//   - global_rule, ssl: NO "name"
//
// APISIX 3.18:
//   - route, service, upstream: have "name"
//   - plugin_config, consumer_group, stream_route, proto, global_rule, ssl: NO "name"
func ResourceSupportsNameFieldForVersion(resourceType APISIXResource, version APISIXVersion) bool {
	// Resources that support name in all versions
	switch resourceType {
	case Route, Service, Upstream:
		return true
	case PluginConfig:
		return version != APISIXVersion318
	case ConsumerGroup, StreamRoute, Proto:
		// Added in 3.13; older schemas do not expose name.
		return version == APISIXVersion313 || version == APISIXVersion317
	case GlobalRule, SSL:
		// Never supported
		return false
	default:
		return false
	}
}

// ResourceSupportsIDInConfig checks if a resource type should have "id" in its APISIX config.
// Note: All resources use ID internally for database and etcd key, but not all include it in the JSON config.
//
// Consumer is special - it doesn't have "id" in config (uses username as key)
// Most other resources have "id" in config
func ResourceSupportsIDInConfig(resourceType APISIXResource) bool {
	// Consumer doesn't have id in config (uses username)
	return resourceType != Consumer
}

// ShouldRemoveFieldBeforeValidationOrPublish determines if a field should be removed from config before publishing to
// APISIX.
// This is used to clean up internal fields that shouldn't be sent to APISIX.
//
// Rules:
// - "id" should be removed for consumer (uses username as key)
// - "name" should be removed if not supported in the target APISIX version
func ShouldRemoveFieldBeforeValidationOrPublish(
	resourceType APISIXResource,
	fieldName string,
	version APISIXVersion,
) bool {
	switch fieldName {
	case "id":
		// Consumer uses username as key. APISIX 3.18 consumer_group no longer exposes id.
		return resourceType == Consumer ||
			(version == APISIXVersion318 && resourceType == ConsumerGroup)
	case "name":
		// Remove name only if the schema doesn't support it for this version
		return !ResourceSupportsNameFieldForVersion(resourceType, version)
	default:
		return false
	}
}

// ResourceRequiresIDInSchema checks if a resource type requires "id" in the APISIX schema validation.
// These resources have "id" in the schema's required array.
//
// Note: consumer_group, plugin_config, global_rule require "id" in the NEW schema,
// but didn't in the old 2.x schema. This function reflects the CURRENT schema requirements.
func ResourceRequiresIDInSchema(resourceType APISIXResource) bool {
	switch resourceType {
	case ConsumerGroup, PluginConfig, GlobalRule:
		return true
	default:
		return false
	}
}

// ResourceRequiresIDInSchemaForVersion checks if a resource type requires "id" in the specified APISIX version's
// schema.
//
// APISIX 3.2/3.3:
//   - plugin_config, global_rule: expose id property but do not require it
//   - consumer_group: still uses the old group_name schema and does not expose id
//
// APISIX 3.11/3.13/3.17:
//   - consumer_group, plugin_config, global_rule: require id
//
// APISIX 3.18:
//   - consumer_group does not expose id
//   - plugin_config and global_rule expose id but do not require it
func ResourceRequiresIDInSchemaForVersion(resourceType APISIXResource, version APISIXVersion) bool {
	switch resourceType {
	case ConsumerGroup, PluginConfig, GlobalRule:
		switch version {
		case APISIXVersion311, APISIXVersion313, APISIXVersion317:
			return true
		default:
			return false
		}
	default:
		return false
	}
}

// ResourceUsesIDField checks if a resource type uses "id" as its identifier field.
// Consumer is special - it uses "username" instead of "id" as the identifier.
func ResourceUsesIDField(resourceType APISIXResource) bool {
	// Consumer uses "username" as identifier, not "id"
	return resourceType != Consumer
}

// GetResourceSchemaCapability returns the schema capability information for a resource type.
// This provides a consolidated view of what fields the resource supports/requires.
func GetResourceSchemaCapability(resourceType APISIXResource) ResourceSchemaCapability {
	return ResourceSchemaCapability{
		SupportsNameField:  ResourceSupportsNameField(resourceType),
		RequiresIDInSchema: ResourceRequiresIDInSchema(resourceType),
		UsesIDField:        ResourceUsesIDField(resourceType),
	}
}
