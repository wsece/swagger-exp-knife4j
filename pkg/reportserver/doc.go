// Package reportserver provides local Web browsing of scan results and Knife4j debug proxy (service layer).
//
// Store: aggregates DB API records and on-disk api-docs session directories.
// Server: HTTP routes (/api/* report API, static pages, Knife4j doc.html, reverse proxy to real targets).
//
// Typical start: swagger-exp-knife4j report server --db-uri sqlite://... --api-doc-path ./output
//
// Web scan dispatch (home page "submit scan task"):
//
//   - POST /api/scan/tasks  multipart fields file / url / urls_text
//   - GET  /api/scan/tasks  task list; GET /api/scan/tasks/{id} poll status
// Results are written to the report server's db-uri and api-doc-path (same as CLI scan --write-db).
package reportserver
