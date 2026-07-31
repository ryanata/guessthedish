type ResultOverlayProps = {
  won: boolean;
  playerScore?: number;
  opponentScore?: number;
  opponentName?: string;
  onPlayAgain?: () => void;
  onLeave?: () => void;
};

export function ResultOverlay({
  won,
  playerScore = won ? 3 : 1,
  opponentScore = won ? 1 : 3,
  opponentName = "MiseEnPlace",
  onPlayAgain,
  onLeave,
}: ResultOverlayProps) {
  const verdict = won ? "You win" : "You lose";
  const note = won ? "Service complete" : `${opponentName} takes the table`;

  return (
    <section className={`result-overlay ${won ? "result-win" : "result-loss"}`}>
      <div className="result-scrim" />
      <div className="result-content">
        <div className="result-verdict">
          <span className="result-kicker">Final ticket / First to 3</span>
          <h1>{verdict}</h1>
          <p>{note}</p>
        </div>
        <div className="result-score" aria-label={`Final score ${playerScore} to ${opponentScore}`}>
          <span><small>You</small><strong>{playerScore}</strong></span>
          <i>:</i>
          <span><small>{opponentName}</small><strong>{opponentScore}</strong></span>
        </div>
        <div className="result-actions">
          <button onClick={onPlayAgain}>Play again</button>
          <button onClick={onLeave}>Back to kitchen</button>
        </div>
      </div>
    </section>
  );
}
