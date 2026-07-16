import Link from '@docusaurus/Link';
import { track } from '@site/src/lib/analytics';
import styles from './styles.module.css';

function CopyChip() {
  const copy = () => {
    navigator.clipboard.writeText('npx ragpack init');
    track('copy_install_command', 'hero');
  };
  return (
    <button className={styles.copyChip} onClick={copy} type="button">
      <span className={styles.prompt}>$</span> npx ragpack init
    </button>
  );
}

export default function Hero() {
  return (
    <div className={styles.hero}>
      <div className={styles.heroInner}>
        <img src="/img/logo.svg" alt="RagPack" width={64} height={64} className={styles.logo} />
        <div className={styles.badge}>Open-source · Self-hosted RAG</div>
        <h1 className={styles.heroTitle}>RAG infrastructure built for performance and low cost.</h1>
        <p className={styles.heroSubtitle}>
          A single Go binary with embedded storage: LanceDB for vectors, SQLite for metadata.
          Add a RAG pipeline to your existing project in minutes.
        </p>
        <div className={styles.heroActions}>
          <Link className={styles.btnPrimary} to="/docs/getting-started" onClick={() => track('click_get_started', 'hero')}>
            Get started
          </Link>
          <CopyChip />
        </div>
      </div>
    </div>
  );
}
