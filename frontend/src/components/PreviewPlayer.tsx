interface PreviewPlayerProps {
  videoUrl?: string;
  ready: boolean;
}

export function PreviewPlayer({ videoUrl, ready }: PreviewPlayerProps) {
  return (
    <section className="panel preview-panel">
      <p className="eyebrow">Preview stage</p>
      <h2>Playback surface</h2>
      {ready && videoUrl ? (
        <video controls preload="metadata" className="preview-player">
          <source src={videoUrl} type="video/mp4" />
        </video>
      ) : (
        <div className="preview-empty-state">
          <p>Final preview playback will appear here once rendering completes.</p>
          <span>The cinematic surface stays ready while the render pipeline finishes.</span>
        </div>
      )}
    </section>
  );
}
