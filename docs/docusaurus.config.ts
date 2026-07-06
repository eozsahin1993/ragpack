import {themes as prismThemes} from 'prism-react-renderer';
import type {Config} from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';

const config: Config = {
  title: 'RagPack — Lightweight Self-Hosted RAG Infrastructure',
  tagline: 'Open-source RAG infrastructure for developers. Self-hosted, works with Ollama, no cloud bill. Add semantic search to your app in minutes.',
  favicon: 'img/favicon.svg',

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
        gtag: {
          trackingID: 'G-5JFM2MK4EE',
          anonymizeIP: true,
        },
      } satisfies Preset.Options,
    ],
  ],

  themeConfig: {
    image: 'img/social-card.png',
    metadata: [
      { name: 'twitter:card', content: 'summary_large_image' },
      { name: 'twitter:site', content: '@ragpackdev' },
      { property: 'og:type', content: 'website' },
      { property: 'og:site_name', content: 'RagPack' },
    ],
    colorMode: {
      defaultMode: 'dark',
      respectPrefersColorScheme: true,
    },
    navbar: {
      title: 'RagPack',
      logo: {
        alt: 'RagPack',
        src: 'img/logo.svg',
      },
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
      links: [],
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.dracula,
      additionalLanguages: ['bash', 'json', 'yaml'],
    },
  } satisfies Preset.ThemeConfig,
};

export default config;
