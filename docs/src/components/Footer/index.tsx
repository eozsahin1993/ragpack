import styles from './styles.module.css';

export default function Footer() {
  return (
    <footer className={styles.footer}>
      MIT License © {new Date().getFullYear()} RagPack
      <a href="/docs/getting-started">Docs</a>
      <a href="https://github.com/eozsahin1993/ragpack" target="_blank" rel="noopener noreferrer">GitHub</a>
      <a href="https://www.npmjs.com/package/ragpack" target="_blank" rel="noopener noreferrer">npm</a>
    </footer>
  );
}
