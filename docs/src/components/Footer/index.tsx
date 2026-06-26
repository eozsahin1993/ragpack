import { track } from '@site/src/lib/analytics';
import styles from './styles.module.css';

export default function Footer() {
  return (
    <footer className={styles.footer}>
      MIT License © {new Date().getFullYear()} RagPack
      <a href="/docs/getting-started" onClick={() => track('click_docs', 'footer')}>Docs</a>
      <a href="https://github.com/eozsahin1993/ragpack" target="_blank" rel="noopener noreferrer" onClick={() => track('click_github', 'footer')}>GitHub</a>
      <a href="https://www.npmjs.com/package/ragpack" target="_blank" rel="noopener noreferrer" onClick={() => track('click_npm', 'footer')}>npm</a>
    </footer>
  );
}
