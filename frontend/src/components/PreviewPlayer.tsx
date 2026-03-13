interface PreviewPlayerProps {
  videoUrl?: string;
  ready: boolean;
}

export function PreviewPlayer({ videoUrl, ready }: PreviewPlayerProps) {
  return (
    <section className="panel">
      <p className="eyebrow">Phase 4</p>
      <h2>Preview Output</h2>
      {ready && videoUrl ? (
        <video controls preload="metadata" className="preview-player">
          <source src={videoUrl} type="video/mp4" />
        </video>
      ) : (
        <p>Final preview playback will appear here once rendering completes.</p>
      )}
    </section>
  );
}
