import Link from '@docusaurus/Link';
import Terminal from '@site/src/components/Terminal';
import styles from './styles.module.css';

export default function Hero() {
  return (
    <div className={styles.hero}>
      <div className={styles.heroInner}>
        <div className={styles.heroLeft}>
          <div className={styles.badge}>Open Source · MIT</div>
          <h1 className={styles.heroTitle}>
            RAG and semantic search,<br />
            <span className={styles.heroTitleAccent}>self-hosted in minutes</span>
          </h1>
          <p className={styles.heroSubtitle}>
            Add semantic search and RAG to your existing app in minutes.
            Self-hosted on your own infra — bring your own AI with Ollama or OpenAI.
          </p>
          <div className={styles.heroActions}>
            <Link className={styles.btnPrimary} to="/docs/getting-started">
              Get started →
            </Link>
            <a
              className={styles.btnSecondary}
              href="https://github.com/eozsahin1993/ragpack"
              target="_blank"
              rel="noopener noreferrer"
            >
              View on GitHub
            </a>
          </div>
        </div>
        <div className={styles.heroRight}>
          <Terminal />
        </div>
      </div>
    </div>
  );
}
