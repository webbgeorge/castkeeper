import clsx from 'clsx';
import Heading from '@theme/Heading';
import styles from './styles.module.css';

const FeatureList = [
  {
    title: 'Flexible Storage',
    icon: '💾',
    description: (
      <>
        Store your podcast backups wherever you prefer — locally, or in your
        cloud storage provider of choice. Keep your collection safe, accessible,
        and always under your control.
      </>
    ),
  },
  {
    title: 'Automatic Downloads',
    icon: '📥',
    description: (
      <>
        Subscribe to your favorite podcasts, and let CastKeeper automatically
        download new episodes as they&apos;re released. Never miss an episode
        — your backups stay current without any effort.
      </>
    ),
  },
  {
    title: 'Simple Self-Hosting',
    icon: '🏠',
    description: (
      <>
        Easily deploy with Docker — perfect for self-hosters who want full
        control. Clear setup instructions make getting started straightforward.
      </>
    ),
  },
];

function Feature({ Svg, title, description, icon }) {
  return (
    <div className={clsx('col col--4')}>
      <div className="text--center">
        <span style={{ fontSize: "72px" }}>{icon}</span>
      </div>
      <div className="text--center padding-horiz--md">
        <Heading as="h3">{title}</Heading>
        <p>{description}</p>
      </div>
    </div>
  );
}

export default function HomepageFeatures() {
  return (
    <section className={styles.features}>
      <div className="container">
        <div className="row">
          {FeatureList.map((props, idx) => (
            <Feature key={idx} {...props} />
          ))}
        </div>
      </div>
    </section>
  );
}
