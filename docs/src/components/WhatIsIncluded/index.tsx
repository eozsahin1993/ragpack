import styles from './styles.module.css';

const items = [
  {icon: '⚡', title: 'REST API', description: 'Collections, ingest, query, and document management — all via a clean REST API.'},
  {icon: '📦', title: 'JS / TS SDK', description: 'Typed client for Node and the browser. Drop RAG into your app in a few lines.'},
  {icon: '🖥️', title: 'Admin UI', description: 'Built-in Next.js interface to manage collections, monitor jobs, and run queries.'},
  {icon: '🔌', title: 'Multiple embedding providers', description: 'Ollama, OpenAI, and HuggingFace TEI supported out of the box.'},
  {icon: '📄', title: 'Supported formats', description: 'Markdown, HTML, PDF, and plain text — via URL, S3, or file upload.'},
  {icon: '✂️', title: 'Chunking strategies', description: 'Format-aware chunking out of the box. Configurable chunk size and overlap.'},
];

export default function WhatIsIncluded() {
  return (
    <div className={styles.wrapper}>
      <div className={styles.inner}>
        <h2>What's included</h2>
        <div className={styles.grid}>
          {items.map(({icon, title, description}) => (
            <div key={title} className={styles.card}>
              <div className={styles.icon}>{icon}</div>
              <h4>{title}</h4>
              <p>{description}</p>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
