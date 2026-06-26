import Link from '@docusaurus/Link';
import styles from './styles.module.css';

export default function CTA() {
  return (
    <div className={styles.wrapper}>
      <div className={styles.inner}>
        <h2>Ready to ship your RAG pipeline?</h2>
        <p>One command. No cloud account. No token limits.</p>
        <Link className={styles.btn} to="/docs/getting-started">
          Read the docs →
        </Link>
      </div>
    </div>
  );
}
