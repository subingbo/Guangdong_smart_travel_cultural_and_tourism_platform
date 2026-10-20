# CloudBase MCP Tools Guide

This document lists all available Tencent CloudBase MCP (Model Context Protocol) tools and their functionalities. These tools allow AI assistants to interact directly with CloudBase resources.

## 1. Authentication & Environment

| Tool Name | Description |
|-----------|-------------|
| `mcp_cloudbase_login` | Log in to the CloudBase environment. Required before most operations. |
| `mcp_cloudbase_logout` | Log out from the CloudBase environment. |
| `mcp_cloudbase_envQuery` | Query environment info. Actions: `list` (all envs), `info` (current env), `domains` (safety domains), `hosting` (hosting config). |
| `mcp_cloudbase_envDomainManagement` | Manage safety domains. Actions: `create` (add domain), `delete` (remove domain). |
| `mcp_cloudbase_domainManagement` | Unified domain management. Actions: `create` (bind), `delete` (unbind), `check` (query), `modify` (update config). |

## 2. Database (NoSQL)

| Tool Name | Description |
|-----------|-------------|
| `mcp_cloudbase_readNoSqlDatabaseStructure` | Read collection structure. Actions: `listCollections`, `describeCollection`, `checkCollection`, `listIndexes`, `checkIndex`. |
| `mcp_cloudbase_writeNoSqlDatabaseStructure` | Modify structure. Actions: `createCollection`, `updateCollection`, `deleteCollection`. |
| `mcp_cloudbase_readNoSqlDatabaseContent` | Query data records. Supports `projection`, `query`, `sort`, `limit`, `offset`. |
| `mcp_cloudbase_writeNoSqlDatabaseContent` | Modify data records. Actions: `insert`, `update`, `delete`. Supports `isMulti`, `upsert`. |

## 3. Database (SQL / MySQL)

| Tool Name | Description |
|-----------|-------------|
| `mcp_cloudbase_executeReadOnlySQL` | Execute read-only SQL queries (SELECT). **Note**: Tables must have `_openid` column for ACL. |
| `mcp_cloudbase_executeWriteSQL` | Execute write SQL statements (INSERT, UPDATE, DELETE, CREATE, ALTER). **Note**: Tables must have `_openid` column. |

## 4. Data Modeling

| Tool Name | Description |
|-----------|-------------|
| `mcp_cloudbase_manageDataModel` | Query/List data models (Read-only). Actions: `list` (names), `get` (details/schema), `docs` (generate SDK docs). |
| `mcp_cloudbase_modifyDataModel` | Create/Update data models using **Mermaid classDiagram**. Supports MySQL/NoSQL. |

## 5. Cloud Functions

| Tool Name | Description |
|-----------|-------------|
| `mcp_cloudbase_getFunctionList` | List functions or get details. Actions: `list`, `detail`. |
| `mcp_cloudbase_createFunction` | Create a new cloud function. Supports env variables, VPC, layers, triggers (timer). |
| `mcp_cloudbase_updateFunctionCode` | Update code for an existing function. |
| `mcp_cloudbase_updateFunctionConfig` | Update configuration (timeout, env vars, VPC) for an existing function. |
| `mcp_cloudbase_invokeFunction` | Invoke a cloud function with parameters. |
| `mcp_cloudbase_getFunctionLogs` | Get function logs (metadata). |
| `mcp_cloudbase_getFunctionLogDetail` | Get detailed log content using RequestId. |
| `mcp_cloudbase_manageFunctionTriggers` | Create or delete function triggers (timer). |
| `mcp_cloudbase_createFunctionHTTPAccess` | Create HTTP access path for a function. |

## 6. Cloud Storage

| Tool Name | Description |
|-----------|-------------|
| `mcp_cloudbase_queryStorage` | Query storage info. Actions: `list` (files), `info` (meta), `url` (temp download link). |
| `mcp_cloudbase_manageStorage` | Manage files. Actions: `upload`, `download`, `delete`. |

## 7. Static Hosting

| Tool Name | Description |
|-----------|-------------|
| `mcp_cloudbase_uploadFiles` | Upload files/directories to static hosting. |
| `mcp_cloudbase_deleteFiles` | Delete files/directories from static hosting. |
| `mcp_cloudbase_findFiles` | Search for files in static hosting. |

## 8. Utilities

| Tool Name | Description |
|-----------|-------------|
| `mcp_cloudbase_searchWeb` | Perform internet search for real-time info. |

---

**Note**: All `mcp_cloudbase_` tools are designed to be used by the AI assistant directly.
