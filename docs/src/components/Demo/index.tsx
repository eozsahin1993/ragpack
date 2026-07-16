import { useEffect, useRef } from 'react';
import styles from './styles.module.css';

// Drop a recording in here once it exists. An mp4/webm URL renders as <video>,
// a youtube.com/embed/... or loom.com/embed/... URL renders as an <iframe>.
const DEMO_VIDEO_SRC = '/video/demo.mp4';

function AutoplayVideo({ src }: { src: string }) {
  const videoRef = useRef<HTMLVideoElement>(null);

  useEffect(() => {
    const video = videoRef.current;
    if (!video) return;

    // Only autoplay once the visitor has actually scrolled - otherwise a page
    // load that happens to land with the video already in view (it's the
    // section right under the hero) would autoplay before any scrolling at all.
    let hasScrolled = false;
    let isIntersecting = false;

    const maybePlay = () => {
      if (hasScrolled && isIntersecting) {
        video.play().catch(() => {}); // autoplay can still be blocked, e.g. low-power mode - controls remain as a fallback
      }
    };

    const onScroll = () => {
      if (window.scrollY > 10) {
        hasScrolled = true;
        maybePlay();
        window.removeEventListener('scroll', onScroll);
      }
    };
    window.addEventListener('scroll', onScroll, { passive: true });

    const observer = new IntersectionObserver(
      ([entry]) => {
        isIntersecting = entry.isIntersecting;
        if (isIntersecting) {
          maybePlay();
        } else {
          video.pause();
        }
      },
      { threshold: 0.5 }
    );
    observer.observe(video);

    return () => {
      observer.disconnect();
      window.removeEventListener('scroll', onScroll);
    };
  }, []);

  return (
    <video
      ref={videoRef}
      className={styles.frame}
      src={src}
      controls
      muted
      playsInline
      preload="metadata"
    />
  );
}

function DemoEmbed({ src }: { src: string }) {
  if (src.includes('youtube.com/embed') || src.includes('loom.com/embed')) {
    return (
      <iframe
        className={styles.frame}
        src={src}
        title="RagPack demo"
        allow="autoplay; fullscreen; picture-in-picture"
        allowFullScreen
      />
    );
  }
  return <AutoplayVideo src={src} />;
}

function DemoPlaceholder() {
  return (
    <div className={styles.placeholder}>
      <div className={styles.playButton}>
        <svg width="22" height="22" viewBox="0 0 24 24" fill="currentColor">
          <path d="M8 5v14l11-7z" />
        </svg>
      </div>
      <p>Demo video coming soon</p>
    </div>
  );
}

export default function Demo() {
  return (
    <div className={styles.wrapper}>
      <div className={styles.inner}>
        <h2>See it in action</h2>
        <p className={styles.subtitle}>
          Ingest a document, query it, and get a grounded answer in under two minutes.
        </p>
        <div className={styles.embedWrap}>
          {DEMO_VIDEO_SRC ? <DemoEmbed src={DEMO_VIDEO_SRC} /> : <DemoPlaceholder />}
        </div>
      </div>
    </div>
  );
}
