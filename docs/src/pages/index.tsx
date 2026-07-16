import type {ReactNode} from 'react';
import Layout from '@theme/Layout';
import Hero from '@site/src/components/Hero';
import Demo from '@site/src/components/Demo';
import Comparison from '@site/src/components/Comparison';
import WhatIsIncluded from '@site/src/components/WhatIsIncluded';
import Chunking from '@site/src/components/Chunking';
import Evals from '@site/src/components/Evals';
import Formats from '@site/src/components/Formats';
import QuickStart from '@site/src/components/QuickStart';
import UseCases from '@site/src/components/UseCases';
import FAQ from '@site/src/components/FAQ';
import CTA from '@site/src/components/CTA';

export default function Home(): ReactNode {
  return (
    <Layout description="Self-hosted, open-source RAG and semantic search API with hybrid search (BM25 + vector) and an embedded vector database. Bring your own embedding model: Ollama, TEI, or OpenAI.">
      <main>
        <Hero />
        <Demo />
        <Comparison />
        <WhatIsIncluded />
        <Chunking />
        <Evals />
        <Formats />
        <QuickStart />
        <UseCases />
        <FAQ />
        <CTA />
      </main>
    </Layout>
  );
}
