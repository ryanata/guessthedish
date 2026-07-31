import type { Player } from "../types";
import { ChefAvatar } from "./ChefAvatar";

type PlayerCardProps = {
  player: Player;
};

export function PlayerCard({ player }: PlayerCardProps) {
  return (
    <section className={`player-card ${player.isYou ? "player-card-you" : ""}`}>
      <div className="player-meta">
        <div className="player-avatar">
          {player.guess && <div className="guess-bubble">{player.guess}</div>}
          <ChefAvatar avatar={player.avatar} />
        </div>
        <span className="connection-dot" aria-label="Connected" />
        <strong>{player.name}</strong>
        {player.isYou && <span className="you-label">You</span>}
      </div>
      <div className="score" aria-label={`${player.score} of 3 rounds won`}>
        {[0, 1, 2].map((point) => (
          <span key={point} className={point < player.score ? "score-won" : ""} />
        ))}
      </div>
    </section>
  );
}
