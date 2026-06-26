import styles from './styles.module.css';

const items = [
  {
    eyebrow: 'Low footprint',
    title: 'Lightweight by design',
    description:
      'The Go backend idles at ~20MB RAM. A single static binary with no runtime dependencies means you can run a full RAG pipeline on the smallest machine your cloud provider offers — not a dedicated $200/mo AI instance.',
  },
  {
    eyebrow: 'Bring your own AI',
    title: 'You choose the model',
    description:
      'Ollama for fully local embeddings, OpenAI, or HuggingFace TEI. Switch providers by changing one env var — no code changes, no re-ingestion. Keep costs down with a local model or scale up when you need it.',
  },
  {
    eyebrow: 'No lock-in',
    title: 'Your data, your infra',
    description:
      'Everything runs in Docker. Vectors and metadata live on your own disk. No usage-based billing, no rate limits imposed by a third party, no vendor to negotiate with as you scale.',
  },
];

export default function ValueProps() {
  return (
    <div className={styles.wrapper}>
      <div className={styles.grid}>
        {items.map(({eyebrow, title, description}) => (
          <div key={title} className={styles.card}>
            <div className={styles.eyebrow}>{eyebrow}</div>
            <h3>{title}</h3>
            <p>{description}</p>
          </div>
        ))}
      </div>
    </div>
  );
}
