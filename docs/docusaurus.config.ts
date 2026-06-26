import {themes as prismThemes} from 'prism-react-renderer';
import type {Config} from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';

const config: Config = {
  title: 'RagPack',
  tagline: 'Self-hostable semantic search and RAG — up in minutes.',
  favicon: 'img/favicon.ico',

  future: {
    v4: true,
  },

  url: 'https://ragpack.dev',
  baseUrl: '/',

  organizationName: 'eozsahin1993',
  projectName: 'ragpack',

  onBrokenLinks: 'throw',

  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  presets: [
    [
      'classic',
      {
        docs: {
          sidebarPath: './sidebars.ts',
          editUrl: 'https://github.com/eozsahin1993/ragpack/tree/main/docs/',
        },
        blog: false,
        theme: {
          customCss: './src/css/custom.css',
        },
      } satisfies Preset.Options,
    ],
  ],

  themeConfig: {
    colorMode: {
      defaultMode: 'dark',
      respectPrefersColorScheme: true,
    },
    navbar: {
      title: 'RagPack',
      items: [
        {
          type: 'docSidebar',
          sidebarId: 'docs',
          position: 'left',
          label: 'Docs',
        },
        {
          href: 'https://github.com/eozsahin1993/ragpack',
          label: 'GitHub',
          position: 'right',
        },
      ],
    },
    footer: {
      style: 'dark',
      links: [
        {
          title: 'Docs',
          items: [
            {label: 'Getting Started', to: '/docs/getting-started'},
            {label: 'API Reference', to: '/docs/api'},
            {label: 'JS SDK', to: '/docs/sdk/js'},
          ],
        },
        {
          title: 'Packages',
          items: [
            {label: 'ragpack (CLI)', href: 'https://www.npmjs.com/package/ragpack'},
            {label: 'ragpack-js (SDK)', href: 'https://www.npmjs.com/package/ragpack-js'},
          ],
        },
        {
          title: 'Source',
          items: [
            {label: 'GitHub', href: 'https://github.com/eozsahin1993/ragpack'},
            {label: 'Issues', href: 'https://github.com/eozsahin1993/ragpack/issues'},
          ],
        },
      ],
      copyright: `MIT License © ${new Date().getFullYear()} RagPack`,
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.dracula,
      additionalLanguages: ['bash', 'json', 'yaml'],
    },
  } satisfies Preset.ThemeConfig,
};

export default config;
