import { Highlight, themes } from 'prism-react-renderer';
import styles from './styles.module.css';

const CLI_LINES: { text: string; type: 'command' | 'success' | 'blank' }[] = [
  { type: 'command', text: 'npx ragpack init' },
  { type: 'success', text: '✔ Created .env.ragpack' },
  { type: 'blank',   text: '' },
  { type: 'command', text: 'npx ragpack start --profile ollama' },
  { type: 'success', text: '✔ Pulling nomic-embed-text...' },
  { type: 'success', text: '✔ API on :9000 · Admin UI on :3000' },
];

const SDK_CODE = `import { RagPack } from "ragpack-js";

const client = new RagPack({
  baseUrl: "http://localhost:9000",
  apiKey: "rp_...",
});

// Get an LLM answer grounded in your docs
const { answer } = await client
  .collection("my-docs")
  .rag({ query: "how does auth work?" });

console.log(answer);`;

function CLITerminal() {
  return (
    <div className={styles.terminal}>
      <div className={styles.header}>
        <span className={styles.dot} style={{ background: '#ff5f57' }} />
        <span className={styles.dot} style={{ background: '#febc2e' }} />
        <span className={styles.dot} style={{ background: '#28c840' }} />
        <span className={styles.title}>1. Install and run with Ollama</span>
      </div>
      <pre className={styles.body}>
        {CLI_LINES.map((line, i) => {
          if (line.type === 'blank') return <div key={i}>&nbsp;</div>;
          if (line.type === 'command') return (
            <div key={i}>
              <span className={styles.prompt}>$</span>
              <span className={styles.cmd}> {line.text}</span>
            </div>
          );
          return <div key={i} className={styles.success}>{line.text}</div>;
        })}
      </pre>
    </div>
  );
}

function SDKTerminal() {
  return (
    <div className={styles.terminal}>
      <div className={styles.header}>
        <span className={styles.dot} style={{ background: '#ff5f57' }} />
        <span className={styles.dot} style={{ background: '#febc2e' }} />
        <span className={styles.dot} style={{ background: '#28c840' }} />
        <span className={styles.title}>2. Add to your existing app</span>
      </div>
      <Highlight theme={themes.nightOwl} code={SDK_CODE} language="typescript">
        {({ className, style, tokens, getLineProps, getTokenProps }) => (
          <pre className={`${styles.body} ${className}`} style={{ ...style, background: '#13131a' }}>
            {tokens.map((line, i) => (
              <div key={i} {...getLineProps({ line })}>
                {line.map((token, key) => (
                  <span key={key} {...getTokenProps({ token })} />
                ))}
              </div>
            ))}
          </pre>
        )}
      </Highlight>
    </div>
  );
}

export default function Terminal() {
  return (
    <div className={styles.terminals}>
      <CLITerminal />
      <SDKTerminal />
    </div>
  );
}
