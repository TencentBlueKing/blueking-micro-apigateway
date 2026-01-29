# AGENTS.md

## Project Overview

BlueKing Micro API Gateway (BK Micro APIGateway) is a control plane for managing Apache APISIX data planes. This repository contains:

- **apiserver**: Go backend service (Gin framework) providing REST APIs for gateway management, details @src/apiserver/AGENTS.md
- **frontend**: Vue 3 frontend application for the management console, details @src/frontend/AGENTS.md

The project manages 11 types of APISIX resources: route, service, upstream, consumer, consumer_group, plugin_config, global_rule, plugin_metadata, protobuf, ssl, and stream_route.