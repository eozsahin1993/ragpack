// Package analytics runs read-only DuckDB SQL directly over the Parquet
// files pkg/telemetry writes, for the admin analytics endpoints. No ETL: a
// small named-query function per dashboard question, each backed by exactly
// one hardcoded, parameterized SQL string — not a generic query DSL. See
// pkg/api/analytics for the HTTP layer.
package analytics
