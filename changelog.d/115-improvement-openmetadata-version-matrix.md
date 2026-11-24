The acceptance suite now covers OpenMetadata 1.13.6 and the 2.0.5–2.0.9
releases, the full window in `.github/openmetadata-versions.txt`.
`make testacc-matrix` runs the same matrix locally; `make testacc-docker` still
runs one version, the newest by default, and accepts `OPENMETADATA_VERSION` to
select another.
