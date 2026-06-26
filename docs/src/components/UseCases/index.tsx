import styles from './styles.module.css';

const items = [
  {
    title: 'Docs chatbot',
    description: 'Let users ask questions about your documentation and get answers grounded in your content.',
    example: '"How do I reset my password?"',
  },
  {
    title: 'Knowledge base search',
    description: 'Make internal wikis, SOPs, and runbooks searchable with natural language — not just keywords.',
    example: '"What\'s our refund policy for enterprise customers?"',
  },
  {
    title: 'AI customer support',
    description: 'Ground your support chatbot in your own product docs so it only answers from what you\'ve written.',
    example: '"Why is my invoice showing the wrong amount?"',
  },
  {
    title: 'Code search',
    description: 'Ingest your codebase and let developers find relevant files, functions, or patterns semantically.',
    example: '"Where do we handle Stripe webhook retries?"',
  },
];

export default function UseCases() {
  return (
    <div className={styles.wrapper}>
      <div className={styles.inner}>
        <h2>What can you build?</h2>
        <div className={styles.grid}>
          {items.map(({title, description, example}) => (
            <div key={title} className={styles.card}>
              <h4>{title}</h4>
              <p>{description}</p>
              <div className={styles.example}>{example}</div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
