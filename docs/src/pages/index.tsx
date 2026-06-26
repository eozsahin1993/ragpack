import type {ReactNode} from 'react';
import {useEffect, useState} from 'react';
import Link from '@docusaurus/Link';
import Layout from '@theme/Layout';
import styles from './index.module.css';

const GITHUB_REPO = 'eozsahin1993/ragpack';

function GitHubStars() {
  const [stars, setStars] = useState<number | null>(null);

  useEffect(() => {
    fetch(`https://api.github.com/repos/${GITHUB_REPO}`)
      .then((r) => r.json())
      .then((d) => setStars(d.stargazers_count))
      .catch(() => {});
  }, []);

  return (
    <a
      href={`https://github.com/${GITHUB_REPO}`}
      target="_blank"
      rel="noopener noreferrer"
      className={styles.githubBtn}
    >
      <svg height="16" viewBox="0 0 16 16" width="16" aria-hidden="true" fill="currentColor">
        <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z" />
      </svg>
      Star on GitHub
      {stars !== null && <span className={styles.starCount}>{stars.toLocaleString()}</span>}
    </a>
  );
}

const features = [
  {
    eyebrow: 'Low footprint',
    title: 'Lightweight by design',
    description:
      'The Go backend idles at ~20MB RAM. A single static binary with no runtime dependencies means you can run a full RAG pipeline on the smallest machine your cloud provider offers — not a dedicated $200/mo AI instance.',
  },
  {
    eyebrow: 'Bring your own AI',
    title: 'You choose the model',
    description:
      'Ollama for fully local embeddings, OpenAI, or HuggingFace TEI. Switch providers by changing one env var — no code changes, no re-ingestion. Keep costs down with a local model or scale up when you need it.',
  },
  {
    eyebrow: 'No lock-in',
    title: 'Your data, your infra',
    description:
      'Everything runs in Docker. Vectors and metadata live on your own disk. No usage-based billing, no rate limits imposed by a third party, no vendor to negotiate with as you scale.',
  },
];

const included = [
  {
    icon: '⚡',
    title: 'REST API',
    description: 'Collections, ingest, query, and document management — all via a clean REST API.',
  },
  {
    icon: '📦',
    title: 'JS / TS SDK',
    description: 'Typed client for Node and the browser. Drop RAG into your app in a few lines.',
  },
  {
    icon: '🖥️',
    title: 'Admin UI',
    description: 'Built-in Next.js interface to manage collections, monitor jobs, and run queries.',
  },
  {
    icon: '🔌',
    title: 'Multiple embedding providers',
    description: 'Ollama, OpenAI, and HuggingFace TEI supported out of the box.',
  },
  {
    icon: '📄',
    title: 'Supported formats',
    description: 'Markdown, HTML, PDF, and plain text — via URL, S3, or file upload.',
  },
  {
    icon: '✂️',
    title: 'Chunking strategies',
    description: 'Format-aware chunking out of the box. Configurable chunk size and overlap.',
  },
];

const useCases = [
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

export default function Home(): ReactNode {
  return (
    <Layout description="Self-hostable semantic search and RAG infrastructure for developers.">
      <main>
        <div className={styles.hero}>
          <div className={styles.heroInner}>
            <div className={styles.heroLeft}>
              <div className={styles.badge}>Open Source · MIT</div>
              <h1 className={styles.heroTitle}>
                RAG and semantic search,<br />
                <span className={styles.heroTitleAccent}>self-hosted in minutes</span>
              </h1>
              <p className={styles.heroSubtitle}>
                Add semantic search and RAG to your existing app in minutes.
                Self-hosted on your own infra — bring your own AI with Ollama or OpenAI.
              </p>
              <div className={styles.heroActions}>
                <Link className={styles.btnPrimary} to="/docs/getting-started">
                  Get started →
                </Link>
                <GitHubStars />
              </div>
            </div>
            <div className={styles.heroRight}>
              <div className={styles.terminal}>
                <div className={styles.terminalHeader}>
                  <span className={styles.dot} style={{background: '#ff5f57'}} />
                  <span className={styles.dot} style={{background: '#febc2e'}} />
                  <span className={styles.dot} style={{background: '#28c840'}} />
                </div>
                <pre className={styles.terminalBody}>{`$ npx ragpack init && npx ragpack start
✔ API on :9000 · Admin UI on :3000

# Drop into any existing app
import { RagPack } from "ragpack-js";

const client = new RagPack({
  baseUrl: "http://localhost:9000",
});

const col = client.collection("my-docs");

await col.ingest({
  uri: "https://your-docs-site.com/guide",
});

const { answer } = await col.rag({
  query: "how does auth work?",
});`}
                </pre>
              </div>
            </div>
          </div>
        </div>

        <div className={styles.features}>
          <div className={styles.featuresInner}>
            {features.map(({eyebrow, title, description}) => (
              <div key={title} className={styles.feature}>
                <div className={styles.featureEyebrow}>{eyebrow}</div>
                <h3>{title}</h3>
                <p>{description}</p>
              </div>
            ))}
          </div>
        </div>

        <div className={styles.included}>
          <div className={styles.includedInner}>
            <h2 className={styles.includedTitle}>What's included</h2>
            <div className={styles.includedGrid}>
              {included.map(({icon, title, description}) => (
                <div key={title} className={styles.includedCard}>
                  <div className={styles.includedIcon}>{icon}</div>
                  <h4>{title}</h4>
                  <p>{description}</p>
                </div>
              ))}
            </div>
          </div>
        </div>

        <div className={styles.useCases}>
          <div className={styles.useCasesInner}>
            <h2 className={styles.includedTitle}>What can you build?</h2>
            <div className={styles.useCasesGrid}>
              {useCases.map(({title, description, example}) => (
                <div key={title} className={styles.useCaseCard}>
                  <h4>{title}</h4>
                  <p>{description}</p>
                  <div className={styles.useCaseExample}>{example}</div>
                </div>
              ))}
            </div>
          </div>
        </div>

        <div className={styles.cta}>
          <div className={styles.ctaInner}>
            <h2>Ready to ship your RAG pipeline?</h2>
            <p>One command. No cloud account. No token limits.</p>
            <Link className={styles.btnPrimary} to="/docs/getting-started">
              Read the docs →
            </Link>
          </div>
        </div>
      </main>
    </Layout>
  );
}
