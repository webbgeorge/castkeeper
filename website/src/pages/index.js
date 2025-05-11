import clsx from 'clsx';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';
import HomepageFeatures from '@site/src/components/HomepageFeatures';

import Heading from '@theme/Heading';
import styles from './index.module.css';

function HomepageHeader() {
  const { siteConfig } = useDocusaurusContext();
  return (
    <header className={clsx('hero hero--primary', styles.heroBanner)}>
      <div className="container">
        <Heading as="h1" className="hero__title">
          {siteConfig.title}
        </Heading>
        <p className="hero__subtitle">{siteConfig.tagline}</p>
        <div className={styles.buttons}>
          <Link
            className="button button--secondary button--lg"
            to="/docs/intro">
            Get Started
          </Link>
        </div>
      </div>
    </header>
  );
}

export default function Home() {
  return (
    <Layout
      title="Backup the podcasts you love"
      description="Keep copies of the podcasts you listen to, safe from future deletion, cancellation and censorship.">
      <HomepageHeader />
      <main>
        <div
          className="text--center padding-horiz--md padding-vert--lg"
          style={{ "max-width": "800px", "margin": "0 auto" }}
        >
          <h1>What is CastKeeper?</h1>
          <p>
            CastKeeper is a free application for archiving podcasts. It is designed to be easy to self-host, using either Docker or a self-contained executable. It supports a variety of file storage options.
          </p>
          <p>
            <strong>CastKeeper was built primarily to do 2 things:</strong>
          </p>
          <p>
            1. Store copies of podcasts, safe from being taken offline or being modified (e.g. due to being abandoned or censored), allowing re-listening in the future.
          </p>
          <p>
            2. Transcribe and index the text content of those podcasts, allowing searching for when certain topics were discussed (coming soon in a future release).
          </p>
        </div>
        <HomepageFeatures />
      </main>
    </Layout >
  );
}
