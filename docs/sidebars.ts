import type {SidebarsConfig} from '@docusaurus/plugin-content-docs';

const sidebars: SidebarsConfig = {
  docs: [
    'getting-started',
    'configuration',
    'api',
    {
      type: 'category',
      label: 'SDKs',
      items: [
        {
          type: 'category',
          label: 'JavaScript / TypeScript',
          items: [
            'sdk/js/setup',
            'sdk/js/collections',
            'sdk/js/ingesting',
            'sdk/js/query',
            'sdk/js/prompts',
            'sdk/js/jobs',
            'sdk/js/error-handling',
          ],
        },
      ],
    },
  ],
};

export default sidebars;
