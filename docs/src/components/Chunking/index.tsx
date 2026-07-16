import styles from './styles.module.css';

const items = [
  { title: 'Context aware', description: 'Preserves headers as breadcrumbs.' },
  { title: 'Paragraph', description: 'Splits on paragraph boundaries.' },
  { title: 'Sliding window', description: 'Fixed 2000-char windows, 200-char overlap.' },
  { title: 'Row group', description: 'Row headers stay attached, for CSV/XLS.' },
  { title: 'Auto', description: 'Picks the right strategy from mime type.' },
];

export default function Chunking() {
  return (
    <div className={styles.wrapper}>
      <div className={styles.inner}>
        <div className={styles.eyebrow}>Chunking strategies</div>
        <h2>Picked per file type.</h2>
        <p className={styles.subtitle}>
          Auto mode picks the right strategy based on the file's mime type, or you can set one yourself.
        </p>
        <div className={styles.grid}>
          {items.map(({ title, description }) => (
            <div key={title} className={styles.card}>
              <h4>{title}</h4>
              <p>{description}</p>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
