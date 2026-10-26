-- Mirrors the initialisation in OpenMetadata's official postgres image
-- (docker/postgresql/postgres-script.sql): the server connects as
-- openmetadata_user to a database it owns. The Airflow database from the
-- official script is omitted because this stack runs no ingestion service.
CREATE DATABASE openmetadata_db;
CREATE USER openmetadata_user WITH PASSWORD 'openmetadata_password';
ALTER DATABASE openmetadata_db OWNER TO openmetadata_user;
