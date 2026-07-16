import { track } from '@site/src/lib/analytics';
import styles from './styles.module.css';

function CopyChip() {
  const copy = () => {
    navigator.clipboard.writeText('npx ragpack init');
    track('copy_install_command', 'cta');
  };
  return (
    <button className={styles.copyChip} onClick={copy} type="button">
      <span className={styles.prompt}>$</span> npx ragpack init
    </button>
  );
}

export default function CTA() {
  return (
    <div className={styles.wrapper}>
      <div className={styles.inner}>
        <h2>Run your own RAG stack today.</h2>
        <p>Add a RAG pipeline to your existing project in minutes, no cloud bill required.</p>
        <div className={styles.actions}>
          <a
            className={styles.btn}
            href="https://github.com/eozsahin1993/ragpack"
            target="_blank"
            rel="noopener noreferrer"
            onClick={() => track('click_github', 'cta')}
          >
            View on GitHub
          </a>
          <CopyChip />
        </div>
      </div>
    </div>
  );
}
