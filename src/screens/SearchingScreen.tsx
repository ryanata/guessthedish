import { ChefAvatar } from "../components/ChefAvatar";
import type { MatchSnapshot } from "../types";

type SearchingScreenProps = {
  match?: MatchSnapshot;
  onCancel: () => void;
};

export function SearchingScreen({ match, onCancel }: SearchingScreenProps) {
  return (
    <main className="centered-screen">
      <section className="status-ticket">
        <span className="status-label">Quick play</span>
        {match && match.phase !== "waiting" ? (
          <>
            <h1>Match found</h1>
            <div className="matchup-players">
              <div className="matchup-player matchup-player-you">
                <span className="matchup-you-label">You</span>
                <ChefAvatar avatar={match.player.avatar} />
                <strong>{match.player.name}</strong>
              </div>
              <span>vs</span>
              <div className="matchup-player">
                <ChefAvatar avatar={match.opponent.avatar} />
                <strong>{match.opponent.name}</strong>
              </div>
            </div>
            <p>Aprons on. First clue coming up.</p>
          </>
        ) : (
          <>
            <div className="searching-mark" aria-hidden="true">
              <span />
              <span />
              <span />
            </div>
            <h1>Setting the table</h1>
            <p>Your local opponent is warming up.</p>
          </>
        )}
        <button className="text-button" onClick={onCancel}>Cancel search</button>
      </section>
    </main>
  );
}
