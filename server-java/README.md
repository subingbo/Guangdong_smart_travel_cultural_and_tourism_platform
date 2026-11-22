# server-java (frozen reference)

This Spring Boot service is **not** the running backend anymore. It was the
first implementation of the catalog / auth / blog / route / guide endpoints and
is kept only as a Java-vs-Go learning reference (see
[`../server-go/JAVA_VS_GO.md`](../server-go/JAVA_VS_GO.md)).

The website and mini program use `server-go` on `:8081`. Features added after
the initial milestone (comments, likes, follows, notifications, uploads, AI
copy/poster/menu OCR, the agent itinerary planner) exist only in Go.

To run it for comparison:

```powershell
cd server-java
mvn spring-boot:run   # :8080
```

It is behind the optional `legacy-java` Compose profile and is not started by
`docker compose up`.
