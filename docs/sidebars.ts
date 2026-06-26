import type {SidebarsConfig} from '@docusaurus/plugin-content-docs';

const sidebars: SidebarsConfig = {
  docs: [
    'getting-started',
    'api',
    {
      type: 'category',
      label: 'SDKs',
      items: ['sdk/js'],
    },
  ],
};

export default sidebars;
