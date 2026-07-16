import styles from './styles.module.css';

const FORMATS = ['.txt', '.md', '.html', '.pdf', '.docx', '.pptx', '.xlsx', '.csv', '.json', '.xml'];

export default function Formats() {
  return (
    <div id="formats" className={styles.wrapper}>
      <div className={styles.inner}>
        <div className={styles.eyebrow}>Supported formats</div>
        <div className={styles.chips}>
          {FORMATS.map((ext) => (
            <div key={ext} className={styles.chip}>{ext}</div>
          ))}
        </div>
      </div>
    </div>
  );
}
