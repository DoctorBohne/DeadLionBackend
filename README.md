# DeadLionBackend

Ein REST-Backend für ein Aufgabenverwaltungssystem (Task-Management). Das Projekt bietet Funktionen zur Verwaltung von Aufgaben und Teilaufgaben, Kanban-Boards, Swimlane-Pools sowie Risikoberechnung anhand von Abgabefristen. Die Authentifizierung erfolgt über JWT-Token eines OpenID-Connect-Providers (Clerk).

## Inhaltsverzeichnis

- [Technologie-Stack](#technologie-stack)
- [Voraussetzungen](#voraussetzungen)
- [Konfiguration](#konfiguration)
- [Software starten](#software-starten)
- [Software testen / ausprobieren](#software-testen--ausprobieren)
- [API-Dokumentation](#api-dokumentation)
- [Projektstruktur](#projektstruktur)

---

## Technologie-Stack

| Komponente        | Technologie                        |
|-------------------|------------------------------------|
| Sprache           | Go 1.24                            |
| Web-Framework     | Gin                                |
| Datenbank         | PostgreSQL 16                      |
| ORM               | GORM (mit AutoMigrate)             |
| Authentifizierung | JWT via Clerk (OpenID Connect)     |
| Container         | Docker / Docker Compose            |

---

## Voraussetzungen

- [Docker](https://docs.docker.com/get-docker/) und [Docker Compose](https://docs.docker.com/compose/install/) **oder** eine lokale Go-Installation (≥ 1.24) mit einem laufenden PostgreSQL-16-Server
- Ein [Clerk](https://clerk.com)-Konto zur Bereitstellung der JWT-Konfiguration (`CLERK_ISS_URL` und `CLERK_JWKS_URL`)

---

## Konfiguration

Die Anwendung wird vollständig über Umgebungsvariablen konfiguriert. Für den lokalen Betrieb kann eine `.env`-Datei verwendet werden.

### Datenbankverbindung

| Variable        | Beschreibung                         | Standardwert |
|-----------------|--------------------------------------|--------------|
| `DB_HOST`       | Hostname des PostgreSQL-Servers      | –            |
| `DB_PORT`       | Port des PostgreSQL-Servers          | `5432`       |
| `DB_USER`       | Datenbankbenutzer                    | –            |
| `DB_PASS`       | Passwort des Datenbankbenutzers      | –            |
| `DB_NAME`       | Name der Datenbank                   | –            |
| `DB_SSLMODE`    | SSL-Modus (`disable`, `require`, …)  | `disable`    |
| `DB_AUTOMIGRATE`| Datenbankschema automatisch anlegen  | `true`       |

### Authentifizierung (Clerk)

| Variable         | Beschreibung                                  |
|------------------|-----------------------------------------------|
| `CLERK_ISS_URL`  | OpenID-Connect-Issuer-URL (**Pflichtfeld**)    |
| `CLERK_JWKS_URL` | URL des JSON-Web-Key-Sets (**Pflichtfeld**)    |

---

## Software starten

### Variante 1 – Docker Compose (empfohlen)

Docker Compose startet sowohl die Anwendung als auch PostgreSQL in einem Schritt.

1. Repository klonen:
   ```bash
   git clone https://github.com/DoctorBohne/DeadLionBackend.git
   cd DeadLionBackend
   ```

2. `.env`-Datei anlegen (Vorlage anpassen):
   ```bash
   cat > .env <<'EOF'
   DB_HOST=db
   DB_PORT=5432
   DB_USER=deadlion
   DB_PASS=geheimes_passwort
   DB_NAME=deadlion
   DB_SSLMODE=disable
   DB_AUTOMIGRATE=true
   CLERK_ISS_URL=https://<deine-clerk-domain>.clerk.accounts.dev
   CLERK_JWKS_URL=https://<deine-clerk-domain>.clerk.accounts.dev/.well-known/jwks.json
   POSTGRES_USER=deadlion
   POSTGRES_PASSWORD=geheimes_passwort
   POSTGRES_DB=deadlion
   EOF
   ```
   > **Hinweis:** Die `docker-compose.yml` lädt Umgebungsvariablen standardmäßig aus `/opt/deadlion-secrets/.env`. Für die lokale Entwicklung kann der Pfad in der `docker-compose.yml` angepasst oder die Variablen direkt als `environment:` eingetragen werden.

3. Container starten:
   ```bash
   docker compose up --build
   ```

4. Der Server ist nun erreichbar unter **http://localhost:8081** (Host-Port 8081 → Container-Port 8080).

### Variante 2 – Direkt mit Go

Voraussetzung: Go ≥ 1.24 und ein laufender PostgreSQL-Server.

```bash
# Abhängigkeiten herunterladen
go mod download

# Server bauen und starten
go run ./cmd/server
```

Der Server startet auf Port **8080**.

---

## Software testen / ausprobieren

### Health-Check

Der einfachste Weg, die Erreichbarkeit der API zu prüfen:

```bash
curl http://localhost:8081/healthz
```

Erwartete Antwort: HTTP 200

### Authentifizierte Anfragen

Alle Endpunkte unter `/api/v1/` erfordern einen gültigen JWT-Bearer-Token von Clerk. Den Token aus der Clerk-Dashboard-Oberfläche (oder aus einer integrierten Frontend-Anwendung) als Header übergeben:

```bash
export TOKEN="<dein-jwt-token>"
```

**Eigenes Benutzerprofil abrufen oder anlegen:**

```bash
curl -H "Authorization: Bearer $TOKEN" http://localhost:8081/api/v1/me
```

**Aufgaben auflisten:**

```bash
curl -H "Authorization: Bearer $TOKEN" http://localhost:8081/api/v1/tasks
```

**Neue Aufgabe erstellen:**

```bash
curl -X POST http://localhost:8081/api/v1/tasks \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Beispielaufgabe",
    "description": "Beschreibung der Aufgabe",
    "deadline": "2026-12-31T23:59:59Z"
  }'
```

**Abgaben mit Risikoliste abrufen:**

```bash
curl -H "Authorization: Bearer $TOKEN" http://localhost:8081/api/v1/abgaben/risklist
```

Eine vollständige Übersicht aller Endpunkte, Parameter und Antwortformate befindet sich in der [API-Dokumentation](#api-dokumentation).

---

## API-Dokumentation

Die vollständige API-Spezifikation liegt als **OpenAPI 3.0.3**-Datei im Repository vor:

📄 [`openapi.yml`](./openapi.yml)

Die Datei kann mit einem beliebigen Swagger/OpenAPI-Viewer geöffnet werden, z. B.:

- **Swagger UI (online):** [editor.swagger.io](https://editor.swagger.io) → Dateiinhalt einfügen
- **Swagger UI (lokal via Docker):**
  ```bash
  docker run -p 8082:8080 \
    -e SWAGGER_JSON=/api/openapi.yml \
    -v "$(pwd)/openapi.yml:/api/openapi.yml" \
    swaggerapi/swagger-ui
  ```
  Danach unter **http://localhost:8082** aufrufen.
- **VS Code:** Erweiterung [OpenAPI (Swagger) Editor](https://marketplace.visualstudio.com/items?itemName=42Crunch.vscode-openapi) installieren und `openapi.yml` öffnen.

---

## Projektstruktur

```
DeadLionBackend/
├── cmd/server/          # Einstiegspunkt der Anwendung (main.go)
├── internal/
│   ├── auth/            # JWT-Verifikation (Clerk-Integration)
│   ├── db/              # Datenbankverbindung und Migrationen
│   ├── http/
│   │   ├── handler/     # HTTP-Handler (Tasks, Boards, Abgaben, …)
│   │   ├── middleware/  # Authentifizierungs-Middleware
│   │   └── router.go    # Routing-Definitionen
│   ├── models/          # GORM-Datenbankmodelle
│   ├── repositories/    # Datenzugriffsschicht
│   ├── services/        # Geschäftslogik
│   └── abgabe/          # Abgaben- und Risikoberechnung
├── openapi.yml          # OpenAPI-3.0.3-Spezifikation
├── docker-compose.yml   # Lokale Entwicklungsumgebung
└── Dockerfile           # Multi-Stage-Build (distroless)
```
