# Java / Go 后端对照说明

两套后端实现 [docs/api/openapi.yaml](../docs/api/openapi.yaml) 的同一套接口，共用同一个 MySQL 库。这份文档解释两边概念如何对应，方便用熟悉的 Java 概念学 Go。

## 分层对照

| 概念 | Spring Boot (server-java) | Gin (server-go) |
| :--- | :--- | :--- |
| 入口 | `TravelGdApplication` + `@SpringBootApplication` | `main.go`（手动组装依赖） |
| 控制器 | `@RestController`（controller 包） | `handler` 包的 struct 方法 |
| 业务层 | `@Service`（service 包） | `service` 包的 struct |
| 数据访问 | MyBatis `@Mapper` 接口 + XML | `repo` 包的 struct + sqlx 手写 SQL |
| 依赖注入 | Spring 容器自动注入 | `main.go` 里手工 new 再传参（构造函数注入） |
| DTO 校验 | `jakarta.validation`（`@NotBlank` 等） | handler 里 `ShouldBindJSON` 后手动判断 |
| JSON 列映射 | MyBatis `TypeHandler`（handler 包） | `model/json_types.go` 的 `Scan`/`Value` |
| 密码哈希 | `BCryptPasswordEncoder` | `golang.org/x/crypto/bcrypt` |
| JWT | jjwt（security 包） | `golang-jwt/jwt/v5`（internal/auth） |
| 全局异常 | `@RestControllerAdvice` | 每个入口调用 `common.Fail(c, err)` |
| 响应信封 | `ApiResponse<T>` 泛型类 | `common.Envelope` struct |
| 事务 | `@Transactional` | `db.Beginx()` / `tx.Commit()`（本期未用到跨表事务） |

## 同一功能的并排阅读

以"收藏景点"为例：

- Java：[FavoriteController.java](../server-java/src/main/java/com/travelgd/controller/FavoriteController.java) → [FavoriteService.java](../server-java/src/main/java/com/travelgd/service/FavoriteService.java) → [FavoriteMapper.java](../server-java/src/main/java/com/travelgd/mapper/FavoriteMapper.java) + [FavoriteMapper.xml](../server-java/src/main/resources/mapper/FavoriteMapper.xml)
- Go：[me_handler.go](../server-go/internal/handler/me_handler.go)（FavoriteHandler 部分）→ [catalog_service.go](../server-go/internal/service/catalog_service.go)（FavoriteService 部分）→ [favorite_repo.go](../server-go/internal/repo/favorite_repo.go)

差异要点：

1. **SQL 写在哪**：Java 放在 XML 里，Go 直接写在 repo 方法里，两边 SQL 语句几乎逐行相同，可以对照看。
2. **错误传播**：Java 抛 `ApiException` 由注解兜底；Go 把 `*common.ApiError` 作为返回值一路往上传，handler 里统一转 HTTP 响应。Go 的错误是值，不是异常。
3. **上下文**：Go 的每个数据库/AI 调用都带 `context.Context`，用于超时与取消；Java 里对应超时写在 HttpClient 上。
4. **空值**：Java 用 `null`；Go 里"查不到"用 `nil 指针 + sql.ErrNoRows` 处理，repo 返回 `nil, nil` 表示不存在，service 层再转成 404。
5. **接口可选性**：两边都在"路由分组"上声明公开/受保护——Java 用 `CurrentUserResolver` + `@CurrentUser`，Go 用 `RequireAuth` / `OptionalAuth` 中间件。

## 运行方式

- Java：`cd server-java && mvn spring-boot:run`（端口 8080）
- Go：`cd server-go && go run .`（端口 8081）
- 前端通过 `VITE_API_BASE_URL` 指向任意一个；两边响应字段一致，前端无感切换。

## 依赖清单

| 用途 | Java | Go |
| :--- | :--- | :--- |
| Web 框架 | spring-boot-starter-web | gin-gonic/gin |
| 数据库 | mybatis-spring-boot-starter + mysql-connector-j | jmoiron/sqlx + go-sql-driver/mysql |
| JWT | jjwt | golang-jwt/jwt/v5 |
| 密码 | spring-security-crypto（BCrypt） | golang.org/x/crypto/bcrypt |
| HTTP 客户端 | java.net.http.HttpClient | net/http |
