import styles from './styles.module.css';

const items = [
  {
    eyebrow: 'Ship faster',
    title: 'RAG in minutes, not days',
    description:
      'Document parsing, chunking, embedding, and semantic search — all handled. Add RAG to an existing app in an afternoon, not a sprint.',
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z" />
      </svg>
    ),
    accent: '#6366f1',
  },
  {
    eyebrow: 'High performance · Low cost',
    title: 'One binary, no external services',
    description:
      'Go backend with LanceDB for vectors and SQLite for metadata — all embedded, no separate services to run. Self-host on any machine, any cloud.',
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <circle cx="12" cy="12" r="10" />
        <path d="M12 6v12M9 9h4.5a2.5 2.5 0 0 1 0 5H9" />
      </svg>
    ),
    accent: '#10b981',
  },
  {
    eyebrow: 'Bring your own AI',
    title: 'Local models or any OpenAI-compatible provider',
    description:
      'First-class support for Ollama and HuggingFace TEI for fully local inference. Or plug in OpenAI or any OpenAI-compatible API.',
    icon: (
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <rect x="4" y="4" width="16" height="16" rx="2" />
        <rect x="9" y="9" width="6" height="6" />
        <path d="M9 2v2M15 2v2M9 20v2M15 20v2M2 9h2M2 15h2M20 9h2M20 15h2" />
      </svg>
    ),
    accent: '#8b5cf6',
  },
];

export default function ValueProps() {
  return (
    <div className={styles.wrapper}>
      <div className={styles.grid}>
        {items.map(({ eyebrow, title, description, icon, accent }) => (
          <div key={title} className={styles.card} style={{ '--accent': accent } as React.CSSProperties}>
            <div className={styles.iconWrap} style={{ color: accent }}>
              {icon}
            </div>
            <div className={styles.eyebrow}>{eyebrow}</div>
            <h3>{title}</h3>
            <p>{description}</p>
          </div>
        ))}
      </div>
    </div>
  );
}
