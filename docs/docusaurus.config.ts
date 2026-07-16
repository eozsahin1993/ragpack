import {themes as prismThemes} from 'prism-react-renderer';
import type {Config} from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';

const config: Config = {
  title: 'RagPack: Lightweight Self-Hosted RAG Infrastructure',
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

  headTags: [
    {
      tagName: 'link',
      attributes: { rel: 'preconnect', href: 'https://fonts.googleapis.com' },
    },
    {
      tagName: 'link',
      attributes: { rel: 'preconnect', href: 'https://fonts.gstatic.com', crossorigin: 'anonymous' },
    },
    {
      tagName: 'link',
      attributes: {
        rel: 'stylesheet',
        href: 'https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700;800&family=JetBrains+Mono:wght@400;500;600;700&display=swap',
      },
    },
  ],

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
      { name: 'keywords', content: 'self-hosted RAG, open source RAG, semantic search API, hybrid search, vector database, bring your own embedding model, Ollama, retrieval augmented generation' },
      { name: 'twitter:card', content: 'summary_large_image' },
      { name: 'twitter:site', content: '@ragpackdev' },
      { property: 'og:type', content: 'website' },
      { property: 'og:site_name', content: 'RagPack' },
    ],
    colorMode: {
      defaultMode: 'light',
      respectPrefersColorScheme: false,
      disableSwitch: true,
    },
    navbar: {
      title: 'RagPack',
      logo: {
        alt: 'RagPack',
        src: 'img/logo.svg',
      },
      items: [
        { href: '/#why', label: 'Why', position: 'right' },
        { href: '/#features', label: 'Features', position: 'right' },
        { href: '/#evals', label: 'Evals', position: 'right' },
        { href: '/#formats', label: 'Formats', position: 'right' },
        {
          type: 'docSidebar',
          sidebarId: 'docs',
          position: 'right',
          label: 'Docs',
        },
        {
          type: 'custom-githubStars',
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
