---
sidebar_position: 6
---

# Jobs

When you ingest a document, RagPack queues an ingestion job and processes it asynchronously. A `Job` is returned immediately so you can track progress without blocking.

## Job statuses

| Status | Meaning |
|---|---|
| `pending` | Queued, not yet started |
| `processing` | Fetching, parsing, chunking, and embedding |
| `complete` | Document is indexed and ready to query |
| `failed` | Something went wrong — check `job.error` for details |

## List jobs for a collection

```ts
const jobs = await collection.jobs.list();

for (const job of jobs) {
  console.log(job.id, job.status, job.file_uri);
}
```

To filter by status, use the REST API directly with a `?status=` query parameter:

```bash
curl "http://localhost:9000/api/v1/collections/my-docs/jobs?status=failed" \
  -H "Authorization: Bearer $RAGPACK_API_KEY"
```

Valid values: `pending`, `processing`, `complete`, `failed`.

## Get a single job

```ts
const job = await collection.jobs.get(jobId);
console.log(job.status, job.error);
```

## Wait for a job to complete

`waitUntilComplete` polls until the job reaches `complete` or `failed`, then resolves or throws accordingly.

```ts
const job = await collection.ingest(file);

await collection.jobs.waitUntilComplete(job.id);
console.log("ready to query");
```

You can tune the polling interval and timeout:

```ts
await collection.jobs.waitUntilComplete(job.id, {
  pollIntervalMs: 2000,   // check every 2 seconds (default: 1500)
  timeoutMs: 60_000,      // give up after 1 minute (default: 5 minutes)
});
```

Throws a `RagPackError` with status `422` if the job fails, or `408` if the timeout is exceeded.

## Job object

| Field | Type | Description |
|---|---|---|
| `id` | `string` | Unique job identifier |
| `file_uri` | `string` | Source URI of the ingested document |
| `mime_type` | `string` | Detected MIME type |
| `extra_json` | `string \| undefined` | Freeform JSON metadata carried from the ingest request onto the resulting document |
| `metadata` | `string \| undefined` | Typed metadata field values carried from the ingest request |
| `intent` | `string` | `ingest` \| `refresh` |
| `force` | `boolean` | Whether this job was forced to re-run despite a matching existing document |
| `status` | `string` | `pending` \| `processing` \| `complete` \| `failed` |
| `error` | `string \| undefined` | Error message if `status` is `failed` |
| `executed_at` | `string \| undefined` | When the worker picked up this job; unset while still `pending` |
| `created_at` | `string` | ISO 8601 timestamp |
| `updated_at` | `string` | ISO 8601 timestamp of last status change |
