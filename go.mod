module github.com/aplbrain/neuPrintHTTP

go 1.14

require (
	github.com/aplbrain/echo-secure v0.0.0-20200415203207-2b260febe9e4
	github.com/blang/semver v3.5.1+incompatible
	github.com/connectome-neuprint/neuPrintHTTP v1.2.1
	github.com/dgraph-io/badger v1.6.1
	github.com/eitanflor/echo-secure v0.0.0-20200723201440-78ce1e3109cd
	github.com/gorilla/sessions v1.2.0 // indirect
	github.com/knightjdr/hclust v1.0.2
	github.com/labstack/echo v3.3.10+incompatible
	github.com/labstack/echo-contrib v0.9.0 // indirect
	github.com/labstack/echo/v4 v4.1.16
	github.com/labstack/gommon v0.3.0
	github.com/satori/go.uuid v1.2.0 // indirect
	github.com/valyala/fasttemplate v1.2.0
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d // indirect
)

// resolves uuid error (argument error)
replace github.com/satori/go.uuid v1.2.0 => github.com/satori/go.uuid v1.2.1-0.20181028125025-b2ce2384e17b
