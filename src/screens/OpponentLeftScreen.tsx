type OpponentLeftScreenProps = {
  onLeave: () => void;
};

export function OpponentLeftScreen({ onLeave }: OpponentLeftScreenProps) {
  return (
    <main className="centered-screen">
      <section className="status-ticket opponent-left-ticket">
        <span className="status-label">Match ended</span>
        <div className="opponent-left-mark" aria-hidden="true">×</div>
        <h1>Opponent left</h1>
        <p>The other chef stepped away, so this match can’t continue.</p>
        <button className="primary-button" onClick={onLeave}>Back to kitchen</button>
      </section>
    </main>
  );
}
