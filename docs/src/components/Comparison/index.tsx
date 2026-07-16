import styles from './styles.module.css';

export default function Comparison() {
  return (
    <div id="why" className={styles.wrapper}>
      <div className={styles.inner}>
        <div className={styles.eyebrow}>Why RagPack</div>
        <h2>A RAG solution shouldn't cost a fortune.</h2>
        <p className={styles.subtitle}>
          Most stacks reach for LangChain and Pinecone. They're fast to get a demo running, but the
          catch shows up later: what they cost to run and maintain as you grow.
        </p>
        <div className={styles.grid}>
          <div className={styles.card}>
            <div className={styles.cardLabel}>The default stack</div>
            <div className={styles.cardTitle}>LangChain + Pinecone</div>
            <p className={styles.cardBody}>
              Managed vector DB billed per pod/index. Extra services to run and keep patched.
              Costs scale with data long before usage justifies it.
            </p>
          </div>
          <div className={`${styles.card} ${styles.cardAccent}`}>
            <div className={styles.cardLabelAccent}>RagPack</div>
            <div className={styles.cardTitle}>One Go binary</div>
            <p className={styles.cardBody}>
              Storage is embedded directly in the process: LanceDB for vectors, SQLite for metadata.
              Nothing else to run, and it's comfortable on a $5 VPS or an EC2 t3.micro.
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
