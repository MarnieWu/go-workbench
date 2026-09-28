# Business migration to write

Create `000001_initial.up.sql` here as described in issue 07. The runner rejects
an empty directory; this README is not a successful migration.

- Use six-digit positive versions: `000001_initial.up.sql`.
- Write transactional SQL without `BEGIN`, `COMMIT`, or `ROLLBACK`.
- Use unqualified table names so isolated test schemas remain isolated.
- Do not include `CREATE DATABASE`, `CREATE SCHEMA`, `SET search_path`, or
  `CREATE INDEX CONCURRENTLY` in these migrations.
- Never edit or remove an applied migration. Add a higher version instead.
- There is no automatic `down` command. Do not delete tables or Docker volumes
  as a rollback shortcut.
