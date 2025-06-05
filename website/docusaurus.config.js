// @ts-check
// `@type` JSDoc annotations allow editor autocompletion and type checking
// (when paired with `@ts-check`).
// There are various equivalent ways to declare your Docusaurus config.
// See: https://docusaurus.io/docs/api/docusaurus-config

import { themes as prismThemes } from 'prism-react-renderer';

/** @type {import('@docusaurus/types').Config} */
const config = {
  title: 'CastKeeper',
  tagline: 'A free application for archiving podcasts',
  favicon: 'data:image/svg+xml,<svg xmlns=%22http://www.w3.org/2000/svg%22 width=%2264%22 height=%2264%22 viewBox=%220 0 64 64%22>  <rect width=%2264%22 height=%2264%22 rx=%2212%22 fill=%22%23242933%22/><text x=%2250%25%22 y=%2254%25%22 text-anchor=%22middle%22 dominant-baseline=%22middle%22 font-family=%22Segoe UI, Arial, sans-serif%22 font-size=%2232%22 fill=%22%23fff%22 font-weight=%22bold%22 letter-spacing=%222%22>CK</text></svg>',
  url: 'https://castkeeper.org',
  // Set the /<baseUrl>/ pathname under which your site is served
  // For GitHub pages deployment, it is often '/<projectName>/'
  baseUrl: '/',

  organizationName: 'webbgeorge',
  projectName: 'castkeeper',

  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'warn',

  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  presets: [
    [
      'classic',
      /** @type {import('@docusaurus/preset-classic').Options} */
      ({
        docs: {
          routeBasePath: '/',
          sidebarPath: './sidebars.js',
          editUrl:
            'https://github.com/webbgeorge/castkeeper/tree/main/website/',
        },
        theme: {
          customCss: './src/css/custom.css',
        },
      }),
    ],
  ],

  themeConfig:
    /** @type {import('@docusaurus/preset-classic').ThemeConfig} */
    ({
      navbar: {
        title: 'CastKeeper',
        items: [
          {
            type: 'docSidebar',
            sidebarId: 'docsSidebar',
            position: 'left',
            label: 'Docs',
          },
          {
            href: 'https://github.com/webbgeorge/castkeeper',
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
              {
                label: 'Getting started',
                to: '/',
              },
              {
                label: 'Installation',
                to: '/getting-started/installation',
              },
              {
                label: 'Configuration',
                to: '/getting-started/configuration',
              },
              {
                label: 'Using CastKeeper',
                to: '/category/using-castkeeper',
              },
            ],
          },
          {
            title: 'Community',
            items: [
              {
                label: 'Raise a feature request',
                href: 'https://github.com/webbgeorge/castkeeper/issues/new?template=feature_request.md',
              },
              {
                label: 'Report a bug',
                href: 'https://github.com/webbgeorge/castkeeper/issues/new?template=bug_report.md',
              },
              {
                label: 'Ask a question',
                href: 'https://github.com/webbgeorge/castkeeper/issues/new?template=ask_question.md',
              },
            ],
          },
          {
            title: 'More',
            items: [
              {
                label: 'GitHub',
                href: 'https://github.com/webbgeorge/castkeeper',
              },
            ],
          },
        ],
        copyright: `Copyright © ${new Date().getFullYear()} George Webb. Built with Docusaurus.`,
      },
      prism: {
        theme: prismThemes.github,
        darkTheme: prismThemes.dracula,
      },
    }),
};

export default config;
