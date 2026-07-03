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

function TerminalDots() {
  return (
    <div className={styles.header}>
      <span className={styles.dot} style={{ background: '#ff5f57' }} />
      <span className={styles.dot} style={{ background: '#febc2e' }} />
      <span className={styles.dot} style={{ background: '#28c840' }} />
    </div>
  );
}

function StartServiceSection() {
  return (
    <div className={styles.section}>
      <div className={styles.inner}>
        <div className={styles.text}>
          <div className={styles.label}>Step 1</div>
          <h2 className={styles.title}>Start the service</h2>
          <p className={styles.description}>
            Install the Ragpack CLI and start the full stack with one command. Ollama is included — no API keys needed to get started.
          </p>
        </div>
        <div className={styles.terminal}>
          <TerminalDots />
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
      </div>
    </div>
  );
}

function IntegrateAppSection() {
  return (
    <div className={styles.section}>
      <div className={styles.inner}>
        <div className={styles.text}>
          <div className={styles.label}>Step 2</div>
          <h2 className={styles.title}>Integrate with your app</h2>
          <p className={styles.description}>
            Install the Ragpack TypeScript SDK and start querying. Works in Node.js and the browser.
          </p>
        </div>
        <div className={styles.terminal}>
          <TerminalDots />
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
      </div>
    </div>
  );
}

export default function QuickStart() {
  return (
    <>
      <StartServiceSection />
      <IntegrateAppSection />
    </>
  );
}
