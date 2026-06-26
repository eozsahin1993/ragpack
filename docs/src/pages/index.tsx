import type {ReactNode} from 'react';
import Layout from '@theme/Layout';
import Hero from '@site/src/components/Hero';
import ValueProps from '@site/src/components/ValueProps';
import WhatIsIncluded from '@site/src/components/WhatIsIncluded';
import UseCases from '@site/src/components/UseCases';
import CTA from '@site/src/components/CTA';

export default function Home(): ReactNode {
  return (
    <Layout description="Self-hostable semantic search and RAG infrastructure for developers.">
      <main>
        <Hero />
        <ValueProps />
        <WhatIsIncluded />
        <UseCases />
        <CTA />
      </main>
    </Layout>
  );
}
