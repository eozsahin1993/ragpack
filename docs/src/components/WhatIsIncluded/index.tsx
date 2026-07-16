import styles from './styles.module.css';

const items = [
  { title: 'Semantic search', description: "RAG endpoints with prompts baked in, so there's no prompt engineering needed to get started." },
  { title: 'Hybrid retrieval', description: 'BM25 keyword search + vector search, merged with Reciprocal Rank Fusion using customizable weights.' },
  { title: 'Bring your own model', description: 'Ollama or TEI for fully local inference, or any OpenAI-compatible provider.' },
  { title: 'Flexible ingestion', description: 'Ingest documents from URLs, file uploads, or S3.' },
  { title: 'Client SDKs', description: 'JS/TS SDKs for dropping into an existing app.' },
  { title: 'Smart refresh', description: 'Runs on a timer, only re-embeds and re-inserts chunks that actually changed.' },
  { title: 'Admin dashboard', description: 'Manage collections, documents, and queries in one place.' },
  { title: 'Built-in analytics', description: 'Embedding/LLM costs, usage metrics, and query evaluations.' },
];

function Card({ title, description }: { title: string; description: string }) {
  return (
    <div className={styles.card}>
      <h4>{title}</h4>
      <p>{description}</p>
    </div>
  );
}

function MongoFilterCard() {
  return (
    <div className={styles.card}>
      <h4>Mongo-style filters</h4>
      <p>Filter on custom document properties.</p>
      <pre className={styles.code}>{`{"$and": [
  {"category": {"$in": ["research","legal"]}},
  {"score": {"$gt": 0.8}}
]}`}</pre>
    </div>
  );
}

export default function WhatIsIncluded() {
  return (
    <div id="features" className={styles.wrapper}>
      <div className={styles.inner}>
        <div className={styles.eyebrow}>Features</div>
        <h2>Everything you need to ship RAG.</h2>
        <div className={styles.grid}>
          {items.slice(0, 5).map((item) => <Card key={item.title} {...item} />)}
          <MongoFilterCard />
          {items.slice(5).map((item) => <Card key={item.title} {...item} />)}
        </div>
      </div>
    </div>
  );
}
