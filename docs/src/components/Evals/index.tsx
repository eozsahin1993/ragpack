import baseline from '../../../../eval/results/baseline.json';
import styles from './styles.module.css';

const round2 = (n: number) => n.toFixed(2);

const ragasStats = [
  { label: 'Faithfulness', value: round2(baseline.squad.faithfulness) },
  { label: 'Answer relevancy', value: round2(baseline.squad.answer_relevancy) },
  { label: 'Context precision', value: round2(baseline.squad.context_precision) },
  { label: 'Context recall', value: round2(baseline.squad.context_recall) },
];

const scifactStats = [
  { label: 'nDCG@10', value: round2(baseline.scifact['ndcg@10']) },
  { label: 'Recall@100', value: round2(baseline.scifact['recall@100']) },
];

export default function Evals() {
  return (
    <div id="evals" className={styles.wrapper}>
      <div className={styles.inner}>
        <div className={styles.eyebrow}>RAGAS evaluations</div>
        <h2>Measured, not marketed.</h2>
        <p className={styles.subtitle}>
          Scored against SQuAD 2.0 (30 real questions, not RagPack's own docs) through the actual{' '}
          <code>/rag</code> endpoint, judged by gpt-4o-mini.
        </p>
        <div className={styles.statGrid}>
          {ragasStats.map(({ label, value }) => (
            <div key={label} className={styles.statCard}>
              <div className={styles.statValue}>{value}</div>
              <div className={styles.statLabel}>{label}</div>
            </div>
          ))}
        </div>
        <p className={styles.subtitle}>Retrieval quality is also checked separately against BEIR's SciFact benchmark:</p>
        <div className={styles.statGridSmall}>
          {scifactStats.map(({ label, value }) => (
            <div key={label} className={styles.statCardSmall}>
              <div className={styles.statValueSmall}>{value}</div>
              <div className={styles.statLabel}>{label}</div>
            </div>
          ))}
        </div>
        <div className={styles.reproduce}>
          <span className={styles.comment}>reproduce it yourself:</span>
          <br />
          <span className={styles.prompt}>$</span> python3 eval/run_eval.py --api-key &lt;key&gt; --model gpt-4o-mini
        </div>
      </div>
    </div>
  );
}
