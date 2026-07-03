import Link from '@docusaurus/Link';
import { track } from '@site/src/lib/analytics';
import styles from './styles.module.css';

export default function CTA() {
  return (
    <div className={styles.wrapper}>
      <div className={styles.inner}>
        <h2>Start building with Ragpack today</h2>
        <p>Follow the quickstart guide and you'll be querying in minutes.</p>
        <Link className={styles.btn} to="/docs/getting-started" onClick={() => track('click_get_started', 'cta')}>
          Click to get started →
        </Link>
      </div>
    </div>
  );
}
