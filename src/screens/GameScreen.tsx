import { AnswerSearch } from "../components/AnswerSearch";
import { PlayerCard } from "../components/PlayerCard";
import { ResultOverlay } from "../components/ResultOverlay";
import type { Dish, MatchSnapshot } from "../types";

type GameScreenProps = {
  dishes: Dish[];
  match: MatchSnapshot;
  error?: string;
  onGuess: (dish: Dish) => void;
  onLeave: () => void;
  onPlayAgain: () => void;
};

function resultText(match: MatchSnapshot) {
  if (match.phase === "finished") {
    return match.matchWinner === "player" ? "You won the match" : `${match.opponent.name} won the match`;
  }
  if (match.roundWinner === "player") return "You got it";
  if (match.roundWinner === "opponent") return `${match.opponent.name} got it`;
  return "Time's up";
}

export function GameScreen({ dishes, match, error, onGuess, onLeave, onPlayAgain }: GameScreenProps) {
  const locked = Boolean(match.lockUntil);
  const playing = match.phase === "playing";
  const earlyClueCount = Math.ceil(match.totalClueCount * 0.25);
  const revealDuration = match.clues.length <= earlyClueCount ? "5s" : "7.5s";

  return (
    <main className="game-screen">
      <header className="game-header">
        <div className="compact-wordmark">Guess the Dish</div>
        <div className="round-label">Round {match.round} <span>First to 3</span></div>
        <button className="leave-button" onClick={onLeave}>Leave</button>
      </header>

      <section className="players">
        <PlayerCard player={{ name: match.player.name, avatar: match.player.avatar, score: match.player.score, guess: match.player.latestGuess, isYou: true }} />
        <div className="versus">vs</div>
        <PlayerCard player={{ name: match.opponent.name, avatar: match.opponent.avatar, score: match.opponent.score, guess: match.opponent.latestGuess }} />
      </section>

      <section className="clue-ticket">
        <div className="clue-heading">
          <div>
            <span>{playing ? "Order up" : resultText(match)}</span>
            <strong>{playing ? "What are we making?" : match.answer?.name}</strong>
          </div>
          {playing && (
            <div className="reveal-progress">
              <span>{match.nextRevealAt ? "Next clue" : "Final guesses"}</span>
              <div className="reveal-progress-track" aria-hidden="true">
                <i key={match.nextRevealAt ?? match.deadlineAt} style={{ animationDuration: revealDuration }} />
              </div>
            </div>
          )}
        </div>
        {match.totalClueCount > 0 && (
          <div className="assists" role="status" aria-live="polite">
            <Assist label="Family" value={match.family} />
            <Assist label="Cuisine" value={match.cuisine} />
          </div>
        )}
        <ol className="ingredients">
          {Array.from({ length: match.totalClueCount }, (_, index) => (
            <li key={index} className={index < match.clues.length ? "ingredient-visible" : ""}>
              <span>{String(index + 1).padStart(2, "0")}</span>
              <strong>{match.clues[index] ?? "Waiting at the pass"}</strong>
              {playing && index === match.clues.length - 1 && <small>New</small>}
            </li>
          ))}
        </ol>
        {match.phase === "result" && match.answer && (
          <div className={`round-result-card round-result-${match.roundWinner}`} role="status">
            <span>{resultText(match)}</span>
            <strong>{match.answer.name}</strong>
            <small>{match.roundWinner === "none" ? "No one got it" : "Point awarded"}</small>
          </div>
        )}
      </section>

      <footer className="answer-dock">
        {match.phase !== "finished" && (
          <AnswerSearch key={match.round} dishes={dishes} disabled={!playing} locked={locked} onSubmit={onGuess} />
        )}
        <div className="answer-hint" role="status">
          {error || (locked
            ? "Not on the ticket. Try again."
            : playing
              ? "Select a dish to send your guess instantly."
              : "Next round coming up...")}
        </div>
      </footer>

      {match.phase === "finished" && (
        <ResultOverlay
          won={match.matchWinner === "player"}
          playerScore={match.player.score}
          opponentScore={match.opponent.score}
          opponentName={match.opponent.name}
          onPlayAgain={onPlayAgain}
          onLeave={onLeave}
        />
      )}
    </main>
  );
}

function Assist({ label, value }: { label: string; value?: string }) {
  const unlocked = Boolean(value);
  return (
    <span className={unlocked ? "assist assist-unlocked" : "assist"}>
      <small>{label}</small>
      <strong aria-label={unlocked ? undefined : `${label} not revealed yet`}>{unlocked ? value : "?"}</strong>
    </span>
  );
}
