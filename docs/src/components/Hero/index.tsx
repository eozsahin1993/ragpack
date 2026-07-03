import Link from '@docusaurus/Link';
import { track } from '@site/src/lib/analytics';
import styles from './styles.module.css';

export default function Hero() {
  return (
    <div className={styles.hero}>
      <div className={styles.heroInner}>
        <div className={styles.badge}>Open Source · MIT</div>
        <h1 className={styles.heroTitle}>
          RAG and semantic search,<br />
          <span className={styles.heroTitleAccent}>self-hosted in minutes</span>
        </h1>
        <p className={styles.heroSubtitle}>
          Ragpack adds RAG and semantic search to your existing app in minutes — no new infrastructure to manage.
          Self-hosted on your own machine, bring your own AI with Ollama or OpenAI.
        </p>
        <div className={styles.heroActions}>
          <Link className={styles.btnPrimary} to="/docs/getting-started" onClick={() => track('click_get_started', 'hero')}>
            Get started →
          </Link>
          <a
            className={styles.btnSecondary}
            href="https://github.com/eozsahin1993/ragpack"
            target="_blank"
            rel="noopener noreferrer"
            onClick={() => track('click_github', 'hero')}
          >
            View on GitHub
          </a>
        </div>
      </div>
    </div>
  );
}
