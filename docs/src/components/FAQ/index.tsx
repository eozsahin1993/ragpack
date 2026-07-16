import styles from './styles.module.css';

const items = [
  {
    q: 'Do I need a GPU?',
    a: "No. Ollama can run small embedding models on CPU, and HuggingFace TEI or any OpenAI-compatible provider are supported if you'd rather not run inference yourself.",
  },
  {
    q: 'Where does my data live?',
    a: 'Inside your own Docker volume: vectors in an embedded LanceDB, metadata in SQLite. Nothing leaves your infrastructure unless you choose a cloud embedding or LLM provider.',
  },
  {
    q: 'What file types can I ingest?',
    a: '.txt, .md, .html, .pdf, .docx, .pptx, .xlsx, .csv, .json, .xml, via URL, S3, or direct file upload.',
  },
  {
    q: 'Can I use OpenAI instead of running models locally?',
    a: "Yes. Embeddings support Ollama, HuggingFace TEI, or any OpenAI-compatible provider. The RAG answer step supports OpenAI, Ollama, or Anthropic.",
  },
  {
    q: 'Is there a hosted version?',
    a: "Not currently. RagPack is self-hosted by design, via Docker Compose or npx ragpack, so it runs entirely on your own infrastructure.",
  },
];

export default function FAQ() {
  return (
    <div className={styles.wrapper}>
      <div className={styles.inner}>
        <h2>FAQ</h2>
        <div className={styles.list}>
          {items.map(({ q, a }) => (
            <details key={q} className={styles.item}>
              <summary>{q}</summary>
              <p>{a}</p>
            </details>
          ))}
        </div>
      </div>
    </div>
  );
}
