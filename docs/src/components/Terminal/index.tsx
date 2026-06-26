import styles from './styles.module.css';

const CODE = `$ npx ragpack init && npx ragpack start
✔ API on :9000 · Admin UI on :3000

# Drop into any existing app
import { RagPack } from "ragpack-js";

const client = new RagPack({
  baseUrl: "http://localhost:9000",
});

const col = client.collection("my-docs");

await col.ingest({
  uri: "https://your-docs-site.com/guide",
});

const { answer } = await col.rag({
  query: "how does auth work?",
});`;

export default function Terminal() {
  return (
    <div className={styles.terminal}>
      <div className={styles.header}>
        <span className={styles.dot} style={{background: '#ff5f57'}} />
        <span className={styles.dot} style={{background: '#febc2e'}} />
        <span className={styles.dot} style={{background: '#28c840'}} />
      </div>
      <pre className={styles.body}>{CODE}</pre>
    </div>
  );
}
